package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
	"gorm.io/gorm/schema"
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
type DeleteWorkflow struct {
	NodeUUIDs []uuid.UUID
	EdgeUUIDs []uuid.UUID
}

type WorkflowRepo interface {
	Create(ctx context.Context, data *model.Workflow) error
	CreateNode(ctx context.Context, data *model.WorkflowNode) error
	GetWorkflowByUUID(ctx context.Context, uuid uuid.UUID) (*model.Workflow, error)
	GetWorkflowGraph(ctx context.Context, userID string, uuid uuid.UUID) (*WorkflowGrpah, error)
	GetWorkflowTemplate(ctx context.Context, labID int64) ([]*WorkflowTemplate, error)
	GetWorkflowTemplateByUUID(ctx context.Context, tplUUID uuid.UUID) (*WorkflowTemplate, error)
	GetWorkflowNode(ctx context.Context, uuid uuid.UUID) (*model.WorkflowNode, error)
	UpdateWorkflowNode(ctx context.Context, workflowUUID uuid.UUID, data *model.WorkflowNode, updateColumns []string) error
	DeleteWorkflowNodes(ctx context.Context, workflowUUIDs []uuid.UUID) (*DeleteWorkflow, error)
	DeleteWorkflowEdges(ctx context.Context, edgeUUIDs []uuid.UUID) ([]uuid.UUID, error)
	Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error)
	UpsertWorkflowEdge(ctx context.Context, datas []*model.WorkflowEdge) error
	UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids []uuid.UUID) map[uuid.UUID]int64
	ID2UUID(ctx context.Context, tableModel schema.Tabler, ids []int64) map[int64]uuid.UUID
}
