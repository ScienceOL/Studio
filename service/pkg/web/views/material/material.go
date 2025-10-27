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

// @Summary 创建物料图
// @Description 在实验室中创建物料图（节点与边）
// @Tags Material
// @Accept json
// @Produce json
// @Param material body material.GraphNodeReq true "物料图创建请求"
// @Success 200 {object} common.Resp{} "创建成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material [post]
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

// @Summary 边缘端创建物料
// @Description 从边缘端创建物料（带 UUID 的节点）
// @Tags Material
// @Accept json
// @Produce json
// @Param material body material.CreateMaterialReq true "边缘端创建物料请求"
// @Success 200 {object} common.Resp{data=material.CreateMaterialResp} "创建成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/edge/create [post]
func (m *Handle) EdgeCreateMaterial(ctx *gin.Context) {
	req := &material.CreateMaterialReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse EdgeCreateMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	resp, err := m.mService.EdgeCreateMaterial(ctx, req)

	common.Reply(ctx, err, resp)
}

// @Summary 边缘端更新/插入物料
// @Description 边缘端批量更新或插入物料
// @Tags Material
// @Accept json
// @Produce json
// @Param material body material.UpsertMaterialReq true "边缘端更新/插入物料请求"
// @Success 200 {object} common.Resp{data=material.UpsertMaterialResp} "操作成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/edge/upsert [post]
func (m *Handle) EdgeUpsertMaterial(ctx *gin.Context) {
	req := &material.UpsertMaterialReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse EdgeCreateMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	resp, err := m.mService.EdgeUpsertMaterial(ctx, req)

	common.Reply(ctx, err, resp)
}

// @Summary 边缘端创建物料连线
// @Description 从边缘端创建物料之间的连线
// @Tags Material
// @Accept json
// @Produce json
// @Param edges body material.CreateMaterialEdgeReq true "创建连线请求"
// @Success 200 {object} common.Resp{} "创建成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/edge/connection [post]
func (m *Handle) EdgeCreateEdge(ctx *gin.Context) {
	req := &material.CreateMaterialEdgeReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse EdgeCreateMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	err := m.mService.EdgeCreateEdge(ctx, req)

	common.Reply(ctx, err)
}

// @Summary 保存物料图
// @Description 保存当前实验室的物料图
// @Tags Material
// @Accept json
// @Produce json
// @Param material body material.SaveGrapReq true "保存物料图请求"
// @Success 200 {object} common.Resp{} "保存成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/save [post]
func (m *Handle) SaveMaterial(ctx *gin.Context) {
	req := &material.SaveGrapReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse CreateLabMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	if err := m.mService.SaveMaterial(ctx, req); err != nil {
		logger.Errorf(ctx, "CreateMaterial err: %+v", err)
		common.ReplyErr(ctx, err)

		return
	}

	common.ReplyOk(ctx)
}

// @Summary 查询物料
// @Description 查询实验室物料
// @Tags Material
// @Accept json
// @Produce json
// @Param query query material.MaterialReq false "查询参数"
// @Success 200 {object} common.Resp{data=material.MaterialResp} "查询成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/query [get]
func (m *Handle) QueryMaterial(ctx *gin.Context) {
	req := &material.MaterialReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "parse LabMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := m.mService.LabMaterial(ctx, req)
	common.Reply(ctx, err, resp)
}

// @Summary 按 UUID 查询物料
// @Description 通过 UUID 列表查询物料信息
// @Tags Material
// @Accept json
// @Produce json
// @Param body body material.MaterialQueryReq true "物料 UUID 列表"
// @Success 200 {object} common.Resp{data=material.MaterialQueryResp} "查询成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/query-by-uuid [post]
func (m *Handle) QueryMaterialByUUID(ctx *gin.Context) {
	req := &material.MaterialQueryReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse QueryMaterialByUUID param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := m.mService.EdgeQueryMaterial(ctx, req)
	common.Reply(ctx, err, resp)
}

// @Summary 边缘端下载物料
// @Description 边缘端下载物料图
// @Tags Material
// @Produce json
// @Success 200 {object} common.Resp{data=material.DownloadMaterialResp} "下载成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/edge/download [get]
func (m *Handle) EdgeDownloadMaterial(ctx *gin.Context) {
	resp, err := m.mService.EdgeDownloadMaterial(ctx)
	common.Reply(ctx, err, resp)
}

func (m *Handle) BatchUpdateMaterial(ctx *gin.Context) {
	req := &material.UpdateMaterialReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse BatchUpdateMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	err := m.mService.BatchUpdateUniqueName(ctx, req)
	common.Reply(ctx, err)
}

func (m *Handle) ResourceList(ctx *gin.Context) {
	req := &material.ResourceReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "parse BatchUpdateMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := m.mService.ResourceList(ctx, req)
	common.Reply(ctx, err, resp)
}

// @Summary 设备可用动作
// @Description 获取设备对应的可用动作列表
// @Tags Material
// @Accept json
// @Produce json
// @Param lab_uuid query string true "实验室UUID"
// @Param name query string true "设备名称"
// @Success 200 {object} common.Resp{data=material.DeviceActionResp} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/actions [get]
func (m *Handle) Actions(ctx *gin.Context) {
	req := &material.DeviceActionReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "parse BatchUpdateMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := m.mService.DeviceAction(ctx, req)
	common.Reply(ctx, err, resp)
}

// @Summary 创建物料连线
// @Description 创建两个物料节点之间的连线
// @Tags Material
// @Accept json
// @Produce json
// @Param edges body material.GraphEdge true "物料连线请求"
// @Success 200 {object} common.Resp{} "创建成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/edge [post]
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
		common.ReplyErr(ctx, code.ParamErr.WithErr(err))
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

// @Summary 获取模板详情
// @Description 获取物料模板详情
// @Tags Material
// @Produce json
// @Param template_uuid path string true "模板UUID"
// @Success 200 {object} common.Resp{data=material.TemplateResp} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /api/lab/material/template/detail/{template_uuid} [get]
func (m *Handle) Template(ctx *gin.Context) {
	req := &material.TemplateReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		logger.Errorf(ctx, "MaterialTemplate err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr.WithErr(err))
		return
	}
	resp, err := m.mService.GetMaterialTemplate(ctx, req)
	common.Reply(ctx, err, resp)
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
