package migrate

import (
	"context"

	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
)

func Table(_ context.Context) error {
	return utils.IfErrReturn(func() error {
		return db.DB().DBIns().AutoMigrate(
			&model.Laboratory{},
			&model.DeviceAction{},
			&model.ResourceNodeTemplate{},
			&model.ResourceHandleTemplate{},
			&model.MaterialNode{},
			&model.MaterialEdge{})
	})
}
