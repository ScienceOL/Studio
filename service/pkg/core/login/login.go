package login

import "context"

type Service interface {
	Login(ctx context.Context, req *LoginReq) (*Resp, error)
	Refresh(ctx context.Context, req *RefreshTokenReq) (*RefreshTokenResp, error)
	Callback(c context.Context, req *CallbackReq) (*CallbackResp, error)
}
