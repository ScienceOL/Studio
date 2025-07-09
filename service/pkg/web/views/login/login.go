package login

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/web/types"
	"golang.org/x/oauth2"
)

type login struct {
	oauthConfig *oauth2.Config
	userInfoURL string
}

func NewLogin() *login {
	conf := webapp.Config().OAuth2
	return &login{
		oauthConfig: &oauth2.Config{
			ClientID:     conf.ClientID,
			ClientSecret: conf.ClientSecret,
			Scopes:       conf.Scopes,
			Endpoint: oauth2.Endpoint{
				TokenURL: conf.TokenURL,
				AuthURL:  conf.AuthURL,
			},
			RedirectURL: conf.RedirectURL,
		},
		userInfoURL: conf.UserInfoURL,
	}
}

func (l *login) Login(ctx *gin.Context) {
	// 生成随机state用于防止CSRF攻击
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	// 将state保存到Redis中，设置5分钟过期时间
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := redis.GetClient().Set(ctx, stateKey, "valid", 5*time.Minute).Err(); err != nil {
		logger.Errorf(ctx, "Failed to save state to Redis: %v", err)
		common.ReplyErr(ctx, code.LoginSetStateErr)
		return
	}

	// 构建授权URL并重定向用户到OAuth2提供商登录页面
	authURL := l.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	ctx.Redirect(http.StatusFound, authURL)
}

func (l *login) Refresh(ctx *gin.Context) {
	// 从请求中获取刷新令牌
	var req types.RefreshTokenReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Errorf(ctx, "Invalid request format: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 创建一个已过期的令牌对象，但包含有效的刷新令牌
	expiredToken := &oauth2.Token{
		RefreshToken: req.RefreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // 确保令牌已过期
	}

	// 使用TokenSource刷新令牌
	tokenSource := l.oauthConfig.TokenSource(ctx, expiredToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		logger.Errorf(ctx, "Failed to refresh token: %v", err)
		common.ReplyErr(ctx, code.RefreshTokenErr)
		return
	}

	common.ReplyOk(ctx, &types.RefreshTokenResp{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		ExpiresIn:    newToken.Expiry.Unix() - time.Now().Unix(),
		TokenType:    newToken.TokenType,
	})
}

func (l *login) Auth() {

}

func (l *login) Callback(c *gin.Context) {
	req := &types.CallBackReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		logger.Errorf(c, "callback param err: %+v", err)
		common.ReplyErr(c, code.CallbackParamErr)
	}

	// 验证state是否存在于Redis中
	stateKey := fmt.Sprintf("oauth_state:%s", req.State)
	redisResult := redis.GetClient().Get(c, stateKey)
	if redisResult.Err() != nil {
		common.ReplyErr(c, code.LoginStateErr)
		return
	}

	// 删除使用过的state
	redis.GetClient().Del(c, stateKey)
	// 用授权码交换token
	token, err := l.oauthConfig.Exchange(c, req.Code, oauth2.AccessTypeOffline)
	if err != nil {
		logger.Errorf(c, "Token exchange failed: %v", err)
		common.ReplyErr(c, code.ExchangeTokenErr)
		return
	}

	// 检查是否收到刷新令牌
	if token.RefreshToken == "" {
		logger.Warnf(c, "No refresh token received from Casdoor")
	} else {
		logger.Infof(c, "Successfully received refresh token from Casdoor")
	}

	// 使用token构建OAuth2客户端
	client := l.oauthConfig.Client(c, token)

	// 获取用户信息
	resp, err := client.Get(l.userInfoURL)
	if err != nil {
		logger.Errorf(c, "Failed to get user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	defer resp.Body.Close()

	// 解析用户信息
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf(c, "Failed to parse user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// 检查API调用是否成功
	if status, ok := result["status"].(string); !ok || status != "ok" {
		logger.Errorf(c, "Failed to get valid user info, result: %+v", result)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get valid user info"})
		return
	}

	// 提取用户数据
	userData, ok := result["data"].(map[string]interface{})
	if !ok {
		logger.Errorf(c, "Invalid user data format, result: %+v", result)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data format"})
		return
	}

	// 将用户信息和token返回给前端
	c.JSON(http.StatusOK, gin.H{
		"user":          userData,
		"token":         token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expires_in":    token.Expiry.Unix() - time.Now().Unix(),
		"token_type":    token.TokenType,
	})
}
