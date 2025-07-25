package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"golang.org/x/oauth2"
)

// 用于认证的错误
var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// ValidateToken 检查令牌是否有效
func ValidateToken(ctx context.Context, tokenType string, token string) (*UserData, error) {
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
	result := &UserInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil ||
		result.Status != "ok" ||
		result.Data == nil {
		logger.Errorf(ctx, "Failed to parse user info: %v", err)
		return nil, err
	}

	// 检查API调用是否成功
	return result.Data, nil

	// 如果没有status字段或者status不是ok，直接将result作为用户数据
	// 验证返回的数据包含必要的用户信息
	// if result["id"] == nil {
	// 	logger.Errorf(ctx, "Invalid user data format: missing required fields")
	// 	return nil, errors.New("invalid user data format")
	// }

	// // 返回验证成功和用户数据
	// return map[string]any{
	// 	"valid": true,
	// 	"user":  result, // 直接使用返回的用户对象
	// }, nil
}

// RequireAuth 中间件函数验证用户是否已登录
func Auth(ctx *gin.Context) {
	// 从请求头获取Authorization
	authHeader := ctx.GetHeader("Authorization")
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

	// 检查格式是否为 "Bearer {token}"
	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
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

	// 验证令牌
	userInfo, err := ValidateToken(ctx, bearerToken[0], bearerToken[1])
	if err != nil {
		logger.Errorf(ctx, "Token validation failed: %v", err)
		ctx.JSON(http.StatusUnauthorized, &common.Resp{
			Code: code.InvalidToken,
			Error: &common.Error{
				Msg: code.InvalidToken.String(),
			},
		})
		ctx.Abort()
		return
	}

	// 将用户信息保存到上下文
	ctx.Set(USERKEY, userInfo)
	ctx.Next()
}

// GetCurrentUser 从上下文中获取当前用户信息
func GetCurrentUser(ctx context.Context) *UserData {
	gCtx, ok := ctx.(*gin.Context)
	if !ok {
		return nil
	}

	user, exists := gCtx.Get(USERKEY)
	if !exists {
		return nil
	}
	return user.(*UserData)
}
