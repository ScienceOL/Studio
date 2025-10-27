package schedule

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/core/schedule"
	"github.com/scienceol/studio/service/pkg/core/schedule/control"
)

type Handle struct {
	ctrl schedule.Control
}

func New(ctx context.Context) *Handle {
	return &Handle{
		ctrl: control.NewControl(ctx),
	}
}

func (m *Handle) Connect(ctx *gin.Context) {
	m.ctrl.Connect(ctx)
}

func (m *Handle) Close(ctx context.Context) {
	m.ctrl.Close(ctx)
}
