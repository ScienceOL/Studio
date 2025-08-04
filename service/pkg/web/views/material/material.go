package material

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/material"
	impl "github.com/scienceol/studio/service/pkg/core/material/material"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"gorm.io/datatypes"
)

type Handle struct {
	mService material.Service
	wsClient *melody.Melody
}

func NewMaterialHandle() *Handle {
	wsClient := melody.New()
	mService := impl.NewMaterial(wsClient)
	// 注册集群通知
	events.NewEvents().Registry(context.Background(), notify.MaterialModify, mService.HandleNotify)

	h := &Handle{
		mService: mService,
		wsClient: wsClient,
	}

	h.initMaterialWebSocket()
	return h
}

func (m *Handle) CreateLabMaterial(ctx *gin.Context) {
	req := &material.GraphNode{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse CreateLabMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	if err := m.mService.CreateMaterial(ctx, req); err != nil {
		logger.Errorf(ctx, "CreateMaterial err: %+v", err)
		common.ReplyErr(ctx, err)

		return
	}

	common.ReplyOk(ctx)
}

func (m *Handle) CreateMaterialEdge(ctx *gin.Context) {
	req := &material.GraphEdge{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse CreateMaterialEdge param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	if err := m.mService.CreateEdge(ctx, req); err != nil {
		logger.Errorf(ctx, "CreateMaterialEdge err: %+v", err)
		common.ReplyErr(ctx, err)

		return
	}

	common.ReplyOk(ctx)
}

func (m *Handle) initMaterialWebSocket() {
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
		}
	})

	m.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		ctxI, ok := s.Get("ctx")
		if !ok {
			s.CloseWithMsg([]byte("no ctx"))
			return
		}

		if err := m.mService.HandleWSMsg(ctxI.(*gin.Context), s, b); err != nil {
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

func (m *Handle) LabMaterial(ctx *gin.Context) {
	req := &material.LabWS{}
	labUUIDStr := ctx.Param("lab_uuid")
	req.LabUUID = common.BinUUID(datatypes.BinUUIDFromString(labUUIDStr))
	userInfo := auth.GetCurrentUser(ctx)

	// 阻塞运行
	m.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		auth.USERKEY: userInfo.ID,
		"ctx":        ctx,
		"lab_uuid":   req.LabUUID,
	})
}
