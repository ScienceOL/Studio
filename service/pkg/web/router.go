package web

import (
	"context"
	"fmt"
	"time"

	_ "github.com/scienceol/studio/service/docs" // 导入自动生成的 docs 包
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/web/views/laboratory"
	"github.com/scienceol/studio/service/pkg/web/views/material"
	"github.com/scienceol/studio/service/pkg/web/views/workflow"

	"github.com/scienceol/studio/service/pkg/web/views"
	"github.com/scienceol/studio/service/pkg/web/views/action"
	"github.com/scienceol/studio/service/pkg/web/views/foo"
	"github.com/scienceol/studio/service/pkg/web/views/labstatus"
	"github.com/scienceol/studio/service/pkg/web/views/login"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(ctx context.Context, g *gin.Engine) {
	installMiddleware(g)
	InstallURL(ctx, g)
}

func installMiddleware(g *gin.Engine) {
	g.ContextWithFallback = true
	server := config.Global().Server
	// g.Use(cors.Default())

	// 配置 CORS，明确允许 authorization 请求头
	g.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:32234", "http://localhost:*", "https://sciol.ac.cn", "https://*.sciol.ac.cn"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	g.Use(otelgin.Middleware(fmt.Sprintf("%s-%s",
		server.Platform,
		server.Service)))
	g.Use(logger.LogWithWriter())
}

func InstallURL(ctx context.Context, g *gin.Engine) {
	api := g.Group("/api")
	api.GET("/health", views.Health)
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 在服务启动时初始化实验室状态服务，确保注册到全局通知器
	labStatusHandle := labstatus.New()
	logger.Infof(ctx, "lab status service initialized and registered to global notifier")

	// Test
	{
		fooGroup := api.Group("/foo")
		fooGroup.GET("", foo.HandleTestAnomy)
		fooGroup.GET("/auth", auth.Auth(), foo.HandleTestAuth)
	}

	// Auth
	{
		l := login.NewLogin()
		// 设置认证相关路由
		authGroup := api.Group("/auth")
		// 登录路由
		authGroup.GET("/login", l.Login)
		// OAuth2 回调处理路由
		authGroup.GET("/callback/casdoor", l.Callback)
		// 刷新令牌路由
		authGroup.POST("/refresh", l.Refresh)
	}

	// V1 API
	{
		v1 := api.Group("/v1")
		wsRouter := v1.Group("/ws", auth.Auth())

		// 实验室状态 WebSocket
		{
			wsRouter.GET("/lab/status", labStatusHandle.ConnectLabStatus)
		}

		// 环境相关
		{
			labRouter := v1.Group("/lab", auth.Auth())

			{
				labHandle := laboratory.NewEnvironment()
				labRouter.POST("", labHandle.CreateLabEnv)                                 // 创建实验室
				labRouter.PATCH("", labHandle.UpdateLabEnv)                                // 更新实验室
				labRouter.GET("/list", labHandle.LabList)                                  // 获取当前用户的所有实验室
				labRouter.GET("/info/:uuid", labHandle.LabInfo)                            // 获取当前用户的所有实验室
				labRouter.POST("/resource", labHandle.CreateLabResource)                   // 从 edge 侧创建资源
				labRouter.GET("/member/:lab_uuid", labHandle.GetLabMemeber)                // 根据实验室获取当前实验室成员
				labRouter.DELETE("/member/:lab_uuid/:member_uuid", labHandle.DelLabMember) // 删除实验室成员
				labRouter.POST("/invite/:lab_uuid", labHandle.CreateInvite)                // 创建邀请链接
				labRouter.GET("/invite/:uuid", labHandle.AcceptInvite)                     // 接受邀请链接
				labRouter.GET("/user/info", labHandle.UserInfo)                            // 获取用户信息
			}

			{
				materialRouter := labRouter.Group("/material")
				materialHandle := material.NewMaterialHandle(ctx)
				materialRouter.POST("", materialHandle.CreateLabMaterial)                      //  创建物料 done
				materialRouter.GET("", materialHandle.QueryMaterial)                           // edge 侧查询物料资源
				materialRouter.PUT("", materialHandle.BatchUpdateMaterial)                     // edge 批量更新物料数据
				materialRouter.POST("/save", materialHandle.SaveMaterial)                      //  保存物料
				materialRouter.GET("/resource", materialHandle.ResourceList)                   // 获取该实验室所有设备列表（简化版）
				materialRouter.GET("/resource/templates", materialHandle.ResourceTemplateList) // 获取资源模板详细信息（包含 actions）
				materialRouter.GET("/device/actions", materialHandle.Actions)                  // 获取实验室所有动作
				materialRouter.POST("/edge", materialHandle.CreateMaterialEdge)                // 创建物料连线 done
				materialRouter.GET("/download/:lab_uuid", materialHandle.DownloadMaterial)     // 下载物料dag done
				materialRouter.GET("/template/:template_uuid", materialHandle.Template)
				// labRouter.GET("/ws/material/:lab_uuid", materialHandle.LabMaterial) // WARN: websocket 是否要放在统一的路由下

				// 后续待优化, 单独拆出去。
				{
					// 实验室 edge 上报接口
					edgeRouter := v1.Group("/edge", auth.Auth())
					materialRouter := edgeRouter.Group("/material")
					materialRouter.POST("", materialHandle.EdgeCreateMaterial)
					materialRouter.PUT("", materialHandle.EdgeUpsertMaterial) // 更新 & 创建
					materialRouter.POST("/edge", materialHandle.EdgeCreateEdge)
					materialRouter.POST("/query", materialHandle.QueryMaterialByUUID)
					materialRouter.GET("/download", materialHandle.EdgeDownloadMaterial)
					// materialRouter.PATCH("", materialHandle.EdgeCreateMaterial)
				}

				wsRouter.GET("/material/:lab_uuid", materialHandle.LabMaterial)

			}

			{
				// 动作执行
				actionHandle := action.NewActionHandle(ctx)
				actionRouter := labRouter.Group("/action")
				actionRouter.POST("/run", actionHandle.RunAction)               // 手动执行设备动作
				actionRouter.GET("/result/:uuid", actionHandle.GetActionResult) // 查询动作执行结果

				// WebSocket 放在独立的 wsRouter 下
				wsRouter.GET("/action/:task_uuid", actionHandle.ActionWebSocket) // WebSocket 实时状态更新
			}

			{
				workflowHandle := workflow.NewWorkflowHandle(ctx)
				workflowRouter := labRouter.Group("/workflow")
				workflowRouter.GET("/task/:uuid", workflowHandle.TaskList)              // 工作流 task 列表 done
				workflowRouter.GET("/task/download/:uuid", workflowHandle.DownloadTask) // 工作流任务下载 done

				{
					// 工作流模板
					tpl := workflowRouter.Group("/template")
					tpl.GET("/detail/:uuid", workflowHandle.GetWorkflowDetail)           // 获取工作流模板详情
					tpl.PUT("/fork", workflowHandle.ForkTemplate)                        // fork 工作流 done
					tpl.GET("/tags", workflowHandle.WorkflowTemplateTags)                // 获取工作流 tags done
					tpl.GET("/tags/:lab_uuid", workflowHandle.WorkflowTemplateTagsByLab) // 按实验室获取工作流模板标签
					tpl.GET("/list", workflowHandle.WorkflowTemplateList)                // 获取工作流模板列表 done
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
					owner.PATCH("", workflowHandle.UpdateWorkflow)     // 更新工作流 done
					owner.POST("", workflowHandle.Create)              // 创建工作流 done
					owner.DELETE("/:uuid", workflowHandle.DelWorkflow) //  删除自己创建的工作流 done
					owner.GET("/list", workflowHandle.GetWorkflowList) // 获取工作流列表  done
					owner.GET("/export", workflowHandle.Export)        // 导出工作流
					owner.POST("/import", workflowHandle.Import)       // 导入工作流
					owner.PUT("/duplicate", workflowHandle.Duplicate)  // 复制工作流
				}

				v1.PUT("/lab/run/workflow", workflowHandle.RunWorkflow)

				workflowRouter.GET("/ws/workflow/:uuid", workflowHandle.LabWorkflow) // TODO: websocket 放在统一的路由下
			}
		}
	}
}
