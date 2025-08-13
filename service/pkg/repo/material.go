package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type NodeInfo struct {
	NodeUUID   uuid.UUID `json:"node_uuid"`
	HandleUUID uuid.UUID `json:"handle_uuid"`
}

type DelNodeInfo struct {
	NodeUUIDs []uuid.UUID `json:"node_uuids"`
	EdgeUUIDs []uuid.UUID `json:"edge_uuids"`
}

type UpdateNode struct {
	UpdateKeys []string
	Data       *model.MaterialNode
}

type MaterialRepo interface {
	TranslateIDOrUUID(ctx context.Context, data any) error
	// 更新或者插入物料
	UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode) error
	// 更新或插入 edge
	UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error
	// 获取所有的 node handle
	GetNodeHandles(ctx context.Context, labID int64, nodeNames []string, handleNames []string) (map[string]map[string]NodeInfo, error)
	// 根据 uuid 获取到所有 node 和 handle
	GetNodeHandlesByUUID(ctx context.Context, nodeUUIDs []uuid.UUID) (map[uuid.UUID]map[uuid.UUID]NodeInfo, error)
	// 删除 nodes，同时会删除对应的 handle 和 edge
	DelNodes(ctx context.Context, nodeUUIDs []uuid.UUID) (*DelNodeInfo, error)
	// 获取所有物料根据 lab id
	GetNodesByLabID(ctx context.Context, labID int64, selectKeys ...string) ([]*model.MaterialNode, error)
	// 根据所有的 uuid 获取所有的edges
	GetEdgesByNodeUUID(ctx context.Context, uuids []uuid.UUID, selectKeys ...string) ([]*model.MaterialEdge, error)
	// 批量 edges
	DelEdges(ctx context.Context, uuids []uuid.UUID) error
	// 批量跟新 node 数据
	UpdateNodeByUUID(ctx context.Context, data *model.MaterialNode, selectKeys ...string) error
	// 根据 uuid 获取节点 ID
	GetNodeIDByUUID(ctx context.Context, nodeUUID uuid.UUID) (int64, error)
	// 批量插入 workflowTpl
	UpsertWorkflowNodeTemplate(ctx context.Context, datas []*model.WorkflowNodeTemplate) error
	// 批量插入 workflowHandleTpl
	UpsertWorkflowHandleTemplate(ctx context.Context, datas []*model.WorkflowHandleTemplate) error
}
