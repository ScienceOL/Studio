package schdule

import (
	"context"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/core/schedule"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

/*
1. edge 实验室 websocket 连接之后双向发送消息，运行指令，上报状态。
2. redis 接受消息。
	a. 工作流启动的队列消息，哪个服务抢到队列的消息，哪个服务运行工作流。
		1. 问题如何解决，两个 schedule pod，edge 连接到 a pod，结果 b 抢到了调度策略，如何把消息发送到 b 让 b 下发任务？
			解决方案，如果 a 抢到了，a 把命令广播出去，b 收到后，直接下发消息。b 收到状态回报消息收，b 修改数据库，广播通知消息，web 服务
			收到后直接发送给 web 侧。
	b. 工作流运行时接收到 websocket 消息上报之后，广播通知所有的客户端终端。
*/

type handle struct {
	service  schedule.Service
	wsClient *melody.Melody
}

func New() *handle {
	wsClient := melody.New()
	wsClient.Config.MaxMessageSize = constant.MaxMessageSize

	h := &handle{
		wsClient: wsClient,
	}

	h.initMaterialWebSocket()
	return h
}

func (m *handle) initMaterialWebSocket() {
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
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "websocket connect keys: %+v", s.Keys)
			// m.mService.OnWSConnect(ctx.(context.Context), s)
		}
	})

	m.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		ctxI, ok := s.Get("ctx")
		if !ok {
			s.CloseWithMsg([]byte("no ctx"))
			return
		}
		_ = ctxI

		// if err := m.mService.OnWSMsg(ctxI.(*gin.Context), s, b); err != nil {
		// 	logger.Errorf(ctxI.(*gin.Context), "material handle msg err: %+v", err)
		// }

	})

	m.wsClient.HandleSentMessage(func(s *melody.Session, b []byte) {
		// 发送完字符串消息后的回调
	})

	m.wsClient.HandleSentMessageBinary(func(s *melody.Session, b []byte) {
		// 发送完二进制消息后的回调
		// 如果发送的是字符串消息，上面的 HandleSentMessage 也会被回调
	})
}
