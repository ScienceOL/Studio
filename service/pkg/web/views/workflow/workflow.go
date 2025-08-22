package workflow

import (
	"context"
	"errors"

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

type workflowHandle struct {
	wsClient *melody.Melody
	wService workflow.Service
}

func NewWorkflowHandle(ctx context.Context) *workflowHandle {
	wsClient := melody.New()
	wsClient.Config.MaxMessageSize = constant.MaxMessageSize
	// mService := impl.NewMaterial(wsClient)
	// 注册集群通知

	h := &workflowHandle{
		wsClient: wsClient,
		wService: impl.New(ctx, wsClient),
	}

	h.initMaterialWebSocket()
	return h
}

// 工作流模板列表
func (w *workflowHandle) TemplateList(ctx *gin.Context) {
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

// 工作流模板详情
func (w *workflowHandle) TemplateDetail(ctx *gin.Context) {}

// 工作流模板 fork
func (w *workflowHandle) ForkTemplate(ctx *gin.Context) {}

// 节点模板列表，节点模板分类
func (w *workflowHandle) NodeTemplateList(ctx *gin.Context) {}

// 节点模板详情
func (w *workflowHandle) NodeTemplateDetail(ctx *gin.Context) {
	req := &workflow.LabWorkflow{}
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
func (w *workflowHandle) UpdateNodeTemplate(ctx *gin.Context) {}

// 我创建的工作流
func (w *workflowHandle) Create(ctx *gin.Context) {
	req := &workflow.WorkflowReq{}
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
func (w *workflowHandle) GetWorkflowList(ctx *gin.Context) {
	req := &workflow.WorkflowListReq{}
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
func (w *workflowHandle) GetWorkflowDetail(ctx *gin.Context) {
	req := &workflow.LabWorkflow{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("workflow uuid is empty"))
		return
	}

	if res, err := w.wService.GetWorkflowDetail(ctx, req.UUID); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

func (m *workflowHandle) initMaterialWebSocket() {
	m.wsClient.HandlePong(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "==================== pong=====================")
		}
	})
	m.wsClient.HandleClose(func(s *melody.Session, i int, s2 string) error {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "client close keys: %+v", s.Keys)
		}
		return nil
	})

	m.wsClient.HandleDisconnect(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "client closed keys: %+v", s.Keys)
		}
	})

	m.wsClient.HandleError(func(s *melody.Session, err error) {
		if errors.Is(err, melody.ErrMessageBufferFull) {
			return
		}

		if closeErr, ok := err.(*websocket.CloseError); ok {
			if closeErr.Code == websocket.CloseGoingAway {
				return
			}
		}

		if ctx, ok := s.Get("ctx"); ok {
			logger.Errorf(ctx.(context.Context), "websocket find keys: %+v, err: %+v", s.Keys, err)
		}
	})

	m.wsClient.HandleConnect(func(s *melody.Session) {
		c, _ := s.Get("ctx")
		if err := m.wService.OnWSConnect(c.(*gin.Context), s); err != nil {
			logger.Errorf(c.(*gin.Context), "check param err: %+v", err)
			s.CloseWithMsg([]byte(err.Error()))
		}
	})

	m.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		c, _ := s.Get("ctx")
		if err := m.wService.OnWSMsg(c.(*gin.Context), s, b); err != nil {
			logger.Errorf(c.(*gin.Context), "material handle msg err: %+v", err)
		}
	})

	m.wsClient.HandleSentMessage(func(s *melody.Session, b []byte) {
		// 发送完消息后的回调
	})

	m.wsClient.HandleSentMessageBinary(func(s *melody.Session, b []byte) {
		// 发送完二进制消息后的回调
		// 如果发送的是字符串消息，上面的 HandleSentMessage 也会被回调
	})
}

// 工作流 websocket
func (w *workflowHandle) LabWorkflow(ctx *gin.Context) {
	req := &workflow.LabWorkflow{}
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
	w.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		"uuid": req.UUID,
		"ctx":  ctx,
	})
}
