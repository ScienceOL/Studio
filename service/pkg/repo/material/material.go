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

// delete nodes、handles、edges
func (m *materialImpl) DelNodes(ctx context.Context, nodeUUIDs []common.BinUUID) (*repo.DelNodeInfo, error) {
	if len(nodeUUIDs) == 0 {
		return &repo.DelNodeInfo{}, nil
	}

	res := &repo.DelNodeInfo{}
	if err := m.ExecTx(ctx, func(txCtx context.Context) error {
		// 获取所有删除 id 的 node id 和 uuid
		// TODO: node 的 parent id 如果被删除了，所有子 node 要删除么？目前处理是吧 parent id 设置为 0
		delNodes := []*model.MaterialNode{}
		if err := m.DBWithContext(txCtx).
			Select("id, uuid").
			Where("uuid in ?", nodeUUIDs).
			Find(&delNodes).Error; err != nil {
			logger.Errorf(ctx, "DelNodes query node fail uuids: %+v, err: %+v", nodeUUIDs, err)
			return code.QueryRecordErr
		}

		nodeIDs := make([]int64, 0, len(delNodes))
		nodeUUIDs := make([]common.BinUUID, 0, len(delNodes))
		for _, n := range delNodes {
			nodeIDs = append(nodeIDs, n.ID)
			nodeUUIDs = append(nodeUUIDs, n.UUID)
		}
		if len(nodeIDs) == 0 {
			return nil
		}

		// 获取所有待删除的 node 的 handle id 和 uuid
		delNodeHandles := []*model.MaterialHandle{}
		if err := m.DBWithContext(txCtx).
			Select("id, node_id, uuid").
			Where("node_id in ?", nodeIDs).
			Find(&delNodeHandles).Error; err != nil {
			logger.Errorf(ctx, "DelNodes query handle fail uuids: %+v, err: %+v", nodeUUIDs, err)
			return code.QueryRecordErr
		}
		handleUUIDs := make([]common.BinUUID, 0, 2*len(delNodeHandles))
		for _, h := range delNodeHandles {
			handleUUIDs = append(handleUUIDs, h.UUID)
		}

		delNodeEdges := []*model.MaterialEdge{}
		if err := m.DBWithContext(txCtx).
			Select("id, uuid").
			Where("source_node_uuid in ? or target_node_uuid in ? ", nodeUUIDs, nodeUUIDs).
			Find(&delNodeEdges).Error; err != nil {
			logger.Errorf(ctx, "DelNodes query edge fail uuids: %+v, err: %+v", nodeUUIDs, err)
			return code.QueryRecordErr
		}
		edgeIDs := make([]int64, 0, len(delNodeEdges))
		edgeUUIDs := make([]common.BinUUID, 0, len(delNodeEdges))
		for _, e := range delNodeEdges {
			edgeIDs = append(edgeIDs, e.ID)
			edgeUUIDs = append(edgeUUIDs, e.UUID)
		}

		// 删除节点
		if err := m.DBWithContext(txCtx).Delete(&model.MaterialNode{}, nodeIDs).Error; err != nil {
			logger.Errorf(txCtx, "DelNodes fail ids: %+v, err: %+v", nodeIDs, err)
			return code.DeleteDateErr.WithMsg(err.Error())
		}
		// 删除 handle
		if err := m.DBWithContext(txCtx).Where("node_id in ?", nodeIDs).Delete(&model.MaterialHandle{}).Error; err != nil {
			logger.Errorf(txCtx, "DelNodes fail ids: %+v, err: %+v", nodeIDs, err)
			return code.DeleteDateErr.WithMsg(err.Error())
		}

		// 删除 edge
		if err := m.DBWithContext(txCtx).Where("id in ?", edgeIDs).Delete(&model.MaterialEdge{}).Error; err != nil {
			logger.Errorf(txCtx, "DelNodes fail ids: %+v, err: %+v", nodeIDs, err)
			return code.DeleteDateErr.WithMsg(err.Error())
		}
		res.NodeUUID = nodeUUIDs
		res.EdgeUUID = edgeUUIDs

		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}
