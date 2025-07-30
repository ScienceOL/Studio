package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type NodeInfo struct {
	NodeUUID   common.BinUUID
	HandleUUID common.BinUUID
}

type MaterialRepo interface {
	UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode) error
	UpsertMaterialHandle(ctx context.Context, datas []*model.MaterialHandle) error
	UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error
	GetNodeHandles(ctx context.Context, labID int64, nodeNames []string, handleNames []string) (map[string]map[string]NodeInfo, error)
}
