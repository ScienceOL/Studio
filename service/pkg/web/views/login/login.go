package login

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

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
// @Tags Authentication
// @Accept json
// @Produce json
// @Param frontend_callback_url query string false "前端回调地址"
// @Success 302 {string} string "重定向到OAuth2授权页面"
// @Header 302 {string} Location "重定向的授权URL地址"
// @Router /api/auth/login [get]
func (l *Login) Login(ctx *gin.Context) {
	req := &ls.LoginReq{}
	// 从查询参数获取前端回调地址（可选）
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "Invalid login request: %v", err)
	}

	resp, err := l.lService.Login(ctx, req)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}
	ctx.Redirect(http.StatusFound, resp.RedirectURL)
}

// @Summary 刷新令牌
// @Description 刷新OAuth2令牌
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh_token body ls.RefreshTokenReq true "刷新令牌请求"
// @Success 200 {object} common.Resp{data=ls.RefreshTokenResp} "刷新令牌成功 code=0"
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
// @Tags Authentication
// @Accept json
// @Produce json
// @Param code query string true "授权码"
// @Param state query string true "防CSRF攻击的状态码"
// @Success 302 {string} string "重定向到前端"
// @Failure 302 {string} string "重定向到前端错误页面"
// @Router /api/auth/callback/casdoor [get]
func (l *Login) Callback(ctx *gin.Context) {
	req := &ls.CallbackReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "callback param err: %+v", err)
		// 获取默认前端地址
		frontendBaseURL := getDefaultFrontendURL()
		// 重定向到前端错误页面
		errorURL := fmt.Sprintf("%s/login/callback?error=%s",
			frontendBaseURL, url.QueryEscape("参数解析错误"))
		ctx.Redirect(http.StatusFound, errorURL)
		return
	}

	resp, err := l.lService.Callback(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "callback service err: %+v", err)
		// 获取默认前端地址
		frontendBaseURL := getDefaultFrontendURL()
		// 重定向到前端错误页面
		errorMsg := "登录处理失败"
		if err.Error() != "" {
			errorMsg = err.Error()
		}
		errorURL := fmt.Sprintf("%s/login/callback?error=%s",
			frontendBaseURL, url.QueryEscape(errorMsg))
		ctx.Redirect(http.StatusFound, errorURL)
		return
	}

	// OAuth2最佳实践: 使用 Cookie 存储 token，避免 URL 过长问题

	// 判断是否为 HTTPS（生产环境）
	isSecure := ctx.Request.TLS != nil || ctx.GetHeader("X-Forwarded-Proto") == "https"

	// 设置 access_token cookie
	// 注意：为了让前端 JavaScript 能读取，httpOnly 设为 false
	// 在生产环境建议使用 httpOnly=true，并通过后端 API 来验证 token
	ctx.SetCookie(
		"access_token",      // name
		resp.Token,          // value
		int(resp.ExpiresIn), // maxAge (seconds)
		"/",                 // path
		"",                  // domain (empty = current domain)
		isSecure,            // secure (HTTPS 时为 true)
		false,               // httpOnly = false，允许 JavaScript 读取
	)

	// 设置 refresh_token cookie
	ctx.SetCookie(
		"refresh_token",   // name
		resp.RefreshToken, // value
		30*24*60*60,       // maxAge: 30天
		"/",               // path
		"",                // domain
		isSecure,          // secure
		false,             // httpOnly = false，允许 JavaScript 读取
	)

	// 将用户信息以 JSON 格式存储在 cookie 中
	if resp.User != nil {
		logger.Infof(ctx, "Marshaling user data to cookie: %+v", resp.User)

		// 只存储核心用户信息，避免 Cookie 过大
		userInfo := map[string]interface{}{
			"id":          resp.User.ID,
			"name":        resp.User.Name,
			"displayName": resp.User.DisplayName,
			"email":       resp.User.Email,
			"avatar":      resp.User.Avatar,
			"type":        resp.User.Type,
			"owner":       resp.User.Owner,
			"phone":       resp.User.Phone,
		}

		userJSON, err := json.Marshal(userInfo)
		if err == nil {
			logger.Infof(ctx, "User JSON for cookie: %s", string(userJSON))
			ctx.SetCookie(
				"user_info",
				base64.URLEncoding.EncodeToString(userJSON),
				int(resp.ExpiresIn),
				"/",
				"",
				isSecure, // secure
				false,    // httpOnly = false，允许前端读取
			)
			logger.Infof(ctx, "Successfully set user_info cookie")
		} else {
			logger.Errorf(ctx, "Failed to marshal user data: %v", err)
		}
	} else {
		logger.Errorf(ctx, "resp.User is nil, cannot set user_info cookie")
	}

	logger.Infof(ctx, "Set cookies: access_token, refresh_token, user_info (secure=%v)", isSecure)

	// 重定向到前端，只传递简单的状态参数
	params := url.Values{}
	params.Set("status", "success")

	frontendURL := fmt.Sprintf("%s?%s", resp.FrontendCallbackURL, params.Encode())
	ctx.Redirect(http.StatusFound, frontendURL)
}

// getDefaultFrontendURL 获取默认的前端地址
func getDefaultFrontendURL() string {
	frontendURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:32234"
	}
	return frontendURL
}
