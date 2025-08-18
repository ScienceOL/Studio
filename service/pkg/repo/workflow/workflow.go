package workflow

import (
	"context"
	"errors"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type workflowImpl struct {
	repo.IDOrUUIDTranslate
}

func New() repo.WorkflowRepo {
	return &workflowImpl{
		IDOrUUIDTranslate: repo.NewBaseDB(),
	}
}

func (w *workflowImpl) Create(ctx context.Context, data *model.Workflow) error {
	if statement := w.DBWithContext(ctx).Create(data); statement.Error != nil {
		logger.Errorf(ctx, "Create fail err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}

func (w *workflowImpl) CreateNode(ctx context.Context, data *model.WorkflowNode) error {
	if statement := w.DBWithContext(ctx).Create(data); statement.Error != nil {
		logger.Errorf(ctx, "CreateNode fail err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}

func (w *workflowImpl) GetWorkflowByUUID(ctx context.Context, uuid uuid.UUID) (*model.Workflow, error) {
	data := &model.Workflow{}
	if err := w.DBWithContext(ctx).Where("uuid = ?", uuid).Take(data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.RecordNotFound
		}
		logger.Errorf(ctx, "GetWorkflowByUUID fail uuid: %+v, error: %+v", uuid, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return data, nil
}

func (w *workflowImpl) IsExist(ctx context.Context, uuid uuid.UUID) (bool, error) {
	data := &model.Workflow{}
	if err := w.DBWithContext(ctx).Select("id").Where("uuid = ?", uuid).Take(data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logger.Errorf(ctx, "IsExist fail uuid: %+v, error: %+v", uuid, err)
		return false, code.QueryRecordErr.WithMsg(err.Error())
	}

	return true, nil
}

func (w *workflowImpl) GetWorkflowGraph(ctx context.Context, userID string, workflowUUID uuid.UUID) (*repo.WorkflowGrpah, error) {
	workflow := &model.Workflow{}
	if err := w.DBWithContext(ctx).
		Where("uuid = ? and user_id = ?", workflowUUID, userID).
		Take(workflow).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowGraph get workflow fail err: %+v", err)
		return nil, code.QueryRecordErr
	}

	nodes := make([]*model.WorkflowNode, 0, 1)
	if err := w.DBWithContext(ctx).
		Where("workflow_id =  ?", workflow.ID).
		Find(&nodes).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowGraph get node fail err: %+v", err)
		return nil, code.QueryRecordErr
	}

	nodeUUIDs := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		return node.UUID, true
	})

	templateIDs := make([]int64, 0, len(nodes))
	for _, node := range nodes {
		templateIDs = utils.AppendUniqSlice(templateIDs, node.TemplateID)
	}

	tplNodes := make([]*model.WorkflowNodeTemplate, 0, len(templateIDs))
	if err := w.DBWithContext(ctx).
		Select("id, uuid, schema, header, footer").
		Where("id in ?", templateIDs).
		Find(&tplNodes).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowGraph get template fail tpl ids: %+v, err: %+v", templateIDs, err)

		return nil, code.QueryRecordErr
	}
	tplNodeMap := utils.SliceToMap(tplNodes, func(tpl *model.WorkflowNodeTemplate) (int64, *model.WorkflowNodeTemplate) {
		return tpl.ID, tpl
	})

	tplHandles := make([]*model.WorkflowHandleTemplate, 0, len(templateIDs))
	if err := w.DBWithContext(ctx).
		Where("node_template_id in ?", templateIDs).
		Find(&tplHandles).Error; err != nil {

		logger.Errorf(ctx, "GetWorkflowGraph get template handle fail template ids: %+v, err: %+v", templateIDs, err)
		return nil, code.QueryRecordErr
	}

	mapTplHandle := make(map[int64][]*model.WorkflowHandleTemplate, len(tplHandles))
	for _, tplHandle := range tplHandles {
		mapTplHandle[tplHandle.NodeTemplateID] = append(mapTplHandle[tplHandle.NodeTemplateID], tplHandle)
	}

	nodesInfo := make([]*repo.WorkflowNodeInfo, 0, len(nodes))
	for _, node := range nodes {
		nodesInfo = append(nodesInfo, &repo.WorkflowNodeInfo{
			Node:     node,
			Handles:  mapTplHandle[node.TemplateID],
			Template: tplNodeMap[node.TemplateID],
		})
	}

	edges := make([]*model.WorkflowEdge, 0, 1)
	if err := w.DBWithContext(ctx).
		Where("source_node_uuid in ? or target_node_uuid in ?", nodeUUIDs, nodeUUIDs).
		Find(&edges).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowGraph get edge fail err: %+v", err)

		return nil, code.QueryRecordErr
	}

	return &repo.WorkflowGrpah{
		Workflow: workflow,
		Nodes:    nodesInfo,
		Edges:    edges,
	}, nil
}

func (w *workflowImpl) GetWorkflowTemplateByUUID(ctx context.Context, tplUUID uuid.UUID) (*repo.WorkflowTemplate, error) {
	tplData := &model.WorkflowNodeTemplate{}
	if err := w.DBWithContext(ctx).
		Where("uuid = ?", tplUUID).
		Take(tplData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetWorkflowTemplateByUUID fail uuid id: %v, err: %+v", tplUUID, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	tplHandleDatas := make([]*model.WorkflowHandleTemplate, 0, 2)
	if err := w.DBWithContext(ctx).
		Where("node_template_id = ?", tplData.ID).
		Find(&tplHandleDatas).Error; err != nil {

		logger.Errorf(ctx, "GetWorkflowTemplateByUUID fail uuid id: %v, err: %+v", tplUUID, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return &repo.WorkflowTemplate{
		Template: tplData,
		Handles:  tplHandleDatas,
	}, nil
}

func (w *workflowImpl) GetWorkflowTemplate(ctx context.Context, labID int64) ([]*repo.WorkflowTemplate, error) {
	tpls := make([]*model.WorkflowNodeTemplate, 0, 1)

	query := w.DBWithContext(ctx).Where("lab_id = ?", labID)
	if err := query.Order("id desc").Find(&tpls).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowTemplate fail lab id: %d, err: %+v", labID, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	tplsIDs := utils.FilterSlice(tpls, func(tpl *model.WorkflowNodeTemplate) (int64, bool) {
		return tpl.ID, true
	})

	handles := make([]*model.WorkflowHandleTemplate, 0, 1)
	if err := w.DBWithContext(ctx).
		Where("node_template_id in ?", tplsIDs).
		Find(&handles).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowTemplate get handle fail lab id: %d, err: %+v", labID, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	handlesMap := utils.SliceToMapSlice(handles, func(h *model.WorkflowHandleTemplate) (
		int64, *model.WorkflowHandleTemplate, bool) {
		return h.NodeTemplateID, h, true
	})

	return utils.FilterSlice(tpls, func(tpl *model.WorkflowNodeTemplate) (*repo.WorkflowTemplate, bool) {
		return &repo.WorkflowTemplate{
			Template: tpl,
			Handles:  handlesMap[tpl.ID],
		}, true
	}), nil
}

func (w *workflowImpl) GetWorkflowNode(ctx context.Context, uuid uuid.UUID) (*model.WorkflowNode, error) {
	data := &model.WorkflowNode{}
	if err := w.DBWithContext(ctx).
		Where("uuid = ?", uuid).
		Take(data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetWorkflowNode fail uuid id: %v, err: %+v", uuid, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return data, nil
}

func (w *workflowImpl) UpdateWorkflowNode(ctx context.Context, workflowUUID uuid.UUID, data *model.WorkflowNode, updateColumns []string) error {
	if err := w.DBWithContext(ctx).Where("uuid = ?", workflowUUID).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "uuid"},
			},
			DoUpdates: clause.AssignmentColumns(append(updateColumns, "updated_at")),
		}).Create(data).Error; err != nil {
		logger.Errorf(ctx, "UpdateWorkflowNode fail uuid:%v, err: %+v", workflowUUID, err)
		return code.UpdateDataErr
	}

	return nil
}

func (w *workflowImpl) DeleteWorkflowNodes(ctx context.Context, workflowUUIDs []uuid.UUID) (*repo.DeleteWorkflow, error) {
	if len(workflowUUIDs) == 0 {
		return &repo.DeleteWorkflow{}, nil
	}

	resp := &repo.DeleteWorkflow{}
	if err := w.ExecTx(ctx, func(txCtx context.Context) error {
		edgeUUIDs := make([]uuid.UUID, 0, len(workflowUUIDs))
		if err := w.DBWithContext(txCtx).Model(&model.WorkflowEdge{}).
			Select("uuid").
			Where("source_node_uuid in ? or target_node_uuid in ?", workflowUUIDs, workflowUUIDs).
			Find(&edgeUUIDs); err != nil {

			logger.Errorf(txCtx, "DeleteWorkflowNodes get workflow edge uuid fail uuid: %+v, err: %+v", workflowUUIDs, err)
			return code.QueryRecordErr
		}

		// 删除工作流，删除边

		if err := w.DBWithContext(ctx).Where("uuid in ?", workflowUUIDs).Delete(&model.WorkflowNode{}).Error; err != nil {
			logger.Errorf(ctx, "DeleteWorkflowNodes fail uuid: %+v, err: %+v", edgeUUIDs, err)
			return code.DeleteDataErr
		}

		if len(edgeUUIDs) > 0 {
			if err := w.DBWithContext(ctx).Where("uuid in ?", edgeUUIDs).Delete(&model.WorkflowEdge{}).Error; err != nil {
				logger.Errorf(ctx, "DeleteWorkflowNodes fail uuid: %+v, err: %+v", edgeUUIDs, err)
				return code.DeleteDataErr
			}
		}

		resp.EdgeUUIDs = edgeUUIDs
		resp.EdgeUUIDs = workflowUUIDs
		return nil
	}); err != nil {
		logger.Errorf(ctx, "DeleteWorkflowNodes fail uuids: %+v, err: %+v", workflowUUIDs, err)
		return nil, err
	}

	return resp, nil
}

func (w *workflowImpl) DeleteWorkflowEdges(ctx context.Context, edgeUUIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(edgeUUIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	if err := w.DBWithContext(ctx).Where("uuid in ?", edgeUUIDs).Delete(&model.WorkflowEdge{}).Error; err != nil {
		logger.Errorf(ctx, "DeleteWorkflowEdges fail uuid: %+v, err: %+v", edgeUUIDs, err)
		return nil, code.DeleteDataErr
	}

	return edgeUUIDs, nil
}

func (w *workflowImpl) UpsertWorkflowEdge(ctx context.Context, datas []*model.WorkflowEdge) error {
	if len(datas) == 0 {
		return nil
	}

	statement := w.DBWithContext(ctx).Clauses(clause.OnConflict{
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
		logger.Errorf(ctx, "UpsertWorkflowEdge err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}

func (w *workflowImpl) GetWorkflowTemplatePage(ctx context.Context, labUUID uuid.UUID, page *common.PageReq) (*common.PageResp[*repo.WorkflowTemplate], error) {
	return nil, nil
	// tpls := make([]*model.WorkflowNodeTemplate, 0, 1)
	// total := int64(0)
	// query := w.DBWithContext(ctx).
	// 	Where("lab_uuid = ?", labUUID).
	// 	Offset(page.Offest()).
	// 	Limit(page.PageSize)
	//
	// if err := query.Order("id desc").Find(&tpls).Error; err != nil {
	// 	logger.Errorf(ctx, "GetWorkflowTemplatePage fail lab uuid: %d, err: %+v", labUUID, err)
	// 	return nil, code.QueryRecordErr.WithMsg(err.Error())
	// }
	//
	// tplsIDs := utils.FilterSlice(tpls, func(tpl *model.WorkflowNodeTemplate) (int64, bool) {
	// 	return tpl.ID, true
	// })
	//
	// handles := make([]*model.WorkflowHandleTemplate, 0, 1)
	// if err := w.DBWithContext(ctx).
	// 	Where("node_template_id in ?", tplsIDs).
	// 	Find(&handles).Error; err != nil {
	// 	logger.Errorf(ctx, "GetWorkflowTemplate get handle fail lab id: %d, err: %+v", labID, err)
	// 	return nil, code.QueryRecordErr.WithMsg(err.Error())
	// }
	//
	// handlesMap := utils.SliceToMapSlice(handles, func(h *model.WorkflowHandleTemplate) (
	// 	int64, *model.WorkflowHandleTemplate, bool) {
	// 	return h.NodeTemplateID, h, true
	// })
	//
	// return utils.FilterSlice(tpls, func(tpl *model.WorkflowNodeTemplate) (*repo.WorkflowTemplate, bool) {
	// 	return &repo.WorkflowTemplate{
	// 		Template: tpl,
	// 		Handles:  handlesMap[tpl.ID],
	// 	}, true
	// }), nil
}

// GetWorkflowList 获取工作流列表
func (w *workflowImpl) GetWorkflowList(ctx context.Context, userID string, labID int64, page *common.PageReq) ([]*model.Workflow, int64, error) {
	workflows := make([]*model.Workflow, 0, 1)
	total := int64(0)

	// 构建查询条件
	query := w.DBWithContext(ctx).Model(&model.Workflow{})

	// 如果指定了实验室ID，则按实验室过滤
	if labID > 0 {
		query = query.Where("lab_id = ?", labID)
	}

	// 按用户ID过滤
	query = query.Where("user_id = ?", userID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowList count fail user_id: %s, lab_id: %d, err: %+v", userID, labID, err)
		return nil, 0, code.QueryRecordErr.WithMsg(err.Error())
	}

	// 分页查询
	if err := query.Offset(page.Offest()).
		Limit(page.PageSize).
		Order("created_at desc").
		Find(&workflows).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowList query fail user_id: %s, lab_id: %d, err: %+v", userID, labID, err)
		return nil, 0, code.QueryRecordErr.WithMsg(err.Error())
	}

	return workflows, total, nil
}
