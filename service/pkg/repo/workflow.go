package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/gorm/schema"
)

type WorkflowNodeInfo struct {
	Node    *model.WorkflowNode
	Action  *model.DeviceAction
	Handles []*model.ActionHandleTemplate
}

type WorkflowGrpah struct {
	Workflow *model.Workflow
	Nodes    []*WorkflowNodeInfo
	Edges    []*model.WorkflowEdge
}

type WorkflowTemplate struct {
	// Template *model.WorkflowNodeTemplate
	// Handles  []*model.WorkflowHandleTemplate
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
	GetDeviceAction(ctx context.Context, condition map[string]any) ([]*model.DeviceAction, error)
	GetDeviceActionHandles(ctx context.Context, actionIDs []int64) ([]*model.ActionHandleTemplate, error)
	GetWorkflowNode(ctx context.Context, condition map[string]any) ([]*model.WorkflowNode, error)
	UpdateWorkflowNode(ctx context.Context, nodeUUID uuid.UUID, data *model.WorkflowNode, updateColumns []string) error
	UpdateWorkflowNodes(ctx context.Context, nodeUUIDs []uuid.UUID, data *model.WorkflowNode, updateColumns []string) error
	DeleteWorkflowNodes(ctx context.Context, workflowUUIDs []uuid.UUID) (*DeleteWorkflow, error)
	DeleteWorkflowEdges(ctx context.Context, edgeUUIDs []uuid.UUID) ([]uuid.UUID, error)
	Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error)
	UpsertWorkflowEdge(ctx context.Context, datas []*model.WorkflowEdge) error
	UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids ...uuid.UUID) map[uuid.UUID]int64
	ID2UUID(ctx context.Context, tableModel schema.Tabler, ids ...int64) map[int64]uuid.UUID
	GetWorkflowTemplatePage(ctx context.Context, labID uuid.UUID, page *common.PageReq) (*common.PageResp[*WorkflowTemplate], error)
	GetWorkflowList(ctx context.Context, userID string, labID int64, page *common.PageReq) ([]*model.Workflow, int64, error)
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
}
