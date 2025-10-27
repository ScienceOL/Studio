package material

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"
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

func (m *materialImpl) UpsertMaterialNode(ctx context.Context, datas []*model.MaterialNode, conflictKeys []string, returns []string, keys ...string) ([]*model.MaterialNode, error) {
	if len(datas) == 0 {
		return nil, nil
	}

	updateKeys := []string{
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
	}

	if len(keys) > 0 {
		updateKeys = keys
	}

	defaultColumns := []clause.Column{
		{Name: "lab_id"},
		{Name: "name"},
		{Name: "parent_id"},
	}
	if len(conflictKeys) > 0 {
		defaultColumns = utils.FilterSlice(conflictKeys, func(name string) (clause.Column, bool) {
			return clause.Column{Name: name}, true
		})
	}

	clauses := make([]clause.Expression, 0, 2)
	clauses = append(clauses, clause.OnConflict{
		Columns:   defaultColumns,
		DoUpdates: clause.AssignmentColumns(updateKeys),
	})

	if len(returns) > 0 {
		r := clause.Returning{}
		for _, key := range returns {
			r.Columns = append(r.Columns, clause.Column{
				Name: key,
			})
		}
		clauses = append(clauses, r)
	}

	statement := m.DBWithContext(ctx).Clauses(clauses...).Create(datas)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertMaterialNode err: %+v", statement.Error)
		return nil, code.CreateDataErr
	}

	return datas, nil
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
		DoUpdates: clause.Assignments(map[string]any{
			"updated_at": time.Now(),
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

func (m *materialImpl) GetNodeHandlesByUUIDV1(ctx context.Context, nodeUUIDs []uuid.UUID) (map[uuid.UUID]map[string]repo.NodeInfo, error) {
	if len(nodeUUIDs) == 0 {
		return make(map[uuid.UUID]map[string]repo.NodeInfo), nil
	}

	res := make([]*NodeHandleInfo, 0, len(nodeUUIDs))
	if err := m.DBWithContext(ctx).Table("material_node as n").
		Select("n.uuid as node_uuid, h.uuid as handle_uuid, h.name as handle_name").
		Joins("inner join resource_handle_template as h on n.resource_node_id = h.resource_node_id").
		Where("n.uuid in ?", nodeUUIDs).
		Find(&res).Error; err != nil {
		logger.Errorf(ctx, "GetNodeHandlesByUUID fail node uuids: %+v, err: %+v", nodeUUIDs, err)

		return nil, code.QueryRecordErr
	}

	resMap := make(map[uuid.UUID]map[string]repo.NodeInfo)
	for _, info := range res {
		if _, ok := resMap[info.NodeUUID]; !ok {
			resMap[info.NodeUUID] = make(map[string]repo.NodeInfo)
		}
		resMap[info.NodeUUID][info.HandleName] = repo.NodeInfo{
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
		if err := m.DBWithContext(txCtx).Raw(`
        WITH RECURSIVE subtree AS (
            SELECT id, uuid
            FROM material_node
            WHERE uuid in ?
            UNION ALL
            SELECT mn.id, mn.uuid
            FROM material_node mn
            INNER JOIN subtree st ON mn.parent_id = st.id
        )
        DELETE FROM material_node
        WHERE id IN (SELECT id FROM subtree)
        RETURNING uuid
    `, nodeUUIDs).Scan(&delNodes).Error; err != nil {
			logger.Errorf(ctx, "delNodes delete node fail uuids: %+v, err: %+v", nodeUUIDs, err)
			return code.DeleteDataErr.WithErr(err)
		}

		res.NodeUUIDs = utils.FilterSlice(delNodes, func(node *model.MaterialNode) (uuid.UUID, bool) {
			return node.UUID, true
		})

		delNodeEdges := []*model.MaterialEdge{}
		if err := m.DBWithContext(txCtx).Clauses(
			clause.Returning{
				Columns: []clause.Column{
					{Name: "uuid"},
				},
			}).
			Where("source_node_uuid in ? or target_node_uuid in ? ", res.NodeUUIDs, res.NodeUUIDs).
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

func (m *materialImpl) UpsertMachine(ctx context.Context, data *model.MaterialMachine) error {
	data.UpdatedAt = time.Now()
	statement := m.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "lab_id"},
			{Name: "user_id"},
			{Name: "image_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"updated_at": time.Now(),
			"machine_id": data.MachineID,
		}),
	}).Create(data)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertMachine err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}

func (m *materialImpl) GetFirstDevice(ctx context.Context, resID int64) *string {
	firstData := &model.MaterialNode{}
	if err := m.DBWithContext(ctx).
		Where("resource_node_id = ?", resID).
		Select("name").
		First(&firstData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		logger.Errorf(ctx, "GetFirstDevice can not found material node resource id: %d", resID)
		return nil
	}

	return &firstData.Name
}

func (m *materialImpl) GetMaterialNodeByPath(ctx context.Context, labID int64, names []string) ([]*model.MaterialNode, error) {
	if len(names) == 0 {
		return []*model.MaterialNode{}, nil
	}

	// 构建参数
	params := []any{labID, names[0]} // lab_id, 根节点名称
	// 动态构建 SQL - 修正递归逻辑
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString(`WITH RECURSIVE node_path AS (
    -- 查找根节点 liquid_handler
    SELECT *, 1 AS depth, ARRAY[name]::character varying[] AS path_names
    FROM material_node
    WHERE lab_id = ? AND name = ? AND parent_id = 0

    UNION ALL

    -- 递归部分 - 在只有一个节点时不会执行
    SELECT mn.*, np.depth + 1 AS depth, np.path_names || mn.name
    FROM material_node mn
    JOIN node_path np ON mn.parent_id = np.id AND mn.lab_id = np.lab_id`)

	if len(names) == 1 {
		sqlBuilder.WriteString(`
			WHERE FALSE  -- 限制只查询第一层
    )
    SELECT * FROM node_path
    ORDER BY depth ASC;`)
	} else {
		sqlBuilder.WriteString(`
        WHERE np.depth < ?  -- 最大深度为数组长度
    )
    SELECT * FROM node_path
    WHERE CASE`)

		params = append(params, len(names))
		for index, name := range names {
			sqlBuilder.WriteString(fmt.Sprintf(`
				WHEN depth = %d THEN name = ?`, index+1))
			params = append(params, name)
		}
		sqlBuilder.WriteString(`
		    ELSE FALSE
		END
    ORDER BY depth ASC;`)
	}

	var results []*model.MaterialNode
	if err := m.DBWithContext(ctx).Raw(sqlBuilder.String(), params...).Scan(&results).Error; err != nil {
		logger.Errorf(ctx, "GetMaterialNodeByPath fail lab id: %d, names: %+v, err: %+v", labID, names, err)
		return nil, code.QueryRecordErr.WithErr(err)
	}

	return results, nil
}

func (m *materialImpl) GetDescendants(ctx context.Context, labID int64, nodeID int64) ([]*model.MaterialNode, error) {
	sql := `
WITH RECURSIVE descendants AS (
    SELECT *, 1 AS depth
    FROM material_node
    WHERE parent_id = ? AND lab_id = ?

    UNION ALL

    SELECT m.*, d.depth + 1
    FROM material_node m
    JOIN descendants d ON m.parent_id = d.id AND m.lab_id = d.lab_id
)
SELECT * FROM descendants
ORDER BY depth, id;
	`
	var results []*model.MaterialNode
	if err := m.DBWithContext(ctx).Raw(sql, nodeID, labID).Scan(&results).Error; err != nil {
		logger.Errorf(ctx, "GetDescendants fail  node id: %d, err: %+v", nodeID, err)
		return nil, err
	}
	return results, nil
}

// FIXME: 优化效率
func (m *materialImpl) UpdateMaterialNodeDataKey(ctx context.Context, labID int64, deviceName string, key string, value any) ([]*model.MaterialNode, error) {
	updatedNodes := make([]*model.MaterialNode, 0, 1)
	if err := m.DBWithContext(ctx).Where("lab_id = ? and name = ?", labID, deviceName).Select("id", "uuid", "data").Find(&updatedNodes).Error; err != nil {
		logger.Errorf(ctx, "UpdateMaterialNodeDataKey find lab id: %d, name: %s, err: %+v", labID, deviceName, err)
		return nil, nil
	}
	if len(updatedNodes) == 0 {
		return nil, nil
	}

	updatedNodes, err := utils.FilterSliceErr(updatedNodes, func(n *model.MaterialNode) (*model.MaterialNode, bool, error) {
		jsonStr, err := sjson.Set(string(n.Data), key, value)
		if err != nil {
			logger.Errorf(ctx, "UpdateMaterialNodeDataKey sjson lab id: %d, name: %s, err: %+v", labID, deviceName, err)
			return nil, false, code.UpdateNodeErr
		}

		n.Data = datatypes.JSON(jsonStr)
		return n, true, nil
	})
	if err != nil {
		return nil, code.UpdateDataErr
	}

	statement := m.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(updatedNodes)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpdateMaterialNodeDataKey update err: %+v", statement.Error)
		return nil, code.UpdateDataErr
	}

	return updatedNodes, nil
}

func (m *materialImpl) GetAncestors(ctx context.Context, nodeUUID uuid.UUID) ([]*model.MaterialNode, error) {
	sql := `
WITH RECURSIVE ancestors AS (
    -- 锚点：选择起始子节点本身
    SELECT *, 0 AS level
    FROM material_node
    WHERE uuid = ?  -- 这里通过子节点的ID来定位

    UNION ALL

    -- 递归部分：通过parent_id向上查找父节点
    SELECT m.*, a.level + 1
    FROM material_node m
    JOIN ancestors a ON m.id = a.parent_id
    -- 通常还需要确保在同一个实验室内，假设parent_id在同一lab内是有效的
    AND m.lab_id = a.lab_id
)
SELECT * FROM ancestors
ORDER BY level DESC; -- 按level降序排列，从根节点到子节点
	`
	var results []*model.MaterialNode
	if err := m.DBWithContext(ctx).Raw(sql, nodeUUID).Scan(&results).Error; err != nil {
		logger.Errorf(ctx, "GetAncestors fail node uuid: %s, err: %+v", nodeUUID.String(), err)
		return nil, err
	}
	return results, nil
}
