package notify

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/uuid"
)

/*
	该模块功能主要是把个 pod 之间消息分布式广播通知，消息可以横向水平拓展
要求如下，
1. 方便的链接websocket 和 redis，识别消息、分发消息
2. 单例模式，一个进程内只有一个示例
*/

type Action string

const (
	MaterialModify Action = "material-modify"
	WorkflowRun    Action = "workflow-run"
)

type SendMsg struct {
	Channel      Action    `json:"action"`
	LabUUID      uuid.UUID `json:"lab_uuid"`
	WorkflowUUID uuid.UUID `json:"work_flow_uud"`
	TaskUUID     uuid.UUID `json:"task_uuid"`
	UserID       string    `json:"user_id"`
	Data         any       `json:"data"`
	UUID         uuid.UUID `json:"uuid"`
	Timestamp    int64     `json:"timestamp"`
}

type HandleFunc func(ctx context.Context, msg string) error

type MsgCenter interface {
	Registry(ctx context.Context, msgName Action, handleFunc HandleFunc) error
	Broadcast(ctx context.Context, msg *SendMsg) error
	Close(ctx context.Context) error
}
