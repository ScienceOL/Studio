package workflow

import (
	"context"
	"errors"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/gorm"
)

type workflowImpl struct {
	*db.Datastore
}

func New() repo.WorkflowRepo {
	return &workflowImpl{
		Datastore: db.DB(),
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
		Where("id in ?", templateIDs).
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
		Where("source_node_uuid in ?", nodeUUIDs).
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
	if err := w.DBWithContext(ctx).
		Where("lab_id = ?", labID).
		Find(&tpls).Error; err != nil {
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
