package material

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/material"
	m "github.com/scienceol/studio/service/pkg/core/material"
	impl "github.com/scienceol/studio/service/pkg/core/material/material"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type materialHandle struct {
	mService m.MaterialService
}

func NewMaterialHandle() *materialHandle {
	return &materialHandle{
		mService: impl.NewMaterial(),
	}
}

func (m *materialHandle) CreateLabMaterial(ctx *gin.Context) {
	reqs := make([]*material.MaterialNode, 0, 1)
	if err := ctx.ShouldBindJSON(&reqs); err != nil {
		logger.Errorf(ctx, "parse CreateLabMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	if err := m.mService.CreateMaterial(ctx, reqs); err != nil {
		logger.Errorf(ctx, "CreateMaterial err: %+v", err)
		common.ReplyErr(ctx, err)

		return
	}

	common.ReplyOk(ctx)
}

func (m *materialHandle) LabMaterial(ctx *gin.Context) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 允许跨域连接 TODO: 生产环境严格限制
			return true
		},
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Errorf(ctx, "WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// WebSocket 连接处理逻辑
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logger.Errorf(ctx, "WebSocket read error: %v", err)
			break
		}

		// 处理接收到的消息
		logger.Infof(ctx, "Received: %s", message)

		// 回显消息，消息不是线程安全的，需要封装 channel
		// github.com/olahol/melody
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			logger.Errorf(ctx, "WebSocket write error: %v", err)
			break
		}
	}
}
