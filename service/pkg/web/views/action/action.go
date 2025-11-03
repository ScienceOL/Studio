package action

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine/action"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
)

type Handle struct {
	rClient *r.Client
}

func NewActionHandle(ctx context.Context) *Handle {
	return &Handle{
		rClient: redis.GetClient(),
	}
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
	req := &action.RunActionReq{}
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

	// 生成任务 UUID
	if req.UUID.IsNil() {
		req.UUID = uuid.NewV4()
	}

	// 将请求数据存储到 Redis
	paramKey := action.ActionKey(req.UUID)
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Errorf(ctx, "marshal RunActionReq err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr.WithErr(err))
		return
	}

	ret := h.rClient.SetEx(ctx, paramKey, reqData, 1*time.Hour)
	if ret.Err() != nil {
		logger.Errorf(ctx, "set action param to redis err: %+v", ret.Err())
		common.ReplyErr(ctx, code.RPCHttpErr.WithErr(ret.Err()))
		return
	}

	// 发送任务到队列
	conf := config.Global().Job
	jobInfo := engine.WorkflowInfo{
		Action:   engine.StartAction,
		TaskUUID: req.UUID,
		LabUUID:  req.LabUUID,
		UserID:   "manual", // 手动触发
	}

	jobData, _ := json.Marshal(jobInfo)
	pushRet := h.rClient.LPush(ctx, conf.JobQueueName, jobData)
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
	retKey := action.ActionRetKey(taskUUID)
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
	resp := &action.RunActionResp{}
	if err := json.Unmarshal([]byte(result.Val()), resp); err != nil {
		logger.Errorf(ctx, "unmarshal action result err: %+v", err)
		common.ReplyErr(ctx, code.RPCHttpErr.WithErr(err))
		return
	}

	common.ReplyOk(ctx, resp)
}
