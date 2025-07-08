package login

import "context"

type LoginService interface {
	Login(ctx context.Context) (*LoginResp, error)
	Refresh(ctx context.Context, req *RefreshTokenReq) (*RefreshTokenResp, error)
	Callback(c context.Context, req *CallbackReq) (*CallbackResp, error)
}
