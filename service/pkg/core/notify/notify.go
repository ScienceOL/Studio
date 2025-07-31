package notify

import "context"

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
	Action Action `json:"action"`
	Data   any    `json:"data"`
}

type MsgCenter interface {
	Registry(ctx context.Context, msgName Action, handleFunc func(ctx context.Context, msg string) error) error
	Broadcast(ctx context.Context, msg *SendMsg) error
}
