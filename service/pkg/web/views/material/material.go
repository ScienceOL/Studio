package material

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/material"
	impl "github.com/scienceol/studio/service/pkg/core/material/material"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type Handle struct {
	mService material.Service
	wsClient *melody.Melody
}

func NewMaterialHandle(ctx context.Context) *Handle {
	wsClient := melody.New()
	wsClient.Config.MaxMessageSize = constant.MaxMessageSize
	mService := impl.NewMaterial(ctx, wsClient)
	// 注册集群通知

	h := &Handle{
		mService: mService,
		wsClient: wsClient,
	}

	h.initMaterialWebSocket()
	return h
}

// @Summary      创建实验室物料
// @Description  创建实验室物料节点和连线
// @Tags         Material
// @Accept       json
// @Produce      json
// @Param        req  body      material.GraphNodeReq  true  "物料创建请求"
// @Success      200  {object}  common.Resp
// @Failure      400  {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/material [post]
func (m *Handle) CreateLabMaterial(ctx *gin.Context) {
	req := &material.GraphNodeReq{}
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

// @Summary      创建物料连线
// @Description  创建物料节点之间的连接边
// @Tags         Material
// @Accept       json
// @Produce      json
// @Param        req  body      material.GraphEdge  true  "物料连线创建请求"
// @Success      200  {object}  common.Resp
// @Failure      400  {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/material/edge [post]
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

// @Summary      下载物料数据
// @Description  下载实验室物料配置图数据，返回JSON文件
// @Tags         Material
// @Accept       json
// @Produce      application/json
// @Param        lab_uuid  path      string  true  "实验室UUID"
// @Success      200       {file}    string  "物料图数据JSON文件"
// @Failure      400       {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/material/download/{lab_uuid} [get]
func (m *Handle) DownloadMaterial(ctx *gin.Context) {
	var err error
	req := &material.DownloadMaterial{}
	if err := ctx.ShouldBindUri(req); err != nil {
		logger.Errorf(ctx, "parse DownloadMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := m.mService.DownloadMaterial(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "DownloadMaterial err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	commonResp := &common.Resp{
		Code: code.Success,
		Data: resp,
	}

	data, err := json.Marshal(commonResp)
	if err != nil {
		logger.Errorf(ctx, "DownloadMaterial err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Content-Disposition", "attachment; filename=material_graph.json")
	ctx.Header("Content-Type", "application/json")
	ctx.Header("Pragma", "public")
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(data)))

	// 创建Reader并发送数据
	reader := bytes.NewReader(data)
	ctx.DataFromReader(http.StatusOK, int64(len(data)), "application/json", reader, nil)
}

func (m *Handle) initMaterialWebSocket() {
	m.wsClient.HandleClose(func(s *melody.Session, _ int, _ string) error {
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
			if err := m.mService.OnWSConnect(ctx.(context.Context), s); err != nil {
				logger.Errorf(ctx.(context.Context), "material OnWSMsg err: %+v", err)
			}
		}
	})

	m.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		ctxI, ok := s.Get("ctx")
		if !ok {
			if err := s.CloseWithMsg([]byte("no ctx")); err != nil {
				logger.Errorf(context.Background(), "HandleMessage ctx not exist CloseWithMsg err: %+v", err)
			}
			return
		}

		if err := m.mService.OnWSMsg(ctxI.(*gin.Context), s, b); err != nil {
			logger.Errorf(ctxI.(*gin.Context), "material handle msg err: %+v", err)
		}
	})

	m.wsClient.HandleSentMessage(func(_ *melody.Session, _ []byte) {
		// 发送完消息后的回调
	})

	m.wsClient.HandleSentMessageBinary(func(_ *melody.Session, _ []byte) {
		// 发送完二进制消息后的回调
		// 如果发送的是字符串消息，上面的 HandleSentMessage 也会被回调
	})
}

// @Summary      物料WebSocket连接
// @Description  建立实验室物料实时通信WebSocket连接，用于实时同步物料状态和操作
// @Tags         Material
// @Accept       json
// @Produce      json
// @Param        lab_uuid  path      string  true  "实验室UUID"
// @Success      101       {string}  string  "WebSocket连接升级成功"
// @Failure      400       {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/ws/material/{lab_uuid} [get]
func (m *Handle) LabMaterial(ctx *gin.Context) {
	req := &material.LabWS{}
	var err error
	labUUIDStr := ctx.Param("lab_uuid")
	req.LabUUID, err = uuid.FromString(labUUIDStr)
	if err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	userInfo := auth.GetCurrentUser(ctx)

	// 阻塞运行
	if err := m.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		auth.USERKEY: userInfo,
		"ctx":        ctx,
		"lab_uuid":   req.LabUUID,
	}); err != nil {
		logger.Errorf(ctx, "LabMaterial HandleRequestWithKeys err: %+v", err)
	}
}
