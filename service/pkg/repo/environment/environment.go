package environment

import (
	"context"
	"errors"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	repo "github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type envImpl struct {
	*db.Datastore
}

func NewEnv() repo.EnvRepo {
	return &envImpl{
		Datastore: db.DB(),
	}
}

func (e *envImpl) CreateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error {
	statement := e.DBWithContext(ctx).Create(data)
	if statement.Error != nil {
		logger.Errorf(ctx, "CreateLaboratoryEnv err: %+v", statement.Error)
		return code.CreateDataErr
	}
	return nil
}

func (e *envImpl) UpdateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error {
	statement := e.DBWithContext(ctx).Model(data).Updates(data)
	if statement.Error != nil {
		logger.Errorf(ctx, "UpdateLaboratoryEnv err: %+v", statement.Error)
		return code.CreateDataErr
	}
	return nil
}

func (e *envImpl) GetLabByUUID(ctx context.Context, UUID common.BinUUID) (*model.Laboratory, error) {
	data := &model.Laboratory{}
	statement := e.DBWithContext(ctx).Where("uuid = ?", UUID).First(data)
	if statement.Error != nil {
		if errors.Is(statement.Error, gorm.ErrRecordNotFound) {
			logger.Errorf(ctx, "GetLabBy uuid: %+v record not found", UUID)
			return nil, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetLabBy uuid: %+v, sql: %+s, err: %+v",
			UUID,
			statement.Statement.SQL.String(),
			statement.Error)
		return nil, code.QueryRecordErr
	}

	return data, nil
}

func (e *envImpl) CreateReg(ctx context.Context, data *model.Registry) error {
	statement := e.DBWithContext(ctx).Where("lab_id = ? and name = ? and version = ?",
		data.LabID, data.Name, data.Version).FirstOrCreate(data)
	if statement.Error != nil {
		logger.Errorf(ctx, "CreateReg err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (e *envImpl) UpsertRegAction(ctx context.Context, datas []*model.RegAction) error {
	if len(datas) == 0 {
		return nil
	}
	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "reg_id"},
			{Name: "name"}, // reg_id + name 是唯一约束
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"goal",
			"goal_default",
			"feedback",
			"result",
			"schema",
			"type",
			"handles",
			"updated_at", // 只更新这些字段，不包括 created_at
		}),
	}).Create(&datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertDeviceAction err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (e *envImpl) UpsertDeviceTemplate(ctx context.Context, data *model.DeviceNodeTemplate) error {
	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "lab_id"},
			{Name: "reg_id"},
			{Name: "name"},
			{Name: "version"}, // 根据 idx_lrnv 推测是这些字段的组合
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"description",
			"icon",
			"header",
			"footer",
			"updated_at", // 指定需要更新的字段
		}),
	}).Create(data)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertDeviceTemplate err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (e *envImpl) UpsertDeviceHandleTemplate(ctx context.Context, datas []*model.DeviceNodeHandleTemplate) error {
	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "node_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"display_name",
			"type",
			"io_type",
			"source",
			"key",
			"side",
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertDeviceHandleTemplate err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (e *envImpl) UpsertDeviceParamTemplate(ctx context.Context, datas []*model.DeviceNodeParamTemplate) error {
	if len(datas) == 0 {
		return nil
	}

	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "node_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"placeholder",
			"type",
			"schema",
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertDeviceParamTemplate err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (e *envImpl) GetRegs(ctx context.Context, labID int64, names []string) (map[string]*repo.RegDeviceInfo, error) {
	regs := make([]*repo.RegDeviceInfo, 0, len(names))
	err := e.DBWithContext(ctx).Raw(`
   		SELECT ranked.name as reg_name, ranked.id as reg_id, device_node_template.id as device_node_template_id FROM (
       SELECT id, name, 
              ROW_NUMBER() OVER (PARTITION BY name ORDER BY version DESC) as rn
       FROM registry 
       WHERE lab_id = ? AND name in ? AND status != ? 
   ) ranked join device_node_template on device_node_template.reg_id = ranked.id WHERE ranked.rn = 1;
    `, labID, names, model.REGDEL).Scan(&regs).Error
	if err != nil {
		logger.Errorf(ctx, "GetRegs lab id: %d, names: %+v, err: %+v", labID, names, err)
		return nil, code.QueryRecordErr
	}

	regMap := make(map[string]*repo.RegDeviceInfo, len(names))
	for _, reg := range regs {
		regMap[reg.RegName] = reg
	}

	return regMap, nil
}

func (e *envImpl) GetDeviceTemplateHandels(ctx context.Context, deviceNodeID int64) ([]*model.DeviceNodeHandleTemplate, error) {
	handles := make([]*model.DeviceNodeHandleTemplate, 0, 1)
	statement := e.DBWithContext(ctx).Where("node_id = ?", deviceNodeID).Find(&handles)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetDeviceTemplateHandels node id: %d, err: %+v", deviceNodeID, statement.Error)
		return nil, code.QueryRecordErr
	}

	return handles, nil
}
