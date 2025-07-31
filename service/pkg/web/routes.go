package web

import (
	"fmt"

	_ "github.com/scienceol/studio/service/docs" // 导入自动生成的 docs 包
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/web/views/laboratory"
	"github.com/scienceol/studio/service/pkg/web/views/material"

	"github.com/scienceol/studio/service/pkg/web/views"
	"github.com/scienceol/studio/service/pkg/web/views/foo"
	"github.com/scienceol/studio/service/pkg/web/views/login"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(g *gin.Engine) {
	installMiddleware(g)
	InstallURL(g)
}

func installMiddleware(g *gin.Engine) {
	// TODO: trace 中间件
	g.ContextWithFallback = true
	server := webapp.Config().Server
	g.Use(cors.Default())
	g.Use(otelgin.Middleware(fmt.Sprintf("%s-%s",
		server.Platform,
		server.Service)))
	g.Use(logger.LogWithWriter())
}

func InstallURL(g *gin.Engine) {
	api := g.Group("/api")
	api.GET("/health", views.Health)
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 登录模块
	{
		l := login.NewLogin()
		// 设置认证相关路由
		authGroup := api.Group("/auth")
		// 登录路由 - 使用改进版处理器
		authGroup.GET("/login", l.Login)
		// OAuth2 回调处理路由 - 使用改进版处理器
		authGroup.GET("/callback/casdoor", l.Callback)
		// 刷新令牌路由
		authGroup.POST("/refresh", l.Refresh)
	}

	{
		v1 := api.Group("/v1")
		// 设置测试路由
		fooGroup := v1.Group("/foo")
		// 设置一个需要认证的路由 - 使用 RequireAuth 中间件进行验证
		fooGroup.GET("/hello", auth.Auth, foo.HandleHelloWorld)

		// 环境相关
		{
			labRouter := v1.Group("/lab", auth.Auth)
			labHandle := laboratory.NewEnvironment()
			labRouter.POST("", labHandle.CreateLabEnv)
			labRouter.PATCH("", labHandle.UpdateLabEnv)
			labRouter.POST("/registry", labHandle.CreateLabReg)

			materialHandle := material.NewMaterialHandle()
			labRouter.POST("/material", materialHandle.CreateLabMaterial)
			labRouter.POST("/material/edge", materialHandle.CreateMaterialEdge)
			labRouter.GET("/material/ws", materialHandle.LabMaterial)
		}
	}
}
