package web

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/web/views"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(g *gin.Engine) {
	installMiddleware(g)
	InstallURL(g)

	g.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]any{
			"success": "ok",
		})
	})

	// FIXME: 安装鉴权中间件
	{
		lab := g.Group("/lab")
		labHandel := views.NewLabHandle()
		lab.GET("/envs", labHandel.GetEnv)
	}

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
