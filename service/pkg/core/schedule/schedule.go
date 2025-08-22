package schedule

import (
	"context"
)

type Control interface {
	Connect(ctx context.Context)
	OnJobMessage(ctx context.Context, msg []byte)
	OnEdgeMessge(ctx context.Context, msg string)
}
