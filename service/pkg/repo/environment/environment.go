package environment

import (
	"context"
	"errors"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	repo "github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MemberCount struct {
	Count int64 `gorm:"column:count"`
	LabID int64 `gorm:"column:lab_id"`
}

type envImpl struct {
	repo.IDOrUUIDTranslate
}

func New() repo.LaboratoryRepo {
	return &envImpl{
		IDOrUUIDTranslate: repo.NewBaseDB(),
	}
}

func (e *envImpl) CreateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error {
	statement := e.DBWithContext(ctx).Create(data)
	if statement.Error != nil {
		logger.Errorf(ctx, "CreateLaboratoryEnv err: %+v", statement.Error)
		return code.CreateDataErr.WithErr(statement.Error)
	}
	return nil
}

func (e *envImpl) UpdateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error {
	statement := e.DBWithContext(ctx).Model(data).Updates(data)
	if statement.Error != nil {
		logger.Errorf(ctx, "UpdateLaboratoryEnv err: %+v", statement.Error)
		return code.UpdateDataErr.WithErr(statement.Error)
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
		return nil, code.QueryRecordErr.WithErr(statement.Error)
	}

	return data, nil
}

func (e *envImpl) GetLabByID(ctx context.Context, labID int64, selectKeys ...string) (*model.Laboratory, error) {
	data := &model.Laboratory{}
	query := e.DBWithContext(ctx).Where("id = ?", labID)
	if len(selectKeys) != 0 {
		query = query.Select(selectKeys)
	}

	statement := query.First(data)
	if statement.Error != nil {
		if errors.Is(statement.Error, gorm.ErrRecordNotFound) {
			logger.Errorf(ctx, "GetLabByID record not found lab_id: %+v", labID)
			return nil, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetLabByID fail lab_id: %+v, sql: %+s, err: %+v",
			labID,
			statement.Statement.SQL.String(),
			statement.Error)
		return nil, code.QueryRecordErr.WithErr(statement.Error)
	}

	return data, nil
}

func (e *envImpl) UpsertWorkflowNodeTemplate(ctx context.Context, datas []*model.WorkflowNodeTemplate) error {
	if len(datas) == 0 {
		return nil
	}
	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "resource_node_id"},
			{Name: "name"}, // res_node_id + name 是唯一约束
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"lab_id",
			"class",
			"goal",
			"goal_default",
			"feedback",
			"result",
			"schema",
			"type",
			"icon",
			"updated_at", // 只更新这些字段，不包括 created_at
		}),
	}).Create(&datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertDeviceAction err: %+v", statement.Error)
		return code.CreateDataErr.WithErr(statement.Error)
	}

	return nil
}

func (e *envImpl) UpsertResourceNodeTemplate(ctx context.Context, datas []*model.ResourceNodeTemplate) error {
	if len(datas) == 0 {
		return nil
	}

	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "lab_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"header",
			"footer",
			"icon",
			"description",
			"model",
			"module",
			"resource_type",
			"language",
			"status_types",
			"tags",
			"data_schema",
			"config_schema",
			"pose",
			"version",
			"updated_at", // 指定需要更新的字段
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertResourceNodeTemplate err: %+v", statement.Error)
		return code.CreateDataErr.WithErr(statement.Error)
	}

	return nil
}

func (e *envImpl) UpsertResourceHandleTemplate(ctx context.Context, datas []*model.ResourceHandleTemplate) error {
	if len(datas) == 0 {
		return nil
	}

	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "resource_node_id"},
			{Name: "name"},
			{Name: "io_type"},
			{Name: "side"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"display_name",
			"type",
			"source",
			"key",
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertResourceHandleTemplate err: %+v", statement.Error)
		return code.CreateDataErr.WithErr(statement.Error)
	}

	return nil
}

func (e *envImpl) GetResourceHandleTemplates(ctx context.Context, resourceNodeIDs []int64) (map[int64][]*model.ResourceHandleTemplate, error) {
	if len(resourceNodeIDs) == 0 {
		return make(map[int64][]*model.ResourceHandleTemplate), nil
	}

	handles := make([]*model.ResourceHandleTemplate, 0, 1)
	statement := e.DBWithContext(ctx).Where("resource_node_id in ?", resourceNodeIDs).Find(&handles)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetDeviceHandelTemplates node ids: %+v, err: %+v", resourceNodeIDs, statement.Error)
		return nil, code.QueryRecordErr
	}

	return utils.SliceToMapSlice(handles, func(h *model.ResourceHandleTemplate) (int64, *model.ResourceHandleTemplate, bool) {
		return h.ResourceNodeID, h, true
	}), nil
}

// 根据 device template node id 获取所有的 uuid
func (e *envImpl) GetResourceNodeTemplates(ctx context.Context, ids []int64) ([]*model.ResourceNodeTemplate, error) {
	if len(ids) == 0 {
		return []*model.ResourceNodeTemplate{}, nil
	}

	datas := make([]*model.ResourceNodeTemplate, 0, len(ids))
	statement := e.DBWithContext(ctx).Select("id, uuid, name").Where("id in ?", ids).Find(&datas)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetResourceNodeTemplateUUID fail ids: %+v, err: %+v", ids, statement.Error)
		return nil, code.QueryRecordErr.WithMsg(statement.Error.Error())
	}

	return datas, nil
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

// 根据实验室
func (e *envImpl) GetLabList(ctx context.Context, userIDs []string, req *common.PageReq) (*common.PageResp[[]*model.Laboratory], error) {
	datas := make([]*model.Laboratory, 0, 1)
	var total int64
	req.Normalize()
	if statement := e.DBWithContext(ctx).
		Model(&model.Laboratory{}).
		Count(&total).
		Where("user_id in ?", userIDs).
		Limit(req.PageSize).
		Offset(req.Offest()).
		Find(&datas); statement.Error != nil {
		logger.Errorf(ctx, "GetLabList fail user ids: %+v, err: %+v", userIDs, statement.Error)
		return nil, code.QueryRecordErr.WithMsg(statement.Error.Error())
	}

	return &common.PageResp[[]*model.Laboratory]{
		Data:     datas,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 创建 action handle
func (e *envImpl) UpsertActionHandleTemplate(ctx context.Context, datas []*model.WorkflowHandleTemplate) error {
	if len(datas) == 0 {
		return nil
	}

	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "workflow_node_id"},
			{Name: "handle_key"},
			{Name: "io_type"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"display_name",
			"type",
			"data_source",
			"data_key",
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertActionHandleTemplate err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}

func (e *envImpl) GetAllResourceName(ctx context.Context, labID int64) []string {
	if labID == 0 {
		return []string{}
	}

	names := make([]string, 0, 1)
	statement := e.DBWithContext(ctx).
		Model(&model.ResourceNodeTemplate{}).Select("name").
		Where("lab_id = ? and resource_type = ?", labID, "device").Find(&names)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetAllResourceName lab id: %+d, err: %+v", labID, statement.Error)
		return []string{}
	}

	return names
}

func (e *envImpl) AddLabMemeber(ctx context.Context, datas ...*model.LaboratoryMember) error {
	if len(datas) == 0 {
		return nil
	}
	statement := e.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "lab_id"},
		},
		DoNothing: true,
	}).Create(&datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "AddLabMemeber err: %+v", statement.Error)
		return code.CreateDataErr.WithErr(statement.Error)
	}

	return nil
}

func (e *envImpl) GetLabByUserID(ctx context.Context, req *common.PageReqT[string]) (*common.PageResp[[]*model.LaboratoryMember], error) {
	if req.Data == "" {
		return nil, code.UserIDIsEmptyErr
	}

	query := e.DBWithContext(ctx).Where("user_id  = ?", req.Data)
	return e.getLabMemeber(ctx, query, &req.PageReq)
}

func (e *envImpl) GetLabByLabID(ctx context.Context, req *common.PageReqT[int64]) (*common.PageResp[[]*model.LaboratoryMember], error) {
	if req.Data == 0 {
		return nil, code.LabIDIsEmptyErr
	}

	query := e.DBWithContext(ctx).Where("lab_id = ?", req.Data)
	return e.getLabMemeber(ctx, query, &req.PageReq)
}

func (e *envImpl) getLabMemeber(ctx context.Context, query *gorm.DB, req *common.PageReq) (*common.PageResp[[]*model.LaboratoryMember], error) {
	var total int64
	datas := make([]*model.LaboratoryMember, 0, 1)
	if err := query.
		Model(&model.LaboratoryMember{}).
		Count(&total).
		Offset(req.Offest()).
		Limit(req.PageSize).
		Order("id asc").
		Find(&datas).Error; err != nil {
		logger.Errorf(ctx, "getLabMemeber fail err: %+v", err)
		return nil, code.QueryRecordErr.WithErr(err)
	}

	return &common.PageResp[[]*model.LaboratoryMember]{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     datas,
	}, nil
}

func (e *envImpl) GetLabMemberCount(ctx context.Context, labIDs ...int64) map[int64]int64 {
	datas := make([]*MemberCount, 0, 10)

	query := e.DBWithContext(ctx).Table("laboratory_member").
		Select("COUNT(user_id) as count, lab_id")
	if len(labIDs) > 0 {
		query.Where("lab_id in ?", labIDs)
	}

	if err := query.
		Group("lab_id").
		Scan(&datas).Error; err != nil {
		return map[int64]int64{}
	}

	return utils.Slice2Map(datas, func(m *MemberCount) (int64, int64) {
		return m.LabID, m.Count
	})
}
