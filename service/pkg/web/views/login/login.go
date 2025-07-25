package login

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	ls "github.com/scienceol/studio/service/pkg/core/login"
	"github.com/scienceol/studio/service/pkg/core/login/casdoor"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"golang.org/x/oauth2"
)

type Login struct {
	oauthConfig *oauth2.Config
	lService    ls.Service
}

func NewLogin() *Login {
	return &Login{
		oauthConfig: auth.GetOAuthConfig(),
		lService:    casdoor.NewCasDoorLogin(),
	}
}

// @Summary 登录
// @Description 检查服务运行状态
// @Tags 登录模块
// @Accept json
// @Produce json
// @Success 302 {string} string "重定向到OAuth2授权页面"
// @Header 302 {string} Location "重定向的授权URL地址"
// @Router /api/auth/login [get]
func (l *Login) Login(ctx *gin.Context) {
	resp, err := l.lService.Login(ctx)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}
	ctx.Redirect(http.StatusFound, resp.RedirectURL)
}

// @Summary 刷新令牌
// @Description 刷新OAuth2令牌
// @Tags 登录模块
// @Accept json
// @Produce json
// @Param refresh_token body types.RefreshTokenReq true "刷新令牌请求"
// @Success 200 {object} common.Resp{data=types.RefreshTokenResp} "刷新令牌成功 code=0"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "参数错误 code = 1011"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "刷新 token 失败 code = 1002"
// @Router /api/auth/refresh [post]
func (l *Login) Refresh(ctx *gin.Context) {
	// 从请求中获取刷新令牌
	req := &ls.RefreshTokenReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "Invalid request format: %v", err)
		common.ReplyErr(ctx, code.RefreshTokenParamErr)
		return
	}

	resp, err := l.lService.Refresh(ctx, req)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary OAuth2回调
// @Description 处理OAuth2授权回调
// @Tags 登录模块
// @Accept json
// @Produce json
// @Param code query string true "授权码"
// @Param state query string true "防CSRF攻击的状态码"
// @Success 200 {object} common.Resp "回调成功"
// @Failure 200 {object} common.Resp "服务器内部错误"
// @Router /api/auth/callback/casdoor [get]
func (l *Login) Callback(ctx *gin.Context) {
	req := &ls.CallbackReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "callback param err: %+v", err)
		common.ReplyErr(ctx, code.CallbackParamErr)
		return
	}
	resp, err := l.lService.Callback(ctx, req)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}
