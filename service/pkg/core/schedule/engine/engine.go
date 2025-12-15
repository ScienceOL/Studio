package engine

import (
	"context"
	"time"

	"github.com/scienceol/studio/service/pkg/common/uuid"
)

/*
	调度引擎模块，抽象调度接口
*/

type Task interface {
	Run(ctx context.Context, job *WorkflowInfo) error
	Stop(ctx context.Context) error
	GetStatus(ctx context.Context) error
	OnJobUpdate(ctx context.Context, data *JobData) error
	ID(ctx context.Context) uuid.UUID // 获取当前任务 id

	// 状态控制
	GetDeviceActionStatus(ctx context.Context, key ActionKey) (ActionValue, bool)
	SetDeviceActionStatus(ctx context.Context, key ActionKey, free bool, needMore time.Duration)
	InitDeviceActionStatus(ctx context.Context, key ActionKey, start time.Time, free bool)
	DelStatus(ctx context.Context, key ActionKey)
}
