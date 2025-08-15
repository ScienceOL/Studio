package workflow

import (
	"context"

	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
)

type Service interface {
	Create(ctx context.Context, data *WorkflowReq) (*WorkflowResp, error)
	NodeTemplateList(ctx context.Context, req *TplPageReq) (*common.PageResp[[]*TemplateNodeResp], error)
	ForkTemplate(ctx context.Context)
	NodeTemplateDetail(ctx context.Context)
	TemplateDetail(ctx context.Context)
	TemplateList(ctx context.Context)
	UpdateNodeTemplate(ctx context.Context)
	OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error
	OnWSConnect(ctx context.Context, s *melody.Session) error
}
