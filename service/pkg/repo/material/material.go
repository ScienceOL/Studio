package material

import (
	"context"

	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type materialImpl struct {
	*db.Datastore
}

func NewMaterialImpl() repo.MaterialRepo {
	return &materialImpl{
		Datastore: db.DB(),
	}
}

func (materialimpl *materialImpl) UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode) error {
	_, _ = ctx, datas
	return nil
}
func (materialimpl *materialImpl) UpsertMaterialHandle(ctx context.Context, datas []*model.MaterialHandle) error {
	_, _ = ctx, datas
	return nil
}
func (materialimpl *materialImpl) UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error {
	_, _ = ctx, datas
	return nil
}
