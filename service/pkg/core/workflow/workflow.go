package workflow

import (
	"bytes"
	"context"

	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/uuid"
)

type Service interface {
	Create(ctx context.Context, data *WorkflowReq) (*WorkflowResp, error)
	NodeTemplateList(ctx context.Context, req *TplPageReq) (*common.PageResp[[]*TemplateNodeResp], error)
	NodeTemplateDetail(ctx context.Context, templateUUID uuid.UUID) (*NodeTemplateDetailResp, error)
	TemplateDetail(ctx context.Context)
	TemplateList(ctx context.Context, req *TplPageReq) (*common.PageResp[[]*TemplateListResp], error)
	TemplateTags(ctx context.Context, req *TemplateTagsReq) ([]string, error)
	UpdateNodeTemplate(ctx context.Context)
	GetWorkflowList(ctx context.Context, req *WorkflowListReq) (*WorkflowListResult, error)
	GetWorkflowDetail(ctx context.Context, req *DetailReq) (*WorkflowDetailResp, error)
	OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error
	OnWSConnect(ctx context.Context, s *melody.Session) error
	WorkflowTaskList(ctx context.Context, req *WorkflowTaskReq) (*common.PageMoreResp[[]*WorkflowTaskResp], error)
	TaskDownload(ctx context.Context, req *WorkflowTaskDownloadReq) (*bytes.Buffer, error)
	UpdateWorkflow(ctx context.Context, req *UpdateReq) error
	DelWorkflow(ctx context.Context, req *DelReq) error
	WorkflowTemplateList(ctx context.Context, req *WorkflowTemplateListReq) (*common.PageResp[[]*WorkflowTemplateListRes], error)
	WorkflowTemplateTags(ctx context.Context) ([]string, error)
	ForkWrokflow(ctx context.Context, req *ForkReq) error
}
