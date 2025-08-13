package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
)

type WorkflowNodeInfo struct {
	Node     *model.WorkflowNode
	Template *model.WorkflowNodeTemplate
	Schema   datatypes.JSON
	Handles  []*model.WorkflowHandleTemplate
}

type WorkflowGrpah struct {
	Workflow *model.Workflow
	Nodes    []*WorkflowNodeInfo
	Edges    []*model.WorkflowEdge
}

type WorkflowTemplate struct {
	Template *model.WorkflowNodeTemplate
	Handles  []*model.WorkflowHandleTemplate
}

type WorkflowRepo interface {
	Create(ctx context.Context, data *model.Workflow) error
	CreateNode(ctx context.Context, data *model.WorkflowNode) error
	GetWorkflowByUUID(ctx context.Context, uuid uuid.UUID) (*model.Workflow, error)
	IsExist(ctx context.Context, uuid uuid.UUID) (bool, error)
	GetWorkflowGraph(ctx context.Context, userID string, uuid uuid.UUID) (*WorkflowGrpah, error)
	GetWorkflowTemplate(ctx context.Context, labID int64) ([]*WorkflowTemplate, error)
	GetWorkflowTemplateByUUID(ctx context.Context, tplUUID uuid.UUID) (*WorkflowTemplate, error)
	GetWorkflowNode(ctx context.Context, uuid uuid.UUID) (*model.WorkflowNode, error)
}
