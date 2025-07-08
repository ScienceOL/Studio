package migrate

import (
	"context"

	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
)

func MigrateTable(ctx context.Context) error {
	return utils.IfErrReturn(func() error {
		return db.DB().DBIns().AutoMigrate(
			&model.Laboratory{},
			&model.Registry{},
			&model.RegAction{},
			&model.DeviceNodeTemplate{},
			&model.DeviceNodeHandleTemplate{},
			&model.DeviceNodeParamTemplate{},
			&model.MaterialNode{},
			&model.MaterialHandle{},
			&model.MaterialEdge{})
	}, func() error {
		return db.DB().DBIns().Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_registry_lab_name_version_active 
			 ON registry (lab_id, name, version) WHERE status != 'del'`).Error
	})
}
