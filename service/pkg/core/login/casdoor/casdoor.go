package casdoor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/login"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"golang.org/x/oauth2"
)

type casdoorLogin struct {
	*r.Client
	oauthConfig *oauth2.Config
}

func NewCasDoorLogin() login.Service {
	return &casdoorLogin{
		Client:      redis.GetClient(),
		oauthConfig: auth.GetOAuthConfig(),
	}
}

func (c *casdoorLogin) Login(ctx context.Context) (*login.Resp, error) {
	state := fmt.Sprintf("%d", time.Now().UnixNano())
	// 将state保存到Redis中，设置5分钟过期时间
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := c.Set(ctx, stateKey, "valid", 5*time.Minute).Err(); err != nil {
		logger.Errorf(ctx, "Failed to save state to Redis: %v", err)
		return nil, code.LoginSetStateErr
	}

	// 构建授权URL并重定向用户到OAuth2提供商登录页面
	authURL := c.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return &login.Resp{RedirectURL: authURL}, nil
}

func (c *casdoorLogin) Refresh(ctx context.Context, req *login.RefreshTokenReq) (*login.RefreshTokenResp, error) {
	// 创建一个已过期的令牌对象，但包含有效的刷新令牌
	expiredToken := &oauth2.Token{
		RefreshToken: req.RefreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // 确保令牌已过期
	}

	// 使用TokenSource刷新令牌
	tokenSource := c.oauthConfig.TokenSource(ctx, expiredToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		logger.Errorf(ctx, "Failed to refresh token: %v", err)
		return nil, code.RefreshTokenErr
	}

	return &login.RefreshTokenResp{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		ExpiresIn:    newToken.Expiry.Unix() - time.Now().Unix(),
		TokenType:    newToken.TokenType,
	}, nil
}

func (c *casdoorLogin) Callback(ctx context.Context, req *login.CallbackReq) (*login.CallbackResp, error) {
	// 验证state是否存在于Redis中
	stateKey := fmt.Sprintf("oauth_state:%s", req.State)
	redisResult := redis.GetClient().Get(ctx, stateKey)
	if redisResult.Err() != nil {
		return nil, code.LoginStateErr
	}

	// 删除使用过的state
	redis.GetClient().Del(ctx, stateKey)
	// 用授权码交换token
	token, err := c.oauthConfig.Exchange(ctx, req.Code, oauth2.AccessTypeOffline)
	if err != nil {
		logger.Errorf(ctx, "Token exchange failed: %v", err)
		return nil, code.ExchangeTokenErr
	}

	// 检查是否收到刷新令牌
	if token.RefreshToken == "" {
		logger.Warnf(ctx, "No refresh token received from Casdoor")
	} else {
		logger.Infof(ctx, "Successfully received refresh token from Casdoor")
	}

	// 使用token构建OAuth2客户端
	client := c.oauthConfig.Client(ctx, token)

	// 获取用户信息
	resp, err := client.Get(webapp.Config().OAuth2.UserInfoURL)
	if err != nil {
		logger.Errorf(ctx, "Failed to get user info: %v", err)
		return nil, code.LoginGetUserInfoErr
	}

	defer resp.Body.Close()

	// 解析用户信息
	result := &model.UserInfo{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil || result.Status != "ok" {
		logger.Errorf(ctx, "Failed to parse user info: %v", err)
		return nil, code.LoginCallbackErr
	}
	return &login.CallbackResp{
		User:         result.Data,
		Token:        token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.Expiry.Unix() - time.Now().Unix(),
	}, nil
}
