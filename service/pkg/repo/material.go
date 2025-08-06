package repo

import (
	"context"

	"github.com/gofrs/uuid/v5"
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

type MaterialRepo interface {
	// 更新或者插入物料
	UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode) error
	// 更新或插入 handle
	UpsertMaterialHandle(ctx context.Context, datas []*model.MaterialHandle) error
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
	// 获取所有物料id 获取所有的 Handles
	GetHandlesByNodeID(ctx context.Context, nodeIDs []int64, selectKeys ...string) ([]*model.MaterialHandle, error)
	// 根据所有的 uuid 获取所有的edges
	GetEdgesByNodeUUID(ctx context.Context, uuids []uuid.UUID, selectKeys ...string) ([]*model.MaterialEdge, error)
	// 批量 edges
	DelEdges(ctx context.Context, uuids []uuid.UUID) error
}
