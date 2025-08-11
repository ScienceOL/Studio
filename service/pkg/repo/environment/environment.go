package environment

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	repo "github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
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
		return code.UpdateDataErr
	}
	return nil
}

func (e *envImpl) GetLabByUUID(ctx context.Context, UUID uuid.UUID, selectKeys ...string) (*model.Laboratory, error) {
	data := &model.Laboratory{}
	query := e.DBWithContext(ctx).Where("uuid = ?", UUID)
	if len(selectKeys) != 0 {
		query = query.Select(selectKeys)
	}

	statement := query.First(data)
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

func (e *envImpl) UpsertDeviceAction(ctx context.Context, datas []*model.DeviceAction) error {
	if len(datas) == 0 {
		return nil
	}
	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "res_node_id"},
			{Name: "name"}, // res_node_id + name 是唯一约束
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

func (e *envImpl) UpsertDeviceTemplate(ctx context.Context, datas []*model.ResourceNodeTemplate) error {
	if len(datas) == 0 {
		return nil
	}
	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "lab_id"},
			{Name: "name"},
			{Name: "version"}, // 根据 idx_lrnv 推测是这些字段的组合
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"description",
			"icon",
			"header",
			"footer",
			"updated_at", // 指定需要更新的字段
			"module",
			"model",
			"language",
			"status_types",
			"data_schema",
			"config_schema",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertDeviceTemplate err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (e *envImpl) UpsertDeviceHandleTemplate(ctx context.Context, datas []*model.ResourceHandleTemplate) error {
	if len(datas) == 0 {
		return nil
	}

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

func (e *envImpl) GetResourceTemplate(ctx context.Context, labID int64, names []string) (map[string]*model.ResourceNodeTemplate, error) {
	datas := make([]*model.ResourceNodeTemplate, 0, len(names))
	statement := e.DBWithContext(ctx).Where("lab_id = ? and name in ?", labID, names).Find(&datas)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetResourceTemplate fail lab id: %d, names: %s, err: %+v", labID, names, statement.Error)
		return nil, code.QueryRecordErr.WithMsg(statement.Error.Error())
	}

	return utils.SliceToMap(datas, func(data *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
		return data.Name, data
	}), nil
}

func (e *envImpl) GetResourceHandleTemplates(ctx context.Context, resIDs []int64) (map[int64][]*model.ResourceHandleTemplate, error) {
	if len(resIDs) == 0 {
		return make(map[int64][]*model.ResourceHandleTemplate), nil
	}

	handles := make([]*model.ResourceHandleTemplate, 0, 1)
	statement := e.DBWithContext(ctx).Where("node_id in ?", resIDs).Find(&handles)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetDeviceHandelTemplates node id: %+v, err: %+v", resIDs, statement.Error)
		return nil, code.QueryRecordErr
	}

	res := make(map[int64][]*model.ResourceHandleTemplate)
	for _, h := range handles {
		res[h.NodeID] = append(res[h.NodeID], h)
	}
	return res, nil
}

// 根据 device template node id 获取所有的 uuid
func (e *envImpl) GetResourceNodeTemplateUUID(ctx context.Context, resIDs []int64) (map[int64]uuid.UUID, error) {
	if len(resIDs) == 0 {
		return make(map[int64]uuid.UUID), nil
	}

	datas := make([]*model.ResourceNodeTemplate, 0, len(resIDs))
	statement := e.DBWithContext(ctx).Select("id, uuid").Where("id in ?", resIDs).Find(&datas)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetResourceNodeTemplateUUID fail ids: %+v, err: %+v", resIDs, statement.Error)
		return nil, code.QueryRecordErr.WithMsg(statement.Error.Error())
	}

	return utils.SliceToMap(datas, func(item *model.ResourceNodeTemplate) (int64, uuid.UUID) {
		return item.ID, item.UUID
	}), nil
}

func (e *envImpl) GetLabByAkSk(ctx context.Context, accessKey string, accessSecret string) (*model.Laboratory, error) {
	data := &model.Laboratory{}
	statement := e.DBWithContext(ctx).Where("access_key= ? and access_secret = ?", accessKey, accessSecret).First(data)
	if statement.Error != nil {
		if errors.Is(statement.Error, gorm.ErrRecordNotFound) {
			logger.Errorf(ctx, "GetLabByAkSk not found")
			return nil, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetLabByAkSk sql: %+s, err: %+v",
			statement.Statement.SQL.String(),
			statement.Error)
		return nil, code.QueryRecordErr
	}

	return data, nil
}

// 根据实验室 id 获取所有的模板信息
func (e *envImpl) GetAllResourceTemplateByLabID(ctx context.Context, labID int64, selectKeys ...string) ([]*model.ResourceNodeTemplate, error) {
	datas := make([]*model.ResourceNodeTemplate, 0, 1)
	if labID == 0 {
		return datas, nil
	}
	query := e.DBWithContext(ctx).Where("lab_id = ?", labID)
	if len(selectKeys) != 0 {
		query = query.Select(selectKeys)
	}

	statement := query.Find(&datas)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetAllResourceTemplateByLabID sql: %+s, err: %+v",
			statement.Statement.SQL.String(),
			statement.Error)
		return nil, code.QueryRecordErr
	}

	return datas, nil
}

// 根据 device ids 获取所有的 handles
func (e *envImpl) GetAllDeviceTemplateHandlesByID(
	ctx context.Context,
	templateIDs []int64,
	selectKeys ...string) (
	[]*model.ResourceHandleTemplate, error,
) {
	datas := make([]*model.ResourceHandleTemplate, 0, 1)
	if len(templateIDs) == 0 {
		return datas, nil
	}
	query := e.DBWithContext(ctx).Where("node_id in ?", templateIDs)
	if len(selectKeys) != 0 {
		query = query.Select(selectKeys)
	}

	statement := query.Find(&datas)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetAllDeviceTemplateHandlesByID sql: %+s, err: %+v",
			statement.Statement.SQL.String(),
			statement.Error)
		return nil, code.QueryRecordErr
	}

	return datas, nil
}

// 根据 uuid 获取 template 数据
func (e *envImpl) GetResourceTemplateByUUD(ctx context.Context, uuid uuid.UUID, selectKeys ...string) (*model.ResourceNodeTemplate, error) {
	if uuid.IsNil() {
		return nil, code.QueryRecordErr
	}

	data := &model.ResourceNodeTemplate{}
	query := e.DBWithContext(ctx).Where("uuid = ?", uuid)
	if len(selectKeys) != 0 {
		query = query.Select(selectKeys)
	}
	statement := query.First(data)
	if statement.Error != nil {
		if errors.Is(statement.Error, gorm.ErrRecordNotFound) {
			return nil, code.RecordNotFound
		}
		logger.Errorf(ctx, "GetResourceTemplateByUUD fail uuid: %+v, err: %+v", uuid, statement.Error)
		return nil, code.QueryRecordErr.WithMsg(statement.Error.Error())
	}

	return data, nil
}

// 根据实验室
func (e *envImpl) GetLabList(ctx *context.Context, userID string, req *common.PageReq) (*common.PageResp, error) {
	// datas := make([]*model.Laboratory, 0, 1)
	// var total int64

	// query := e.DBWithContext(ctx).Count(&total).Where("")

	return nil, nil
}
