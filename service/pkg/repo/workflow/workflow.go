package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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
			return nil, code.RecordNotFound.WithMsgf("uuid: %s", uuid)
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
		return nil, code.QueryRecordErr.WithMsg("get node fail")
	}

	workflowIDs := utils.FilterUniqSlice(nodes, func(node *model.WorkflowNode) (int64, bool) {
		if node.WorkflowNodeID == 0 {
			return 0, false
		}

		return node.WorkflowNodeID, true
	})

	actions := make([]*model.WorkflowNodeTemplate, 0, 1)
	if err := w.DBWithContext(ctx).
		Where("id in  ?", workflowIDs).
		Find(&actions).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowGraph get node fail err: %+v", err)
		return nil, code.QueryRecordErr.WithMsg("get node fail")
	}
	actionMap := utils.Slice2Map(actions, func(action *model.WorkflowNodeTemplate) (int64, *model.WorkflowNodeTemplate) {
		return action.ID, action
	})

	actionHandles := make([]*model.WorkflowHandleTemplate, 0, 1)
	if err := w.DBWithContext(ctx).
		Where("workflow_node_id in  ?", workflowIDs).
		Find(&actionHandles).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowGraph get action handles fail err: %+v", err)
		return nil, code.QueryRecordErr.WithMsg("get action handles")
	}

	actionHandleMap := utils.SliceToMapSlice(actionHandles, func(item *model.WorkflowHandleTemplate) (int64, *model.WorkflowHandleTemplate, bool) {
		return item.WorkflowNodeID, item, true
	})

	nodeUUIDs := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		return node.UUID, true
	})

	nodesInfo := make([]*repo.WorkflowNodeInfo, 0, len(nodes))
	for _, node := range nodes {
		nodesInfo = append(nodesInfo, &repo.WorkflowNodeInfo{
			Node:    node,
			Action:  actionMap[node.WorkflowNodeID],
			Handles: actionHandleMap[node.WorkflowNodeID],
		})
	}

	edges := make([]*model.WorkflowEdge, 0, 1)
	if err := w.DBWithContext(ctx).
		Where("source_node_uuid in ? or target_node_uuid in ?", nodeUUIDs, nodeUUIDs).
		Find(&edges).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowGraph get edge fail err: %+v", err)

		return nil, code.QueryRecordErr.WithMsg("get edge fail")
	}

	return &repo.WorkflowGrpah{
		Workflow: workflow,
		Nodes:    nodesInfo,
		Edges:    edges,
	}, nil
}

// func (w *workflowImpl) GetWorkflowTemplateByUUID(ctx context.Context, tplUUID uuid.UUID) (*repo.WorkflowTemplate, error) {
// 	tplData := &model.WorkflowNodeTemplate{}
// 	if err := w.DBWithContext(ctx).
// 		Where("uuid = ?", tplUUID).
// 		Take(tplData).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, code.RecordNotFound
// 		}
//
// 		logger.Errorf(ctx, "GetWorkflowTemplateByUUID fail uuid id: %v, err: %+v", tplUUID, err)
// 		return nil, code.QueryRecordErr.WithMsg(err.Error())
// 	}
// 	return nil, nil
//
// 	// tplHandleDatas := make([]*model.WorkflowHandleTemplate, 0, 2)
// 	// if err := w.DBWithContext(ctx).
// 	// 	Where("node_template_id = ?", tplData.ID).
// 	// 	Find(&tplHandleDatas).Error; err != nil {
// 	//
// 	// 	logger.Errorf(ctx, "GetWorkflowTemplateByUUID fail uuid id: %v, err: %+v", tplUUID, err)
// 	// 	return nil, code.QueryRecordErr.WithMsg(err.Error())
// 	// }
// 	//
// 	// return &repo.WorkflowTemplate{
// 	// 	Template: tplData,
// 	// 	Handles:  tplHandleDatas,
// 	// }, nil
// }

func (w *workflowImpl) GetWorkflowNodeTemplate(ctx context.Context, condition map[string]any) ([]*model.WorkflowNodeTemplate, error) {
	workflowNodeTpls := make([]*model.WorkflowNodeTemplate, 0, 1)

	query := w.DBWithContext(ctx).Where(condition)
	if err := query.Order("id desc").Find(&workflowNodeTpls).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowNodeTemplate fail lab id: %v, err: %+v", condition, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return workflowNodeTpls, nil
}

func (w *workflowImpl) GetWorkflowHandleTemplates(ctx context.Context, wfTemaplteIDs []int64) ([]*model.WorkflowHandleTemplate, error) {
	handles := make([]*model.WorkflowHandleTemplate, 0, 1)
	query := w.DBWithContext(ctx).Where("workflow_node_id in ?", wfTemaplteIDs)
	if err := query.Order("id desc").Find(&handles).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowHandleTemaplates fail action ids: %v, err: %+v", wfTemaplteIDs, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return handles, nil
}

func (w *workflowImpl) GetWorkflowNodes(ctx context.Context, condition map[string]any) ([]*model.WorkflowNode, error) {
	data := make([]*model.WorkflowNode, 0, 1)
	if err := w.DBWithContext(ctx).
		Where(condition).
		Find(&data).Error; err != nil {

		logger.Errorf(ctx, "GetWorkflowNode fail condition: %v, err: %+v", condition, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return data, nil
}

func (w *workflowImpl) GetWorkflowEdges(ctx context.Context, nodeUUIDs []uuid.UUID) ([]*model.WorkflowEdge, error) {
	data := make([]*model.WorkflowEdge, 0, 1)
	if err := w.DBWithContext(ctx).
		Where("source_node_uuid in ? or target_node_uuid in ?", nodeUUIDs, nodeUUIDs).
		Find(&data).Error; err != nil {

		logger.Errorf(ctx, "GetWorkflowNode fail uuids: %v, err: %+v", nodeUUIDs, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return data, nil
}

func (w *workflowImpl) UpdateWorkflowNode(ctx context.Context, nodeUUID uuid.UUID, data *model.WorkflowNode, updateColumns []string) error {
	data.UUID = nodeUUID
	if err := w.DBWithContext(ctx).
		Where("uuid = ?", nodeUUID).
		Select(append(updateColumns, "updated_at")).
		Updates(data).Error; err != nil {
		logger.Errorf(ctx, "UpdateWorkflowNode fail uuid:%v, err: %+v", nodeUUID, err)
		return code.UpdateDataErr.WithErr(err)
	}

	return nil
}

func (w *workflowImpl) UpdateWorkflowNodes(ctx context.Context, nodeUUIDs []uuid.UUID, data *model.WorkflowNode, updateColumns []string) error {
	if err := w.DBWithContext(ctx).
		Where("uuid in ?", nodeUUIDs).
		Select(append(updateColumns, "updated_at")).
		Updates(data).Error; err != nil {
		logger.Errorf(ctx, "UpdateWorkflowNodes fail uuid:%v, err: %+v", nodeUUIDs, err)

		return code.UpdateDataErr.WithErr(err)
	}

	return nil
}

func (w *workflowImpl) DeleteWorkflowNodes(ctx context.Context, workflowUUIDs []uuid.UUID) (*repo.DeleteWorkflow, error) {
	if len(workflowUUIDs) == 0 {
		return &repo.DeleteWorkflow{}, nil
	}

	resp := &repo.DeleteWorkflow{}
	if err := w.ExecTx(ctx, func(txCtx context.Context) error {
		nodes := make([]*model.WorkflowNode, 0, len(workflowUUIDs))
		if err := w.DBWithContext(txCtx).Clauses(
			clause.Returning{
				Columns: []clause.Column{
					{Name: "id"},
					{Name: "uuid"},
					{Name: "type"},
				},
			}).
			Where("uuid in ?", workflowUUIDs).
			Delete(&nodes).Error; err != nil {

			logger.Errorf(ctx, "DeleteWorkflowGroupNodes fail uuids: %+v, err: %+v", workflowUUIDs, err)
			return code.DeleteDataErr.WithErr(err)
		}

		normalUUIDs := make([]uuid.UUID, 0, len(nodes))
		parentIDs := utils.FilterSlice(nodes, func(item *model.WorkflowNode) (int64, bool) {
			if item.Type == model.WorkflowNodeGroup {
				return item.ID, true
			}
			normalUUIDs = append(normalUUIDs, item.UUID)
			return item.ID, false
		})

		if len(parentIDs) > 0 {
			updateData := &model.WorkflowNode{}
			updateData.UpdatedAt = time.Now()
			if err := w.DBWithContext(txCtx).
				Where("parent_id in ?", parentIDs).
				Select("parent_id", "updated_at").
				Updates(updateData).Error; err != nil {
				logger.Errorf(ctx, "DeleteWorkflowGroupNodes set parent node fail parent ids: %+v, err: %+v", parentIDs, err)
				return code.DeleteDataErr.WithErr(err)
			}
		}

		if len(normalUUIDs) > 0 {
			edges := make([]*model.WorkflowEdge, 0, len(workflowUUIDs))
			if err := w.DBWithContext(txCtx).Clauses(
				clause.Returning{
					Columns: []clause.Column{
						{Name: "uuid"},
					},
				}).
				Where("source_node_uuid in ? or target_node_uuid in ?", normalUUIDs, normalUUIDs).
				Delete(&edges).Error; err != nil {

				logger.Errorf(ctx, "DeleteWorkflowGroupNodes fail uuids: %+v, err: %+v", workflowUUIDs, err)
				return code.DeleteDataErr.WithErr(err)
			}
			resp.EdgesUUIDs = utils.FilterSlice(edges, func(edge *model.WorkflowEdge) (uuid.UUID, bool) {
				return edge.UUID, true
			})
		}

		resp.NodeUUIDs = workflowUUIDs
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

// GetTemplateList 获取模板列表（分页）
func (w *workflowImpl) GetTemplateList(ctx context.Context,
	req *common.PageReqT[*repo.QueryTemplage]) (*common.PageResp[[]*model.WorkflowNodeTemplate],
	error,
) {
	if req.Data.LabID == 0 {
		return nil, code.QueryRecordErr.WithMsg("lab id is zero")
	}

	templates := make([]*model.WorkflowNodeTemplate, 0, 1)
	total := int64(0)

	// 构建查询条件
	query := w.DBWithContext(ctx).Model(&model.WorkflowNodeTemplate{})

	// 按实验室ID过滤
	query = query.Where("lab_id = ?", req.Data.LabID)
	if req.Data.Name != "" {
		query = query.Where("name like ?", "%"+req.Data.Name+"%")
	}

	if len(req.Data.ResourceNodeIDs) > 0 {
		query = query.Where("resource_node_id in ?", req.Data.ResourceNodeIDs)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf(ctx, "GetTemplateList count fail query: %+v, err: %+v", *req, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	// 分页查询
	if err := query.Offset(req.Offest()).
		Limit(req.PageSize).
		Order("created_at desc").
		Find(&templates).Error; err != nil {
		logger.Errorf(ctx, "GetTemplateList query fail query: %+v, err: %+v", *req, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return &common.PageResp[[]*model.WorkflowNodeTemplate]{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     templates,
	}, nil
}

// GetNodeTemplateByUUID 根据UUID获取节点模板详情
func (w *workflowImpl) GetNodeTemplateByUUID(ctx context.Context, templateUUID uuid.UUID) (*model.WorkflowNodeTemplate, error) {
	template := &model.WorkflowNodeTemplate{}

	if err := w.DBWithContext(ctx).
		Where("uuid = ?", templateUUID).
		First(template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf(ctx, "GetNodeTemplateByUUID record not found uuid: %+v", templateUUID)
			return nil, code.RecordNotFound.WithMsgf("template not found: %s", templateUUID)
		}
		logger.Errorf(ctx, "GetNodeTemplateByUUID fail uuid: %+v, err: %+v", templateUUID, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return template, nil
}

func (w *workflowImpl) UpsertNodes(ctx context.Context, nodes []*model.WorkflowNode) error {
	if len(nodes) == 0 {
		return nil
	}

	if err := w.DBWithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{
					Name: "uuid",
				},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"icon",
				"pose",
				"param",
				"footer",
				"device_name",
				"disabled",
				"minimized",
				"updated_at",
			}),
		}).
		Create(&nodes).Error; err != nil {

		logger.Errorf(ctx, "UpsertNodes fail, err: %+v", err)
		return code.UpdateDataErr.WithErr(err)
	}

	return nil
}

func (w *workflowImpl) DuplicateEdge(ctx context.Context, edges []*model.WorkflowEdge) error {
	if len(edges) == 0 {
		return nil
	}

	if err := w.DBWithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{
					Name: "source_node_uuid",
				},
				{
					Name: "target_node_uuid",
				},
				{
					Name: "source_handle_uuid",
				},
				{
					Name: "target_handle_uuid",
				},
			},
			DoNothing: true,
		}).Create(edges).Error; err != nil {

		logger.Errorf(ctx, "UpsertEdge fail err: %+v", err)
		return code.UpdateDataErr.WithErr(err)
	}

	return nil
}

func (w *workflowImpl) UpsertEdge(ctx context.Context, edges []*model.WorkflowEdge) error {
	if len(edges) == 0 {
		return nil
	}

	if err := w.DBWithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{
					Name: "uuid",
				},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at",
			}),
		}).Create(edges).Error; err != nil {

		logger.Errorf(ctx, "UpsertEdge fail err: %+v", err)
		return code.UpdateDataErr.WithErr(err)
	}

	return nil
}

func (w *workflowImpl) CreateJobs(ctx context.Context, datas []*model.WorkflowNodeJob) error {
	if statement := w.DBWithContext(ctx).Create(datas); statement.Error != nil {
		logger.Errorf(ctx, "CreateJobs fail err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}

func (w *workflowImpl) UpsertJobs(ctx context.Context, datas []*model.WorkflowNodeJob) error {
	if len(datas) == 0 {
		return nil
	}

	if err := w.DBWithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{
					Name: "uuid",
				},
			},

			DoUpdates: clause.AssignmentColumns([]string{
				"status",
				"data",
				"updated_at",
			}),
		}).Create(datas).Error; err != nil {

		logger.Errorf(ctx, "UpsertEdge fail err: %+v", err)
		return code.UpdateDataErr.WithErr(err)
	}

	return nil
}

func (w *workflowImpl) CreateWorkflowTask(ctx context.Context, data *model.WorkflowTask) error {
	if data == nil {
		return code.CreateDataErr.WithMsg("data is nil")
	}

	if err := w.DBWithContext(ctx).Create(data).Error; err != nil {
		logger.Errorf(ctx, "CreateWorkflowTask fail err: %+v", err)
		return code.CreateDataErr.WithErr(err)
	}

	return nil
}

func (w *workflowImpl) GetWorkflowTasks(ctx context.Context,
	req *common.PageReqT[*repo.TaskReq]) (*common.PageMoreResp[[]*model.WorkflowTask],
	error,
) {
	tasks := make([]*model.WorkflowTask, 0, 1)
	total := int64(0)

	// 构建查询条件
	query := w.DBWithContext(ctx).Model(&model.WorkflowTask{})

	// 如果指定了实验室ID，则按实验室过滤
	query = query.Where("lab_id = ?", req.Data.LabID)

	// 按用户ID过滤
	query = query.Where("user_id = ?", req.Data.UserID)

	// 工作流
	query = query.Where("workflow_id = ?", req.Data.WrokflowID)

	req.Normalize()

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowTasks count fail param: %+v, err: %+v", req, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	// 分页查询
	if err := query.Offset(req.Offest()).
		Limit(req.PageSize).
		Order("created_at desc").
		Find(&tasks).Error; err != nil {
		logger.Errorf(ctx, "GetWorkflowTasks query fail param: %+v, err: %+v", req, err)
		return nil, code.QueryRecordErr.WithMsg(err.Error())
	}

	return &common.PageMoreResp[[]*model.WorkflowTask]{
		HasMore:  total > int64(req.Page+1)*int64(req.PageSize),
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     tasks,
	}, nil
}

func (w *workflowImpl) DelWorkflow(ctx context.Context, workflowID int64) error {
	if workflowID == 0 {
		return code.WorkflowNotExistErr
	}

	return w.ExecTx(ctx, func(txCtx context.Context) error {
		nodes := make([]*model.WorkflowNode, 0, 1)
		if err := w.DBWithContext(txCtx).Clauses(
			clause.Returning{
				Columns: []clause.Column{
					{Name: "id"},
					{Name: "uuid"},
					{Name: "type"},
				},
			}).
			Where("workflow_id in ?").
			Delete(&nodes).Error; err != nil {

			logger.Errorf(ctx, "DeleteWorkflow fail id: %d, err: %+v", workflowID, err)
			return code.DeleteDataErr.WithErr(err)
		}

		nodeUUIDs := utils.FilterSlice(nodes, func(n *model.WorkflowNode) (uuid.UUID, bool) {
			return n.UUID, true
		})

		return w.DBWithContext(txCtx).
			Where("source_node_uuid in ? or target_node_uuid in ?", nodeUUIDs, nodeUUIDs).
			Delete(&model.WorkflowEdge{}).Error
	})
}

func (w *workflowImpl) GetWorkflow(ctx context.Context, req *common.PageReqT[*repo.QueryWorkflow]) (*common.PageResp[[]*model.Workflow], error) {
	/*
		tagsToFind := []string{"AI", "ML"}

		// 方法 1: 使用 Where 条件 包含其中一个
		db.Where("tags ?| array[?]", tagsToFind).Find(&workflows)



		tagsToFind := []string{"AI", "ML"}

		// 将标签数组转换为 JSON 字符串 包含所有
		tagsJSON, _ := json.Marshal(tagsToFind)

		// 方法 1: 使用 Where 条件
		db.Where("tags @> ?", string(tagsJSON)).Find(&workflows)
	*/

	query := w.DBWithContext(ctx).
		Model(&model.Workflow{}).
		Where("published = true")
	if len(req.Data.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Data.Tags)
		query.Where("tags @> ?", tagsJSON)
	}

	var count int64
	var datas []*model.Workflow
	if err := query.Count(&count).
		Offset(req.Offest()).
		Limit(req.PageSize).Select("uuid", "name", "created_at", "user_id").
		Find(&datas).Error; err != nil {
		return nil, err
	}

	return &common.PageResp[[]*model.Workflow]{
		Total:    count,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     datas,
	}, nil
}

func (w *workflowImpl) GetTemplateTags(ctx context.Context, tagType model.TagType) ([]string, error) {
	var tags []string
	if err := w.DBWithContext(ctx).
		Model(&model.Tags{}).
		Where("type = ?", tagType).
		Select("name").
		Find(&tags).Error; err != nil {
		return nil, code.QueryRecordErr.WithErr(err)
	}

	return tags, nil
}
