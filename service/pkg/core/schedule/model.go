package schedule

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/model"
)

const (
	LABINFO = "LAB_INFO"
)

type ActionType string

const (
	// 服务端下发
	JobStart          ActionType = "job_start"          // 下发动作
	QueryActionStatus ActionType = "query_action_state" // 查询动作是否能执行
	Pong              ActionType = "pong"               // 心跳
	CancelTask        ActionType = "cancel_task"        // 取消任务

	// edge 上行数据
	JobStatus         ActionType = "job_status"          // 任务状态回调
	DeviceStatus      ActionType = "device_status"       // 设备状态
	Ping              ActionType = "ping"                // 心跳
	ReportActionState ActionType = "report_action_state" // 上报 action status
)

type SendAction[T any] struct {
	Action ActionType `json:"action"`
	Data   T          `json:"data,omitempty"`
}

type LabInfo struct {
	LabUser *model.UserData
	LabData *model.Laboratory
}

type ControlTask struct {
	Task   engine.Task
	Cancle context.CancelFunc
	Ctx    context.Context
}

type DeviceValue struct {
	PropertyName string  `json:"property_name"`
	Status       any     `json:"status"`
	Timestamp    float32 `json:"timestamp"`
}

type DeviceData struct {
	DeviceID string      `json:"device_id"`
	Data     DeviceValue `json:"data"`
}

type UpdateMaterialData struct {
	UUID uuid.UUID   `json:"uuid"`
	Data DeviceValue `json:"data"`
}
