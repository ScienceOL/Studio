package auth

import (
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"golang.org/x/oauth2"
)

type UserData struct {
	Owner             string `json:"owner"`
	Name              string `json:"name"`
	ID                string `json:"id"`
	Avatar            string `json:"avatar"`
	Type              string `json:"type"`
	DisplayName       string `json:"display_name"`
	SignupApplication string `json:"signup_application"`
	AccessToken       string `json:"access_token"`
}

type UserInfo struct {
	Status string    `json:"status"`
	Msg    string    `json:"msg"`
	Sub    string    `json:"sub"`
	Name   string    `json:"name"`
	Data   *UserData `json:"data,omitempty"`
}

type Config struct {
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

	USERKEY = "AUTH_USER_KEY"
)

// GetOAuthConfig 获取OAuth2配置
func GetOAuthConfig() *oauth2.Config {
	if oauthConfig == nil {
		authConf := webapp.Config().OAuth2
		oauthConfig = &oauth2.Config{
			ClientID:     authConf.ClientID,
			ClientSecret: authConf.ClientSecret,
			Scopes:       authConf.Scopes,
			Endpoint: oauth2.Endpoint{
				TokenURL: authConf.TokenURL,
				AuthURL:  authConf.AuthURL,
			},
			RedirectURL: authConf.RedirectURL,
		}
	}

	return oauthConfig
}
