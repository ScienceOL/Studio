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
	Action  *model.WorkflowNodeTemplate
	Handles []*model.WorkflowHandleTemplate
}

type WorkflowGrpah struct {
	Workflow *model.Workflow
	Nodes    []*WorkflowNodeInfo
	Edges    []*model.WorkflowEdge
}

type DeleteWorkflow struct {
	NodeUUIDs  []uuid.UUID
	EdgesUUIDs []uuid.UUID
}

type TaskReq struct {
	UserID     string
	LabID      int64
	WrokflowID int64
}

type QueryTemplage struct {
	Name            string
	ResourceNodeIDs []int64
	LabID           int64
}

type WorkflowRepo interface {
	FindDatas(ctx context.Context, datas any, condition map[string]any, keys ...string) error
	UpdateData(ctx context.Context, data any, condition map[string]any, keys ...string) error
	Create(ctx context.Context, data *model.Workflow) error
	CreateNode(ctx context.Context, data *model.WorkflowNode) error
	GetWorkflowByUUID(ctx context.Context, uuid uuid.UUID) (*model.Workflow, error)
	GetWorkflowGraph(ctx context.Context, userID string, uuid uuid.UUID) (*WorkflowGrpah, error)
	GetWorkflowNodeTemplate(ctx context.Context, condition map[string]any) ([]*model.WorkflowNodeTemplate, error)
	GetWorkflowHandleTemaplates(ctx context.Context, actionIDs []int64) ([]*model.WorkflowHandleTemplate, error)
	GetWorkflowNodes(ctx context.Context, condition map[string]any) ([]*model.WorkflowNode, error)
	GetWorkflowEdges(ctx context.Context, nodeUUIDs []uuid.UUID) ([]*model.WorkflowEdge, error)
	UpdateWorkflowNode(ctx context.Context, nodeUUID uuid.UUID, data *model.WorkflowNode, updateColumns []string) error
	UpdateWorkflowNodes(ctx context.Context, nodeUUIDs []uuid.UUID, data *model.WorkflowNode, updateColumns []string) error
	DeleteWorkflowNodes(ctx context.Context, workflowUUIDs []uuid.UUID) (*DeleteWorkflow, error)
	DeleteWorkflowEdges(ctx context.Context, edgeUUIDs []uuid.UUID) ([]uuid.UUID, error)
	Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error)
	UpsertWorkflowEdge(ctx context.Context, datas []*model.WorkflowEdge) error
	UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids ...uuid.UUID) map[uuid.UUID]int64
	ID2UUID(ctx context.Context, tableModel schema.Tabler, ids ...int64) map[int64]uuid.UUID
	GetWorkflowList(ctx context.Context, userID string, labID int64, page *common.PageReq) ([]*model.Workflow, int64, error)
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
	UpsertNodes(ctx context.Context, nodes []*model.WorkflowNode) error
	UpsertEdge(ctx context.Context, edges []*model.WorkflowEdge) error
	DuplicateEdge(ctx context.Context, edges []*model.WorkflowEdge) error
	CreateJobs(ctx context.Context, datas []*model.WorkflowNodeJob) error
	UpsertJobs(ctx context.Context, datas []*model.WorkflowNodeJob) error
	GetTemplateList(ctx context.Context, req *common.PageReqT[*QueryTemplage]) (*common.PageResp[[]*model.WorkflowNodeTemplate], error)
	GetNodeTemplateByUUID(ctx context.Context, templateUUID uuid.UUID) (*model.WorkflowNodeTemplate, error)
	CreateWorkflowTask(ctx context.Context, data *model.WorkflowTask) error
	GetWorkflowTasks(ctx context.Context, req *common.PageReqT[*TaskReq]) (*common.PageMoreResp[[]*model.WorkflowTask], error)
}
