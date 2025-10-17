package schedule

import (
	"context"

	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/model"
)

const (
	LABINFO = "LAB_INFO"
)

type ActionType string

const (
	JobStart     ActionType = "job_start"     // 下发动作
	JobStatus    ActionType = "job_status"    // 任务状态回调
	DeviceStatus ActionType = "device_status" // 设备状态
)

type LabInfo struct {
	LabUser *model.UserData
	LabData *model.Laboratory
}

type ControlTask struct {
	Task   engine.Task
	Cancle context.CancelFunc
	ctx    context.Context
}
