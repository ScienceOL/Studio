package engine

import (
	"context"

	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/schedule/device"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
)

type TaskParam struct {
	Devices device.Service
	Session *melody.Session
	Cancle  context.CancelFunc
}

type WorkflowTaskKey struct {
	TaskUUID     uuid.UUID `json:"task_uuid"`
	WorkflowUUID uuid.UUID `json:"workflow_id"` // 任务 id
}

type WorkflowInfo struct {
	Action       WorkflowAction    `json:"action"`
	TaskUUID     uuid.UUID         `json:"task_uuid"`
	WorkflowUUID uuid.UUID         `json:"workflow_id"` // 任务 id
	LabUUID      uuid.UUID         `json:"lab_uuid"`
	UserID       string            `json:"user_id"` // 提交用户 id
	LabData      *model.Laboratory `json:"-"`
}

type WorkflowAction string

const (
	StartJob  WorkflowAction = "start_job"
	StopJob   WorkflowAction = "stop_job"
	StatusJob WorkflowAction = "status_job"
)

type ServerInfo struct {
	SendTimestamp float64 `json:"send_timestamp"`
}

type SendActionData struct {
	DeviceID   string         `json:"device_id"`
	Action     string         `json:"action"`
	ActionType string         `json:"action_type"`
	ActionArgs datatypes.JSON `json:"action_args"`
	JobID      string         `json:"job_id"`
	NodeID     string         `json:"node_id"`
	ServerInfo ServerInfo     `json:"server_info"`
}

type BoardMsg struct {
	Header         string    `json:"header"`
	NodeUUID       uuid.UUID `json:"node_uuid"`
	WorkflowStatus string    `json:"workflow_status"`
	Status         string    `json:"status"`
	Type           string    `json:"type"`
	Msg            []string  `json:"msg"`
	StackTrace     []string  `json:"stack_trace"`
	ReturnInfos    []string  `json:"return_infos"`
}
