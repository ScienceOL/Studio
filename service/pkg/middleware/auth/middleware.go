package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/casdoor"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
	"golang.org/x/oauth2"
)

type userAuth struct {
	casDoorClient repo.Account
}

// 用于认证的错误
var (
	ErrInvalidToken = errors.New("invalid or expired token")

	authClient *userAuth
	once       sync.Once
)

// ValidateToken 检查令牌是否有效
func ValidateToken(ctx context.Context, tokenType string, token string) (*model.UserData, error) {
	// 获取OAuth2配置
	oauthConfig := GetOAuthConfig()
	// 创建一个包含传入token的oauth2.Token对象
	oauthToken := &oauth2.Token{
		AccessToken: token,
		TokenType:   tokenType,
	}

	// 使用token构建OAuth2客户端
	client := oauthConfig.Client(ctx, oauthToken)

	// 获取配置中的用户信息URL
	config := webapp.Config()

	// 获取用户信息 - 如果token有效，这个请求将成功
	resp, err := client.Get(config.OAuth2.UserInfoURL)
	if err != nil {
		logger.Errorf(ctx, "Failed to get user info: %v", err)
		return nil, ErrInvalidToken
	}
	logger.Infof(ctx, "Response status: %d", resp.StatusCode)
	defer resp.Body.Close()

	// 如果状态码不是2xx，则认为token无效
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Errorf(ctx, "Invalid token, status code: %d", resp.StatusCode)
		return nil, ErrInvalidToken
	}

	// 解析用户信息
	result := &model.UserInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil ||
		result.Status != "ok" ||
		result.Data == nil {
		logger.Errorf(ctx, "Failed to parse user info: %v", err)
		return nil, err
	}

	// 检查API调用是否成功
	return result.Data, nil
}

func Auth() func(ctx *gin.Context) {
	once.Do(func() {
		authClient = &userAuth{
			casDoorClient: casdoor.NewCasClient(),
		}
	})
	return authClient.AuthUser
}

// RequireAuth 中间件函数验证用户是否已登录
func (u *userAuth) AuthUser(ctx *gin.Context) {
	// 从请求头获取Authorization
	cookie, _ := ctx.Cookie("access_token_v2")
	authHeader := ctx.GetHeader("Authorization")
	queryToken := ctx.Query("access_token_v2")
	authHeader = utils.Or(cookie, queryToken, authHeader)
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, &common.Resp{
			Code: code.UnLogin,
			Error: &common.Error{
				Msg: code.UnLogin.String(),
			},
		})
		ctx.Abort()
		return
	}

	tokens := strings.Split(authHeader, " ")
	if len(tokens) != 2 {
		ctx.JSON(http.StatusUnauthorized,
			&common.Resp{
				Code: code.LoginFormatErr,
				Error: &common.Error{
					Msg: code.LoginFormatErr.String(),
				},
			})
		ctx.Abort()
		return
	}

	var userInfo *model.UserData
	authKey := USERKEY
	switch tokens[0] {
	case "Bearer":
		userInfo = u.getNormalUser(ctx, authHeader)
	case "lab":
		userInfo = u.getLabUser(ctx, authHeader)
		authKey = LABKEY
	case "api":
		userInfo = u.getLabUser(ctx, authHeader)
	}

	if userInfo == nil {
		ctx.JSON(http.StatusUnauthorized,
			&common.Resp{
				Code: code.LoginFormatErr,
				Error: &common.Error{
					Msg: code.LoginFormatErr.String(),
				},
			})
		ctx.Abort()
		return
	}

	// 将用户信息保存到上下文
	ctx.Set(authKey, userInfo)
	ctx.Next()
}

func (u *userAuth) getLabUser(ctx *gin.Context, authHeader string) *model.UserData {
	// 检查格式是否为 "Bearer {token}"
	baseSplit := strings.Split(authHeader, " ")
	if len(baseSplit) != 2 {
		logger.Errorf(ctx, "getLabUser authHeader format invalid")
		return nil
	}

	baseStr, err := base64.StdEncoding.DecodeString(baseSplit[1])
	if err != nil {
		logger.Errorf(ctx, "getLabUser decode auth header err: %s", err.Error())
		return nil
	}

	keys := strings.Split(string(baseStr), ":")
	if len(keys) != 2 {
		logger.Errorf(ctx, "getLabUser base formate err not 2")
		return nil
	}
	userInfo, err := u.casDoorClient.GetLabUserInfo(ctx, &model.LabAkSk{
		AccessKey:    keys[0],
		AccessSecret: keys[1],
	})
	if err != nil {
		logger.Errorf(ctx, "getLabUser GetLabUserInfo err: %s", err.Error())
		return nil
	}

	return userInfo
}

func (u *userAuth) getNormalUser(ctx *gin.Context, authHeader string) *model.UserData {
	// 检查格式是否为 "Bearer {token}"
	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		logger.Errorf(ctx, "bearer format err: %s", authHeader)
		return nil
	}

	// 验证令牌
	userInfo, err := ValidateToken(ctx, bearerToken[0], bearerToken[1])
	if err != nil {
		logger.Errorf(ctx, "Token validation failed: %v", err)
		return nil
	}
	return userInfo
}

// GetCurrentUser 从上下文中获取当前用户信息
func GetCurrentUser(ctx context.Context) *model.UserData {
	gCtx, ok := ctx.(*gin.Context)
	if !ok {
		return nil
	}

	user, exists := gCtx.Get(USERKEY)
	if !exists {
		return nil
	}
	return user.(*model.UserData)
}

func GetLabUser(ctx context.Context) *model.UserData {
	gCtx, ok := ctx.(*gin.Context)
	if !ok {
		return nil
	}

	user, exists := gCtx.Get(LABKEY)
	if !exists {
		return nil
	}
	return user.(*model.UserData)
}
