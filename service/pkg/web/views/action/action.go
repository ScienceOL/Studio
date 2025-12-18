package action

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/core/schedule/edge"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	actionEngine "github.com/scienceol/studio/service/pkg/core/schedule/engine/action"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/utils"
)

const (
	ActionRunChannel notify.Action = "action-run"
)

type Handle struct {
	rClient    *r.Client
	wsClient   *melody.Melody
	boardEvent notify.MsgCenter
}

func NewActionHandle(ctx context.Context) *Handle {
	wsClient := melody.New()
	wsClient.Config.MaxMessageSize = constant.MaxMessageSize

	h := &Handle{
		rClient:    redis.GetClient(),
		wsClient:   wsClient,
		boardEvent: events.NewEvents(),
	}

	// 注册通知处理
	if err := h.boardEvent.Registry(ctx, ActionRunChannel, h.HandleNotify); err != nil {
		logger.Errorf(ctx, "register action notify failed: %+v", err)
	}

	h.initActionWebSocket()
	return h
}

// @Summary 		执行设备动作
// @Description 	手动触发设备执行指定动作
// @Tags 			Action
// @Accept 			json
// @Produce 		json
// @Security 		BearerAuth
// @Param 			action body action.RunActionReq true "动作执行请求"
// @Success 		200 {object} common.Resp{data=action.RunActionResp} "执行成功"
// @Failure 		200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 			/v1/lab/action/run [post]
func (h *Handle) RunAction(ctx *gin.Context) {
	req := &actionEngine.RunActionReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse RunAction param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	// 验证必填参数
	if req.LabUUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("lab_uuid is required"))
		return
	}
	if req.DeviceID == "" {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("device_id is required"))
		return
	}
	if req.Action == "" {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("action is required"))
		return
	}
	if req.ActionType == "" {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("action_type is required"))
		return
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		common.ReplyErr(ctx, code.UnLogin)
		return
	}

	// 打印当前时间
	now := time.Now()
	logger.Infof(ctx, "RunAction request received at: %s", now.Format(time.RFC3339))

	// 生成任务 UUID
	if req.UUID.IsNil() {
		req.UUID = uuid.NewV4()
	}

	if exists, err := h.rClient.Exists(ctx, utils.LabHeartName(req.LabUUID)).Result(); err != nil || exists == 0 {
		common.ReplyErr(ctx, code.EdgeNotStartedErr)
		return
	}

	data := edge.ApiControlData[engine.WorkflowInfo]{
		ApiControlMsg: edge.ApiControlMsg{
			Action: edge.StartAction,
		},
		Data: engine.WorkflowInfo{
			TaskUUID:     req.UUID,
			WorkflowUUID: req.UUID,
			LabUUID:      req.LabUUID,
			UserID:       userInfo.ID,
		},
	}

	// 将请求数据存储到 Redis
	paramKey := actionEngine.ActionKey(req.UUID)
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Errorf(ctx, "marshal RunActionReq err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr.WithErr(err))
		return
	}

	ret := h.rClient.SetEx(ctx, paramKey, reqData, 24*time.Hour)
	if ret.Err() != nil {
		logger.Errorf(ctx, "set action param to redis err: %+v", ret.Err())
		common.ReplyErr(ctx, code.RPCHttpErr.WithErr(ret.Err()))
		return
	}

	// 发送任务到队列
	jobData, _ := json.Marshal(data)
	pushRet := h.rClient.LPush(ctx, utils.LabControlName(req.LabUUID), jobData)
	if pushRet.Err() != nil {
		logger.Errorf(ctx, "push job to queue err: %+v", pushRet.Err())
		common.ReplyErr(ctx, code.RPCHttpErr.WithErr(pushRet.Err()))
		return
	}

	logger.Infof(ctx, "action task created successfully, uuid: %s", req.UUID.String())

	// 返回任务 UUID
	common.ReplyOk(ctx, map[string]any{
		"task_uuid": req.UUID,
		"message":   "Action task created successfully",
	})
}

// @Summary 		查询动作执行结果
// @Description 	根据任务 UUID 查询动作执行结果
// @Tags 			Action
// @Accept 			json
// @Produce 		json
// @Security 		BearerAuth
// @Param 			uuid path string true "任务UUID"
// @Success 		200 {object} common.Resp{data=action.RunActionResp} "查询成功"
// @Failure 		200 {object} common.Resp{code=code.ErrCode} "查询失败"
// @Router 			/v1/lab/action/result/{uuid} [get]
func (h *Handle) GetActionResult(ctx *gin.Context) {
	uuidStr := ctx.Param("uuid")
	taskUUID, err := uuid.FromString(uuidStr)
	if err != nil {
		logger.Errorf(ctx, "parse uuid err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr.WithMsg("invalid uuid"))
		return
	}

	// 从 Redis 获取结果
	retKey := actionEngine.ActionRetKey(taskUUID)
	result := h.rClient.Get(ctx, retKey)
	if result.Err() != nil {
		if result.Err() == r.Nil {
			common.ReplyErr(ctx, code.RecordNotFound.WithMsg("result not found, task may still be running"))
		} else {
			logger.Errorf(ctx, "get action result from redis err: %+v", result.Err())
			common.ReplyErr(ctx, code.RPCHttpErr.WithErr(result.Err()))
		}
		return
	}

	// 解析结果
	resp := &actionEngine.RunActionResp{}
	if err := json.Unmarshal([]byte(result.Val()), resp); err != nil {
		logger.Errorf(ctx, "unmarshal action result err: %+v", err)
		common.ReplyErr(ctx, code.RPCHttpErr.WithErr(err))
		return
	}

	common.ReplyOk(ctx, resp)
}

// HandleNotify 处理通知消息并通过 WebSocket 广播给前端
func (h *Handle) HandleNotify(ctx context.Context, msg string) error {
	notifyData := &notify.SendMsg{}
	if err := json.Unmarshal([]byte(msg), notifyData); err != nil {
		logger.Errorf(ctx, "HandleNotify unmarshal data err: %+v", err)
		return err
	}

	d := &common.Resp{
		Code: code.Success,
		Data: &common.WSData[any]{
			WsMsgType: common.WsMsgType{
				Action:  "action_status_update",
				MsgUUID: notifyData.UUID,
			},
			Data: notifyData.Data,
		},
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(d)
	// 广播给所有订阅了该任务的客户端
	return h.wsClient.BroadcastFilter(data, func(s *melody.Session) bool {
		sessionValue, ok := s.Get("task_uuid")
		if !ok {
			return false
		}

		if sessionValue.(uuid.UUID) == notifyData.TaskUUID {
			return true
		}

		return false
	})
}

// initActionWebSocket 初始化 WebSocket 处理器
func (h *Handle) initActionWebSocket() {
	h.wsClient.HandlePong(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "==================== action ws pong =====================")
		}
	})

	h.wsClient.HandleClose(func(s *melody.Session, _ int, _ string) error {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "action client close keys: %+v", s.Keys)
		}
		return nil
	})

	h.wsClient.HandleDisconnect(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "action client disconnected keys: %+v", s.Keys)
		}
	})

	h.wsClient.HandleError(func(s *melody.Session, err error) {
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
			logger.Errorf(ctx.(context.Context), "action websocket error keys: %+v, err: %+v", s.Keys, err)
		}
	})

	h.wsClient.HandleConnect(func(s *melody.Session) {
		c, _ := s.Get("ctx")
		ctx := c.(*gin.Context)
		logger.Infof(ctx, "action client connected keys: %+v", s.Keys)
	})
}

// @Summary 		Action WebSocket 连接
// @Description 	连接到指定任务的 WebSocket 会话，用于接收实时动作执行状态更新
// @Tags 			Action
// @Accept 			json
// @Produce 		json
// @Security 		BearerAuth
// @Param 			task_uuid path string true "任务UUID"
// @Success 		200 {object} common.Resp{} "连接成功（协议升级）"
// @Failure 		200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 			/v1/ws/action/{task_uuid} [get]
func (h *Handle) ActionWebSocket(ctx *gin.Context) {
	taskUUIDStr := ctx.Param("task_uuid")
	taskUUID, err := uuid.FromString(taskUUIDStr)
	if err != nil {
		logger.Errorf(ctx, "parse task_uuid err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr.WithMsg("invalid task_uuid"))
		return
	}

	// 阻塞运行 WebSocket 连接
	if err := h.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		"task_uuid": taskUUID,
		"ctx":       ctx,
	}); err != nil {
		logger.Errorf(ctx, "action HandleRequestWithKeys err: %+v", err)
	}
}
