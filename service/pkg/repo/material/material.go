package material

import (
	"context"
	"errors"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NodeHandleInfo struct {
	NodeName   string    `gorm:"column:node_name"`
	NodeUUID   uuid.UUID `gorm:"column:node_uuid"`
	HandleName string    `gorm:"column:handle_name"`
	HandleUUID uuid.UUID `gorm:"column:handle_uuid"`
}

type materialImpl struct {
	repo.IDOrUUIDTranslate
}

func NewMaterialImpl() repo.MaterialRepo {
	return &materialImpl{
		IDOrUUIDTranslate: repo.NewBaseDB(),
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
			{Name: "parent_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"display_name",
			"description",
			"status",
			"type",
			"resource_node_id",
			"class",
			"init_param_data",
			"schema",
			"data",
			"pose",
			"model",
			"icon",
			"updated_at",
		}),
	}).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertMaterialNode err: %+v", statement.Error)
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
		DoNothing: true,
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
	handleNames []string,
) (map[string]map[string]repo.NodeInfo, error) {
	res := make([]*NodeHandleInfo, 0, len(handleNames))
	if err := m.DBWithContext(ctx).Table("material_node as n").
		Select("n.uuid as node_uuid, n.name as node_name, h.uuid as handle_uuid, h.name as handle_name").
		Joins("inner join resource_handle_template as h on n.resource_node_id = h.resource_node_id").
		Where("n.lab_id = ? and n.name in ? and h.name in ?", labID, nodeNames, handleNames).
		Find(&res).Error; err != nil {
		logger.Errorf(ctx, "GetNodeHandles fail lab id: %d, node names: %+v, handle names: %+v, err: %+v", labID, nodeNames, handleNames, err)

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

// 根据 uuid 获取到所有 node 和 handle
func (m *materialImpl) GetNodeHandlesByUUID(ctx context.Context, nodeUUIDs []uuid.UUID) (map[uuid.UUID]map[uuid.UUID]repo.NodeInfo, error) {
	if len(nodeUUIDs) == 0 {
		return make(map[uuid.UUID]map[uuid.UUID]repo.NodeInfo), nil
	}

	res := make([]*NodeHandleInfo, 0, len(nodeUUIDs))
	if err := m.DBWithContext(ctx).Table("material_node as n").
		Select("n.uuid as node_uuid, h.uuid as handle_uuid").
		Joins("inner join resource_handle_template as h on n.resource_node_id = h.resource_node_id").
		Where("n.uuid in ?", nodeUUIDs).
		Find(&res).Error; err != nil {
		logger.Errorf(ctx, "GetNodeHandlesByUUID fail node uuids: %+v, err: %+v", nodeUUIDs, err)

		return nil, code.QueryRecordErr
	}

	resMap := make(map[uuid.UUID]map[uuid.UUID]repo.NodeInfo)
	for _, info := range res {
		if _, ok := resMap[info.NodeUUID]; !ok {
			resMap[info.NodeUUID] = make(map[uuid.UUID]repo.NodeInfo)
		}
		resMap[info.NodeUUID][info.HandleUUID] = repo.NodeInfo{
			NodeUUID:   info.NodeUUID,
			HandleUUID: info.HandleUUID,
		}
	}

	return resMap, nil
}

// delete nodes、handles、edges
func (m *materialImpl) DelNodes(ctx context.Context, nodeUUIDs []uuid.UUID) (*repo.DelNodeInfo, error) {
	if len(nodeUUIDs) == 0 {
		return &repo.DelNodeInfo{}, nil
	}

	res := &repo.DelNodeInfo{}
	if err := m.ExecTx(ctx, func(txCtx context.Context) error {
		// 获取所有删除 id 的 node id 和 uuid
		delNodes := []*model.MaterialNode{}
		if err := m.DBWithContext(txCtx).Clauses(
			clause.Returning{
				Columns: []clause.Column{
					{Name: "id"},
					{Name: "uuid"},
				},
			}).
			Where("uuid in ?", nodeUUIDs).
			Delete(&delNodes).Error; err != nil {
			logger.Errorf(ctx, "DelNodes delete query node fail uuids: %+v, err: %+v", nodeUUIDs, err)
			return code.DeleteDataErr.WithErr(err)
		}

		res.NodeUUIDs = utils.FilterSlice(delNodes, func(node *model.MaterialNode) (uuid.UUID, bool) {
			return node.UUID, true
		})

		if len(res.NodeUUIDs) == 0 {
			return nil
		}

		delNodeEdges := []*model.MaterialEdge{}
		if err := m.DBWithContext(txCtx).Clauses(
			clause.Returning{
				Columns: []clause.Column{
					{Name: "id"},
					{Name: "uuid"},
				},
			}).
			Where("source_node_uuid in ? or target_node_uuid in ? ", nodeUUIDs, nodeUUIDs).
			Delete(&delNodeEdges).Error; err != nil {
			logger.Errorf(ctx, "DelNodes delete edge fail uuids: %+v, err: %+v", nodeUUIDs, err)
			return code.QueryRecordErr.WithErr(err)
		}

		res.EdgeUUIDs = utils.FilterSlice(delNodeEdges, func(edge *model.MaterialEdge) (uuid.UUID, bool) {
			return edge.UUID, true
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}

// 获取所有物料根据 lab id
func (m *materialImpl) GetNodesByLabID(ctx context.Context, labID int64, selectKeys ...string) ([]*model.MaterialNode, error) {
	datas := make([]*model.MaterialNode, 0, 1)
	if labID == 0 {
		return datas, nil
	}
	query := m.DBWithContext(ctx).Where("lab_id = ?", labID)
	if len(selectKeys) != 0 {
		query = query.Select(selectKeys)
	}

	statement := query.Order("id asc").Find(&datas)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetNodesByLabID sql: %+s, err: %+v",
			statement.Statement.SQL.String(),
			statement.Error)
		return nil, code.QueryRecordErr
	}

	return datas, nil
}

// 根据所有的 uuid 获取所有的edges
func (m *materialImpl) GetEdgesByNodeUUID(ctx context.Context, uuids []uuid.UUID, selectKeys ...string) ([]*model.MaterialEdge, error) {
	datas := make([]*model.MaterialEdge, 0, 1)
	if len(uuids) == 0 {
		return datas, nil
	}
	query := m.DBWithContext(ctx).Where("source_node_uuid in ? or target_node_uuid in ?", uuids, uuids)
	if len(selectKeys) != 0 {
		query = query.Select(selectKeys)
	}

	statement := query.Find(&datas)
	if statement.Error != nil {
		logger.Errorf(ctx, "GetEdgesByNodeUUID sql: %+s, err: %+v",
			statement.Statement.SQL.String(),
			statement.Error)
		return nil, code.QueryRecordErr
	}

	return datas, nil
}

// 批量 edges
func (m *materialImpl) DelEdges(ctx context.Context, uuids []uuid.UUID) error {
	if len(uuids) == 0 {
		return nil
	}

	if err := m.DBWithContext(ctx).Where("uuid in ?", uuids).Delete(&model.MaterialEdge{}).Error; err != nil {
		logger.Errorf(ctx, "DelEdges fail ids: %+v, err: %+v", uuids, err)
		return code.DeleteDataErr.WithMsg(err.Error())
	}

	return nil
}

// 批量跟新 node 数据
func (m *materialImpl) UpdateNodeByUUID(ctx context.Context, data *model.MaterialNode, selectKeys ...string) error {
	if err := m.DBWithContext(ctx).
		Model(&model.MaterialNode{}).
		Select(selectKeys).
		Where("uuid = ?", data.UUID).
		Updates(data).Error; err != nil {
		logger.Errorf(ctx, "UpdateNodeByUUID fail data: %+v, err: %+v", data, err)
		return code.UpdateDataErr.WithMsg(err.Error())
	}

	return nil
}

// 根据 uuid 获取节点 ID
func (m *materialImpl) GetNodeIDByUUID(ctx context.Context, nodeUUID uuid.UUID) (int64, error) {
	data := &model.MaterialNode{}
	if err := m.DBWithContext(ctx).
		Select("id, uuid").
		Where("uuid = ?", nodeUUID).
		First(data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetNodeIDByUUID query node fail uuid: %s, err: %+v", nodeUUID, err)
		return 0, code.QueryRecordErr
	}

	return data.ID, nil
}
