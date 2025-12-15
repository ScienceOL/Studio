package schedule

import (
	"context"
)

type Control interface {
	// 用户连接
	Connect(ctx context.Context)
	// job 运行工作流消息
	// OnJobMessage(ctx context.Context, msg []byte)
	// // edge 侧发送消息
	// OnEdgeMessge(ctx context.Context, s *melody.Session, b []byte)
	// 处理关闭逻辑
	Close(ctx context.Context)
}
