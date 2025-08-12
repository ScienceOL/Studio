package workflow

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/core/workflow"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type workflowHandle struct {
	wsClient *melody.Melody
	wService workflow.Service
}

func NewWorkflowHandle() *workflowHandle {
	wsClient := melody.New()
	wsClient.Config.MaxMessageSize = constant.MaxMessageSize
	// mService := impl.NewMaterial(wsClient)
	// 注册集群通知

	h := &workflowHandle{
		wsClient: wsClient,
	}

	h.initMaterialWebSocket()
	return &workflowHandle{}
}

// 工作流模板列表
func (w *workflowHandle) TemplateList(ctx *gin.Context) {}

// 工作流模板详情
func (w *workflowHandle) TemplateDetail(ctx *gin.Context) {}

// 工作流模板 fork
func (w *workflowHandle) ForkTemplate(ctx *gin.Context) {}

// 节点模板列表，节点模板分类
func (w *workflowHandle) NodeTemplateList(ctx *gin.Context) {}

// 节点模板详情
func (w *workflowHandle) NodeTemplateDetail(ctx *gin.Context) {}

// 节点模板编辑
func (w *workflowHandle) UpdateNodeTemplate(ctx *gin.Context) {}

// 我创建的工作流
func (w *workflowHandle) Add(ctx *gin.Context) {}

func (m *workflowHandle) initMaterialWebSocket() {
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
		if ctx, ok := s.Get("ctx"); ok {
			logger.Errorf(ctx.(context.Context), "websocket find keys: %+v, err: %+v", s.Keys, err)
		}
	})

	m.wsClient.HandleConnect(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "websocket connect keys: %+v", s.Keys)
			m.wService.OnWSConnect(ctx.(context.Context), s)
		}
	})

	m.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		ctxI, ok := s.Get("ctx")
		if !ok {
			s.CloseWithMsg([]byte("no ctx"))
			return
		}

		if err := m.wService.OnWSMsg(ctxI.(*gin.Context), s, b); err != nil {
			logger.Errorf(ctxI.(*gin.Context), "material handle msg err: %+v", err)
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
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	userInfo := auth.GetCurrentUser(ctx)

	// 阻塞运行
	w.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		auth.USERKEY: userInfo,
		"ctx":        ctx,
		"uuid":       req.UUID,
	})
}
