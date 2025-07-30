package material

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/gorm/clause"
)

type materialImpl struct {
	*db.Datastore
}

func NewMaterialImpl() repo.MaterialRepo {
	return &materialImpl{
		Datastore: db.DB(),
	}
}

func (m *materialImpl) UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode) error {
	if len(datas) == 0 {
		return nil
	}

	statement := m.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "lab_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"parent_id",
			"lab_id",
			"display_name",
			"description",
			"type",
			// "status",
			"device_node_template_id",
			"reg_id",
			"init_param_data",
			"schema",
			"data",
			"dirs",
			"position",
			"pose",
			"model",
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertMaterialNode err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (m *materialImpl) UpsertMaterialHandle(ctx context.Context, datas []*model.MaterialHandle) error {
	_, _ = ctx, datas
	return nil
}

func (m *materialImpl) UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error {
	_, _ = ctx, datas
	return nil
}
