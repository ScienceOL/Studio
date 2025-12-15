package edge

import (
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/repo"
)

type LabInfo struct {
	UUID      uuid.UUID
	ID        int64
	LabUserID string
	// Name    string
	Session *melody.Session
	Sandbox repo.Sandbox // 脚本运行沙箱
}

type ApiAction string // api 服务和 schedule 交互消息, 通过 redis 发送

const (
	StartWorkflow ApiAction = "start_job"      // 启动工作流
	StartNotebook ApiAction = "start_notebook" // 启动实验记录本
)

type ApiMsg struct {
	Action ApiAction `json:"action"`
}

type ApiData[T any] struct {
	ApiMsg
	Data T `json:"data"`
}

type ApiControlAction string // api  服务和  schedule  交互，控制类消息

const (
	StartAction    ApiControlAction = "start_action"    // 启动单个action，属于快速队列消息
	StopJob        ApiControlAction = "stop_job"        // 停止任务
	StatusJob      ApiControlAction = "status_job"      // 任务状态
	AddMaterial    ApiControlAction = "add_material"    // 增加物料
	UpdateMaterial ApiControlAction = "update_material" // 更新物料
	RemoveMaterial ApiControlAction = "remove_material" // 移除物料
)

type ApiControlMsg struct {
	Action ApiControlAction `json:"action"`
}

type ApiControlData[T any] struct {
	ApiControlMsg
	Data T `json:"data"`
}

type StopJobReq struct {
	UUID   uuid.UUID `json:"uuid"`
	UserID string    `json:"user_id"`
}

type EdgeAction string // edge 交互消息, 通过 websocket 交互

const (
	// 服务端下发
	JobStart          EdgeAction = "job_start"          // 下发动作
	QueryActionStatus EdgeAction = "query_action_state" // 查询动作是否能执行
	Pong              EdgeAction = "pong"               // 心跳
	CancelTask        EdgeAction = "cancel_task"        // 取消任务

	// edge 上行数据
	JobStatus         EdgeAction = "job_status"          // 任务状态回调
	DeviceStatus      EdgeAction = "device_status"       // 设备状态
	Ping              EdgeAction = "ping"                // 心跳
	ReportActionState EdgeAction = "report_action_state" // 上报 action status
	HostNodeReady     EdgeAction = "host_node_ready"     // edge 初始化完成
	NormalExist       EdgeAction = "normal_exit"         // edge 正常退出
)

type EdgeMsg struct {
	Action EdgeAction `json:"action"`
}

type EdgeData[T any] struct {
	EdgeMsg
	Data T `json:"data"`
}

type ActionPong struct {
	PingID          string  `json:"ping_id"`
	ClientTimestamp float64 `json:"client_timestamp"`
	ServerTimestamp float64 `json:"server_timestamp"`
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

type ActionStatus struct {
	engine.ActionKey
	engine.ActionValue
}

type EdgeReady struct {
	Status    string  `json:"status"`
	Timestamp float64 `json:"timestamp"`
}
