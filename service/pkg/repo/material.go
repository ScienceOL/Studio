package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/model"
	"gorm.io/gorm/schema"
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
	UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids ...uuid.UUID) map[uuid.UUID]int64
	ID2UUID(ctx context.Context, tableModel schema.Tabler, ids ...int64) map[int64]uuid.UUID
	FindDatas(ctx context.Context, datas any, condition map[string]any, keys ...string) error
	GetData(ctx context.Context, data schema.Tabler, condition map[string]any, keys ...string) error
	ExecTx(ctx context.Context, fn func(txCtx context.Context) error) error
	UpdateData(ctx context.Context, data any, condition map[string]any, keys ...string) error
	// 更新或者插入物料
	UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode, conflictKeys []string, returns []string, keys ...string) ([]*model.MaterialNode, error)
	// 更新或插入 edge
	UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error
	// 批量更新节点数据，根据 uuid , uuid 不存在不创建新数据
	// UpsertMaterialNodes(ctx context.Context, datas []*model.MaterialNode)
	// 获取所有的 node handle
	GetNodeHandles(ctx context.Context, labID int64, nodeNames []string, handleNames []string) (map[string]map[string]NodeInfo, error)
	// 根据 uuid 获取到所有 node 和 handle uuid
	GetNodeHandlesByUUID(ctx context.Context, nodeUUIDs []uuid.UUID) (map[uuid.UUID]map[uuid.UUID]NodeInfo, error)
	// 根据 uuid 获取到所有 node 和 handle name
	GetNodeHandlesByUUIDV1(ctx context.Context, nodeUUIDs []uuid.UUID) (map[uuid.UUID]map[string]NodeInfo, error)
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
	// 创建仿真机器记录
	UpsertMachine(ctx context.Context, data *model.MaterialMachine) error
	// 获取指定模板的第一个设备
	GetFirstDevice(ctx context.Context, resID int64) *string
	// 根据路径查询节点
	GetMaterialNodeByPath(ctx context.Context, labID int64, names []string) ([]*model.MaterialNode, error)
	// 获取所有节点的子孙节点，不包含该节点本身
	GetDescendants(ctx context.Context, labID int64, nodeID int64) ([]*model.MaterialNode, error)
	// 更新 data 的指定 key
	UpdateMaterialNodeDataKey(ctx context.Context, labID int64, deviceName string, key string, value any) ([]*model.MaterialNode, error)
	// 根据节点的 uuid 获取所有的父节点，包含自身
	GetAncestors(ctx context.Context, nodeUUID uuid.UUID) ([]*model.MaterialNode, error)
}
