package auth

import (
	"context"
	"sync"

	"golang.org/x/oauth2"
)

type AuthConfig struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
	TokenURL     string
	AuthURL      string
	RedirectURL  string
	UserInfoURL  string
}

var (
	oauthConfig *oauth2.Config
	once        sync.Once
)

// InitOAuth 初始化OAuth2配置
func InitOAuth(ctx context.Context, config *AuthConfig) error {
	var initErr error
	once.Do(func() {
		oauthConfig = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       config.Scopes,
			Endpoint: oauth2.Endpoint{
				TokenURL: config.TokenURL,
				AuthURL:  config.AuthURL,
			},
			RedirectURL: config.RedirectURL,
		}
	})
	return initErr
}

// GetOAuthConfig 获取OAuth2配置
func GetOAuthConfig() *oauth2.Config {
	return oauthConfig
}
