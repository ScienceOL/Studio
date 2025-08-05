package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type NodeInfo struct {
	NodeUUID   common.BinUUID `json:"node_uuid"`
	HandleUUID common.BinUUID `json:"handle_uuid"`
}

type DelNodeInfo struct {
	NodeUUID []common.BinUUID `json:"node_uuid"`
	EdgeUUID []common.BinUUID `json:"edge_uuid"`
}

type MaterialRepo interface {
	UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode) error
	UpsertMaterialHandle(ctx context.Context, datas []*model.MaterialHandle) error
	UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error
	GetNodeHandles(ctx context.Context, labID int64, nodeNames []string, handleNames []string) (map[string]map[string]NodeInfo, error)
	DelNodes(ctx context.Context, nodeUUIDs []common.BinUUID) (*DelNodeInfo, error)
}
