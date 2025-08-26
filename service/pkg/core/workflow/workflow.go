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
	ForkTemplate(ctx context.Context)
	NodeTemplateDetail(ctx context.Context, templateUUID uuid.UUID) (*NodeTemplateDetailResp, error)
	TemplateDetail(ctx context.Context)
	TemplateList(ctx context.Context, req *TplPageReq) (*common.PageResp[[]*TemplateListResp], error)
	TemplateTags(ctx context.Context, req *TemplateTagsReq) ([]string, error)
	UpdateNodeTemplate(ctx context.Context)
	GetWorkflowList(ctx context.Context, req *WorkflowListReq) (*WorkflowListResult, error)
	GetWorkflowDetail(ctx context.Context, workflowUUID uuid.UUID) (*WorkflowDetailResp, error)
	OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error
	OnWSConnect(ctx context.Context, s *melody.Session) error
	WorkflowTaskList(ctx context.Context, req *WorkflowTaskReq) (*common.PageMoreResp[[]*WorkflowTaskResp], error)
	TaskDownload(ctx context.Context, req *WorkflowTaskDownloadReq) (*bytes.Buffer, error)
}
