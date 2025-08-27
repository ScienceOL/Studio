package workflow

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/core/workflow"
	impl "github.com/scienceol/studio/service/pkg/core/workflow/workflow"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type Handle struct {
	wsClient *melody.Melody
	wService workflow.Service
}

func NewWorkflowHandle(ctx context.Context) *Handle {
	wsClient := melody.New()
	wsClient.Config.MaxMessageSize = constant.MaxMessageSize
	// mService := impl.NewMaterial(wsClient)
	// 注册集群通知

	h := &Handle{
		wsClient: wsClient,
		wService: impl.New(ctx, wsClient),
	}

	h.initMaterialWebSocket()
	return h
}

// 工作流模板列表
func (w *Handle) TemplateList(ctx *gin.Context) {
	req := workflow.TplPageReq{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	if res, err := w.wService.TemplateList(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

func (w *Handle) TemplateTags(ctx *gin.Context) {
	req := workflow.TemplateTagsReq{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	if res, err := w.wService.TemplateTags(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// 工作流模板详情
func (w *Handle) TemplateDetail(ctx *gin.Context) {}

// 工作流模板 tags
func (w *Handle) WorkflowTemplateTags(ctx *gin.Context) {
	if res, err := w.wService.WorkflowTemplateTags(ctx); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// 工作流模板列表
func (w *Handle) WorkflowTemplateList(ctx *gin.Context) {
	req := workflow.TemplateListReq{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.WorkflowTemplateList(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// 工作流模板 fork
func (w *Handle) ForkTemplate(ctx *gin.Context) {
	req := &workflow.ForkReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if err := w.wService.ForkWrokflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx)
	}
}

// 获取工作流 task 列表
func (w *Handle) TaskList(ctx *gin.Context) {
	req := workflow.TaskReq{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.WorkflowTaskList(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// 下载 task
func (w *Handle) DownloadTask(ctx *gin.Context) {
	req := workflow.TaskDownloadReq{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.TaskDownload(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		ctx.Header("Content-Disposition", "attachment; filename=task.csv")
		ctx.Header("Content-Type", "text/csv")
		ctx.Header("Pragma", "public")
		ctx.Header("Content-Length", fmt.Sprintf("%d", len(res.Bytes())))

		// 发送文件数据
		ctx.Data(http.StatusOK, "text/csv", res.Bytes())
	}
}

// 节点模板列表，节点模板分类

// 节点模板详情
func (w *Handle) NodeTemplateDetail(ctx *gin.Context) {
	req := &workflow.NodeTemplateReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("template uuid is empty"))
		return
	}

	if res, err := w.wService.NodeTemplateDetail(ctx, req.UUID); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// 节点模板编辑
func (w *Handle) UpdateWorkflow(ctx *gin.Context) {
	req := &workflow.UpdateReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("workflow uuid is empty"))
		return
	}

	if err := w.wService.UpdateWorkflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx)
	}
}

// 删除工作流
func (w *Handle) DelWrokflow(ctx *gin.Context) {
	req := &workflow.DelReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if err := w.wService.DelWorkflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx)
	}
}

// 我创建的工作流
func (w *Handle) Create(ctx *gin.Context) {
	req := &workflow.CreateReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.Create(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// GetWorkflowList 获取工作流列表
func (w *Handle) GetWorkflowList(ctx *gin.Context) {
	req := &workflow.ListReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.GetWorkflowList(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// GetWorkflowDetail 获取工作流详情
func (w *Handle) GetWorkflowDetail(ctx *gin.Context) {
	req := &workflow.DetailReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("workflow uuid is empty"))
		return
	}

	if res, err := w.wService.GetWorkflowDetail(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

func (w *Handle) initMaterialWebSocket() {
	w.wsClient.HandlePong(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "==================== pong=====================")
		}
	})
	w.wsClient.HandleClose(func(s *melody.Session, _ int, _ string) error {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "client close keys: %+v", s.Keys)
		}
		return nil
	})

	w.wsClient.HandleDisconnect(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "client closed keys: %+v", s.Keys)
		}
	})

	w.wsClient.HandleError(func(s *melody.Session, err error) {
		if errors.Is(err, melody.ErrMessageBufferFull) {
			return
		}

		if closeErr, ok := err.(*websocket.CloseError); ok {
			if closeErr.Code == websocket.CloseGoingAway {
				return
			}
		}

		if strings.Contains(err.Error(), "use of closed network connection") {
			return
		}

		if ctx, ok := s.Get("ctx"); ok {
			logger.Errorf(ctx.(context.Context), "websocket find keys: %+v, err: %+v", s.Keys, err)
		}
	})

	w.wsClient.HandleConnect(func(s *melody.Session) {
		c, _ := s.Get("ctx")
		if err := w.wService.OnWSConnect(c.(*gin.Context), s); err != nil {
			logger.Errorf(c.(*gin.Context), "check param err: %+v", err)
			if err := s.CloseWithMsg([]byte(err.Error())); err != nil {
				logger.Errorf(c.(*gin.Context), "workflow HandleConnect CloseWithMsg err: %+v", err)
			}
		}
	})

	w.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		c, _ := s.Get("ctx")
		if err := w.wService.OnWSMsg(c.(*gin.Context), s, b); err != nil {
			logger.Errorf(c.(*gin.Context), "material handle msg err: %+v", err)
		}
	})

	w.wsClient.HandleSentMessage(func(_ *melody.Session, _ []byte) {
		// 发送完消息后的回调
	})

	w.wsClient.HandleSentMessageBinary(func(_ *melody.Session, _ []byte) {
		// 发送完二进制消息后的回调
		// 如果发送的是字符串消息，上面的 HandleSentMessage 也会被回调
	})
}

// 工作流 websocket
func (w *Handle) LabWorkflow(ctx *gin.Context) {
	req := &workflow.NodeTemplateReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		logger.Errorf(ctx, "unmarshal uuid err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		logger.Errorf(ctx, "unmarshal uuid is empty")
		common.ReplyErr(ctx, code.ParamErr.WithMsg("uuid is empty"))
		return
	}

	// 阻塞运行
	if err := w.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		"uuid": req.UUID,
		"ctx":  ctx,
	}); err != nil {
		logger.Errorf(ctx, "workflow HandleRequestWithKeys err: %+v", err)
	}
}
