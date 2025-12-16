package edge

import (
	"context"

	"github.com/olahol/melody"
)

type Edge interface {
	// edge 侧发送消息
	OnEdgeMessge(ctx context.Context, s *melody.Session, b []byte)
	// job 运行工作流消息
	OnJobMessage(ctx context.Context, msg string)
	// 心跳消息
	OnPongMessage(ctx context.Context)
	// 处理关闭逻辑
	Close(ctx context.Context)
}
