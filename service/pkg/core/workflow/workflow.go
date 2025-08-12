package workflow

import (
	"context"

	"github.com/olahol/melody"
)

type Service interface {
	Add(ctx context.Context, data *WorkflowReq) (*WorkflowResp, error)
	NodeTemplateList(ctx context.Context)
	ForkTemplate(ctx context.Context)
	NodeTemplateDetail(ctx context.Context)
	TemplateDetail(ctx context.Context)
	TemplateList(ctx context.Context)
	UpdateNodeTemplate(ctx context.Context)
	OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error
	OnWSConnect(ctx context.Context, s *melody.Session) error
}
