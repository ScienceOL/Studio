package web

import (
	"context"
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
	"github.com/scienceol/studio/service/pkg/web/views/workflow"

	"github.com/scienceol/studio/service/pkg/web/views"
	"github.com/scienceol/studio/service/pkg/web/views/foo"
	"github.com/scienceol/studio/service/pkg/web/views/login"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(ctx context.Context, g *gin.Engine) {
	installMiddleware(g)
	InstallURL(ctx, g)
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

func InstallURL(ctx context.Context, g *gin.Engine) {
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
		fooGroup.GET("/hello", auth.Auth(), foo.HandleHelloWorld)

		// 环境相关
		{
			labRouter := v1.Group("/lab", auth.Auth())

			{
				labHandle := laboratory.NewEnvironment()
				labRouter.POST("", labHandle.CreateLabEnv)                                 // 创建实验室
				labRouter.PATCH("", labHandle.UpdateLabEnv)                                // 更新实验室
				labRouter.GET("/list", labHandle.LabList)                                  // 获取当前用户的所有实验室
				labRouter.POST("/resource", labHandle.CreateLabResource)                   // 从 edge 侧创建资源
				labRouter.GET("/member/:lab_uuid", labHandle.GetLabMemeber)                // 根据实验室获取当前实验室成员
				labRouter.DELETE("/member/:lab_uuid/:member_uuid", labHandle.DelLabMember) // 删除实验室成员
				labRouter.POST("/invite/:lab_uuid", labHandle.CreateInvite)                // 创建邀请链接
				labRouter.GET("/invite/:uuid", labHandle.AcceptInvite)                     // 接受邀请链接
			}

			{
				materialHandle := material.NewMaterialHandle(ctx)
				materialRouter := labRouter.Group("/material")
				materialRouter.POST("", materialHandle.CreateLabMaterial)                  //  创建物料 done
				materialRouter.POST("/edge", materialHandle.CreateMaterialEdge)            // 创建物料连线 done
				materialRouter.GET("/download/:lab_uuid", materialHandle.DownloadMaterial) // 下载物料dag done

				labRouter.GET("/ws/material/:lab_uuid", materialHandle.LabMaterial) // WARN: websocket 是否要放在统一的路由下
			}

			{
				workflowHandle := workflow.NewWorkflowHandle(ctx)
				workflowRouter := labRouter.Group("/workflow")
				workflowRouter.GET("/task/:uuid", workflowHandle.TaskList)              // 工作流 task 列表 done
				workflowRouter.GET("/task/download/:uuid", workflowHandle.DownloadTask) // 工作流任务下载 done

				{
					// 工作流模板
					tpl := workflowRouter.Group("/template")
					tpl.GET("/detail/:uuid", workflowHandle.GetWorkflowDetail) // 获取工作流模板详情
					tpl.PUT("/fork", workflowHandle.ForkTemplate)              // fork 工作流 TODO:
					tpl.GET("/tags", workflowHandle.WorkflowTemplateTags)      // 获取工作流 tags done
					tpl.GET("/list", workflowHandle.WorkflowTemplateList)      // 获取工作流模板列表 done
				}
				{
					// 工作流节点模板
					nodeTpl := workflowRouter.Group("/node/template")
					nodeTpl.GET("/tags/:lab_uuid", workflowHandle.TemplateTags)     // 节点模板 tags done
					nodeTpl.GET("/list", workflowHandle.TemplateList)               // 模板列表 done
					nodeTpl.GET("/detail/:uuid", workflowHandle.NodeTemplateDetail) // 节点模板详情 done

				}
				{
					// 我的工作流
					owner := workflowRouter.Group("owner")
					owner.PATCH("", workflowHandle.UpdateWorkflow)        // 更新工作流 done
					owner.POST("", workflowHandle.Create)                 // 创建工作流 done
					owner.DELETE("/:uuid", workflowHandle.DelWrokflow)    // 删除自己创建的工作流 done
					owner.GET("/list", workflowHandle.GetWorkflowList)    // 获取工作流列表  done
					owner.GET("/export", workflowHandle.GetWorkflowList)  // 导出工作流 TODO:
					owner.POST("/import", workflowHandle.GetWorkflowList) // 导入工作流 TODO:
				}

				workflowRouter.GET("/ws/workflow/:uuid", workflowHandle.LabWorkflow) // WARN: websocket 是否放在统一的路由下
			}
		}
	}
}
