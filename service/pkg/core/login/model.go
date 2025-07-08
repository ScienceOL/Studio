package login

import "github.com/scienceol/studio/service/pkg/middleware/auth"

type LoginResp struct {
	RedirectURL string `json:"redirect_url"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type CallbackReq struct {
	Code  string `json:"code" form:"code" binding:"required"`
	State string `json:"state" form:"state" binding:"required"`
}

type CallbackResp struct {
	User         *auth.UserData `json:"user"`
	Token        string         `json:"token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
}
