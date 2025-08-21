package web

import (
	"context"

	_ "github.com/scienceol/studio/service/docs" // 导入自动生成的 docs 包
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"

	"github.com/scienceol/studio/service/pkg/web/views"
)

func NewSchedule(ctx context.Context, g *gin.Engine) {
	installMiddleware(g)
	InstallScheduleURL(ctx, g)
}

func InstallScheduleURL(ctx context.Context, g *gin.Engine) {
	api := g.Group("/api")
	api.GET("/health", views.Health)
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	// websocket 启动

	{
		v1 := api.Group("/v1")
		_ = v1
	}

}
