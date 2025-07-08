package web

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/middleware/auth"

	"github.com/scienceol/studio/service/pkg/web/views/foo"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(g *gin.Engine) {
	installMiddleware(g)
	InstallURL(g)

	api := g.Group("/api")

	api.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]any{
			"success": "ok",
		})
	})

	// 设置认证相关路由
	authGroup := api.Group("/auth")
	// 登录路由 - 使用改进版处理器
	authGroup.GET("/login", auth.HandleLogin())

	// OAuth2 回调处理路由 - 使用改进版处理器
	authGroup.GET("/callback/casdoor", auth.HandleCallback())

	// 刷新令牌路由
	authGroup.POST("/refresh", auth.HandleRefresh())

	// 设置测试路由
	fooGroup := api.Group("/foo")
	// 设置一个需要认证的路由 - 使用 RequireAuth 中间件进行验证
	fooGroup.GET("/hello", auth.RequireAuth(), foo.HandleHelloWorld())

	// {
	// 	lab := g.Group("/lab")
	// 	labHandel := views.NewLabHandle()
	// 	lab.GET("/envs", labHandel.GetEnv)
	// }

}

func installMiddleware(g *gin.Engine) {
	// TODO: trace 中间件
	g.ContextWithFallback = true
	server := webapp.Config().Server
	g.Use(cors.Default())
	g.Use(otelgin.Middleware(fmt.Sprintf("%s-%s",
		server.Platform,
		server.Service)))

}

func InstallURL(g *gin.Engine) {

}
