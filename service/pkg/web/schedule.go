package web

import (
	"context"

	_ "github.com/scienceol/studio/service/docs" // 导入自动生成的 docs 包
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/web/views/schedule"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"

	"github.com/scienceol/studio/service/pkg/web/views"
)

func NewSchedule(ctx context.Context, g *gin.Engine) context.CancelFunc {
	installMiddleware(g)
	return InstallScheduleURL(ctx, g)
}

func InstallScheduleURL(ctx context.Context, g *gin.Engine) context.CancelFunc {
	api := g.Group("/api")
	api.GET("/health", views.Health)
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	handle := schedule.New(ctx)

	{
		v1 := api.Group("/v1")
		v1.GET("/lab", auth.AuthLab(), handle.Connect)
	}

	return func() {
		handle.Close(ctx)
	}
}
