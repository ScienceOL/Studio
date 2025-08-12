package workflow

import (
	"context"

	"github.com/olahol/melody"
)

type workflowImpl struct{}

func (workflowimpl *workflowImpl) Add(ctx context.Context) {
}

func (workflowimpl *workflowImpl) NodeTemplateList(ctx context.Context) {
}

func (workflowimpl *workflowImpl) ForkTemplate(ctx context.Context) {
}

func (workflowimpl *workflowImpl) NodeTemplateDetail(ctx context.Context) {
}

func (workflowimpl *workflowImpl) TemplateDetail(ctx context.Context) {
}

func (workflowimpl *workflowImpl) TemplateList(ctx context.Context) {
}

func (workflowimpl *workflowImpl) UpdateNodeTemplate(ctx context.Context) {
}

func (workflowimpl *workflowImpl) OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error {
	return nil
}

func (workflowimpl *workflowImpl) OnWSConnect(ctx context.Context, s *melody.Session) error {
	return nil
}
