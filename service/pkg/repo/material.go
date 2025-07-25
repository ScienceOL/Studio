package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/repo/model"
)

type MaterialRepo interface {
	UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode) error
	UpsertMaterialHandle(ctx context.Context, datas []*model.MaterialHandle) error
	UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error
}
