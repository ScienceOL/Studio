package engine

import "context"

/*
	调度引擎模块，抽象调度接口
*/

type Task interface {
	Run(ctx context.Context, job *WorkflowInfo) error
	Stop(ctx context.Context) error
	GetStatus(ctx context.Context) error
}
