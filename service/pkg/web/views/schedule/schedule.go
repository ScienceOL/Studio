package schedule

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/core/schedule"
)

type handle struct {
	ctrl schedule.Control
}

func New(ctx context.Context) *handle {
	return &handle{
		ctrl: schedule.NewControl(ctx),
	}
}

func (m *handle) Connect(ctx *gin.Context) {
	m.ctrl.Connect(ctx)
}
