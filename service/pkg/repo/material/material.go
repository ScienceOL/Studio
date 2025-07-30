package material

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/gorm/clause"
)

type NodeHandleInfo struct {
	NodeName   string         `gorm:"column:node_name"`
	NodeUUID   common.BinUUID `gorm:"column:node_uuid"`
	HandleName string         `gorm:"column:handle_name"`
	HandleUUID common.BinUUID `gorm:"column:handle_uuid"`
}

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
	if len(datas) == 0 {
		return nil
	}

	statement := m.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "node_id"},
			{Name: "name"},
			{Name: "side"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"display_name",
			"type",
			"io_type",
			"source",
			"key",
			"connected",
			"required",
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertMaterialHandle err: %+v", statement.Error)
		return code.CreateDataErr
	}

	return nil
}

func (m *materialImpl) UpsertMaterialEdge(ctx context.Context, datas []*model.MaterialEdge) error {
	if len(datas) == 0 {
		return nil
	}

	statement := m.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "source_node_uuid"},
			{Name: "target_node_uuid"},
			{Name: "source_handle_uuid"},
			{Name: "target_handle_uuid"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertMaterialEdge err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}

func (m *materialImpl) GetNodeHandles(
	ctx context.Context,
	labID int64,
	nodeNames []string,
	handleNames []string) (map[string]map[string]repo.NodeInfo, error) {
	res := make([]*NodeHandleInfo, 0, len(handleNames))
	if err := m.DBWithContext(ctx).Table("material_node as n").
		Select("n.uuid as node_uuid, n.name as node_name, h.uuid as handle_uuid, h.name as handle_name").
		Joins("inner join material_handle as h on n.id = h.node_id").
		Where("n.name in ? and h.name in ?", nodeNames, handleNames).
		Find(&res).Error; err != nil {
		logger.Errorf(ctx, "GetNodeHandles fail lab id: %d, node names: %+v, handle names: %+v", labID, nodeNames, handleNames)

		return nil, code.QueryRecordErr
	}

	resMap := make(map[string]map[string]repo.NodeInfo)
	for _, info := range res {
		if _, ok := resMap[info.NodeName]; !ok {
			resMap[info.NodeName] = make(map[string]repo.NodeInfo)
		}
		resMap[info.NodeName][info.HandleName] = repo.NodeInfo{
			NodeUUID:   info.NodeUUID,
			HandleUUID: info.HandleUUID,
		}
	}

	return resMap, nil
}
