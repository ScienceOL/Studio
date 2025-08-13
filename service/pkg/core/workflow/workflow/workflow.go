package workflow

import (
	"context"
	"encoding/json"

	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/workflow"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/repo"
	el "github.com/scienceol/studio/service/pkg/repo/environment"
	mStore "github.com/scienceol/studio/service/pkg/repo/material"
	"github.com/scienceol/studio/service/pkg/repo/model"
	wfl "github.com/scienceol/studio/service/pkg/repo/workflow"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/datatypes"
)

type workflowImpl struct {
	workflowStore repo.WorkflowRepo
	labStore      repo.LaboratoryRepo
	materialStore repo.MaterialRepo
}

func New() workflow.Service {
	return &workflowImpl{
		workflowStore: wfl.New(),
		labStore:      el.New(),
		materialStore: mStore.NewMaterialImpl(),
	}
}

func (w *workflowImpl) Add(ctx context.Context, data *workflow.WorkflowReq) (*workflow.WorkflowResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}
	lab, err := w.labStore.GetLabByUUID(ctx, data.LabUUID)
	if err != nil {
		return nil, err
	}

	d := &model.Workflow{
		UserID:      userInfo.ID,
		Name:        utils.Or(data.Name, "Untitled"),
		Description: data.Description,
		LabID:       lab.ID,
	}

	err = w.workflowStore.Create(ctx, d)
	if err != nil {
		return nil, err
	}

	return &workflow.WorkflowResp{
		UUID:        d.UUID,
		Name:        d.Name,
		Description: d.Description,
	}, nil
}

func (w *workflowImpl) NodeTemplateList(ctx context.Context) {
	// TODO: 未实现
}

func (w *workflowImpl) ForkTemplate(ctx context.Context) {
	// TODO: 未实现
}

func (w *workflowImpl) NodeTemplateDetail(ctx context.Context) {
	// TODO: 未实现
}

func (w *workflowImpl) TemplateDetail(ctx context.Context) {
	// TODO: 未实现
}

func (w *workflowImpl) TemplateList(ctx context.Context) {
	// TODO: 未实现
}

func (w *workflowImpl) UpdateNodeTemplate(ctx context.Context) {
	// TODO: 未实现
}

func (w *workflowImpl) OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error {
	msgType := &common.WsMsgType{}
	err := json.Unmarshal(b, msgType)
	if err != nil {
		return err
	}

	switch workflow.ActionType(msgType.Action) {
	case workflow.FetchGrpah: // 首次获取组态图
		return w.fetchGraph(ctx, s, msgType.MsgUUID)
	case workflow.FetchTemplate: // 首次获取模板
		return w.fetchNodeTemplate(ctx, s, msgType.MsgUUID)
	case workflow.CreateNode:
		return w.createNode(ctx, s, b)
	case workflow.UpdateNode: // 批量更新节点
		return w.upateNode(ctx, s, b)
	case workflow.BatchDelNode: // 批量删除节点
		return w.batchDelNode(ctx, s, b)
	case workflow.BatchCreateEdge: // 批量创建边
		return w.batchCreateEdge(ctx, s, b)
	case workflow.BatchDelEdge: // 批量删除边
		return w.batchDelEdge(ctx, s, b)
	default:
		return common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, code.UnknownWSActionErr)
	}
}

// 获取工作流 dag 图
func (w *workflowImpl) fetchGraph(ctx context.Context, s *melody.Session, msgUUID uuid.UUID) error {
	uuidI, _ := s.Get("uuid")
	workflowUUID := uuidI.(uuid.UUID)
	userInfo := auth.GetCurrentUser(ctx)
	resp, err := w.workflowStore.GetWorkflowGraph(ctx, userInfo.ID, workflowUUID)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.FetchGrpah), msgUUID, err)
		return err
	}

	nodeIDUUIDMap := utils.SliceToMap(resp.Nodes, func(node *repo.WorkflowNodeInfo) (int64, uuid.UUID) {
		return node.Node.ID, node.Node.UUID
	})

	nodes := utils.FilterSlice(resp.Nodes, func(node *repo.WorkflowNodeInfo) (*workflow.WSNode, bool) {
		data := &workflow.WSNode{
			UUID: node.Node.UUID,
			TemplateUUID: utils.SafeValue(func() uuid.UUID {
				return node.Template.UUID
			}, uuid.UUID{}),
			ParentUUID: nodeIDUUIDMap[node.Node.ParentID],
			UserID:     node.Node.UserID,
			Status:     node.Node.Status,
			Type:       node.Node.Type,
			Icon:       node.Node.Icon,
			Pose:       node.Node.Pose,
			Param:      node.Node.Param,
			Schema: utils.SafeValue(func() datatypes.JSON {
				return node.Template.Schema
			}, datatypes.JSON{}),
			Handles: utils.FilterSlice(node.Handles, func(h *model.WorkflowHandleTemplate) (*workflow.WSNodeHandle, bool) {
				return &workflow.WSNodeHandle{
					UUID:        h.UUID,
					HandleKey:   h.HandleKey,
					IoType:      h.IoType,
					DisplayName: h.DisplayName,
					Type:        h.Type,
					DataSource:  h.DataSource,
					DataKey:     h.DataKey,
				}, true
			}),
		}
		return data, true
	})

	edges := utils.FilterSlice(resp.Edges, func(edge *model.WorkflowEdge) (*workflow.WSWorkflowEdge, bool) {
		return &workflow.WSWorkflowEdge{
			UUID:             edge.UUID,
			SourceNodeUUID:   edge.SourceNodeUUID,
			TargetNodeUUID:   edge.TargetNodeUUID,
			SourceHandleUUID: edge.SourceHandleUUID,
			TargetHandleUUID: edge.TargetHandleUUID,
		}, true
	})

	wsResp := &workflow.WSGraph{
		Nodes: nodes,
		Edges: edges,
	}

	return common.ReplyWSOk(s, string(workflow.FetchGrpah), msgUUID, wsResp)
}

// 获取实验所有节点模板
func (w *workflowImpl) fetchNodeTemplate(ctx context.Context, s *melody.Session, msgUUID uuid.UUID) error {
	data, err := w.getWorkflow(ctx, s)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.FetchTemplate), msgUUID, err)
		return err
	}

	resp, err := w.workflowStore.GetWorkflowTemplate(ctx, data.LabID)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.FetchTemplate), msgUUID, err)
		return err
	}

	templateInfos := utils.FilterSlice(resp, func(item *repo.WorkflowTemplate) (*workflow.WSTemplateHandles, bool) {
		return &workflow.WSTemplateHandles{
			Template: &workflow.WSTemplate{
				UUID:          item.Template.UUID,
				Name:          item.Template.Name,
				DisplayName:   item.Template.DisplayName,
				Header:        item.Template.Header,
				Footer:        item.Template.Footer,
				ParamType:     item.Template.ParamType,
				Schema:        item.Template.Schema,
				ExecuteScript: item.Template.ExecuteScript,
				NodeType:      item.Template.NodeType,
				Icon:          item.Template.Icon,
			},
			Handles: utils.FilterSlice(item.Handles, func(h *model.WorkflowHandleTemplate) (*workflow.WSNodeHandle, bool) {
				return &workflow.WSNodeHandle{
					UUID:        h.UUID,
					HandleKey:   h.HandleKey,
					IoType:      h.IoType,
					DisplayName: h.DisplayName,
					Type:        h.Type,
					DataSource:  h.DataSource,
					DataKey:     h.DataKey,
				}, true

			}),
		}, true
	})

	return common.ReplyWSOk(s, string(workflow.FetchTemplate), msgUUID, &workflow.WSTemplates{
		Templates: templateInfos,
	})
}

// 创建工作流节点
func (w *workflowImpl) createNode(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[*workflow.WSCreateNode]{}
	err := json.Unmarshal(b, req)
	if err != nil || req.Data == nil {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.UnLogin)
		return nil
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.CreateNode), uuid.NewV4(), nil)
		return err
	}

	reqData := req.Data
	if reqData.TemplateUUID.IsNil() {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}

	tplNode, err := w.workflowStore.GetWorkflowTemplateByUUID(ctx, reqData.TemplateUUID)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}
	// TODO: 如果有 parent uuid 获取 paternt id
	parentID := int64(0)
	if !reqData.ParentUUID.IsNil() {
		if parentNode, err := w.workflowStore.GetWorkflowNode(ctx, reqData.ParentUUID); err != nil {
			common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
			return err
		} else {
			parentID = parentNode.ID
		}
	}

	nodeData := &model.WorkflowNode{
		WorkflowID: wk.ID,
		TemplateID: tplNode.Template.ID,
		ParentID:   parentID,
		UserID:     userInfo.ID,
		Status:     "draft",
		Type:       utils.Or(reqData.Type, "Device"),
		Icon:       tplNode.Template.Icon,
		Pose:       reqData.Pose,
		Param:      reqData.Param,
	}
	err = w.workflowStore.CreateNode(ctx, nodeData)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, err)
		return err
	}

	respData := &workflow.WSNode{
		UUID:         nodeData.UUID,
		TemplateUUID: reqData.TemplateUUID,
		ParentUUID:   reqData.ParentUUID,
		UserID:       nodeData.UserID,
		Status:       nodeData.Type,
		Type:         nodeData.Type,
		Icon:         nodeData.Icon,
		Pose:         nodeData.Pose,
		Param:        nodeData.Param,
		Schema:       tplNode.Template.Schema,
		Handles: utils.FilterSlice(tplNode.Handles, func(h *model.WorkflowHandleTemplate) (*workflow.WSNodeHandle, bool) {
			return &workflow.WSNodeHandle{
				UUID:        h.UUID,
				HandleKey:   h.HandleKey,
				IoType:      h.IoType,
				DisplayName: h.DisplayName,
				Type:        h.Type,
				DataSource:  h.DataSource,
				DataKey:     h.DataKey,
			}, true
		}),
	}

	return common.ReplyWSOk(s, string(workflow.CreateNode), req.MsgUUID, respData)
}

// 更新工作流节点
func (w *workflowImpl) upateNode(ctx context.Context, s *melody.Session, b []byte) error {
	return nil
}

// 批量删除工作流节点
func (w *workflowImpl) batchDelNode(ctx context.Context, s *melody.Session, b []byte) error {
	return nil
}

// 批量创建边
func (w *workflowImpl) batchCreateEdge(ctx context.Context, s *melody.Session, b []byte) error {
	return nil
}

// 批量删除边
func (w *workflowImpl) batchDelEdge(ctx context.Context, s *melody.Session, b []byte) error {
	return nil
}

func (w *workflowImpl) getWorkflow(ctx context.Context, s *melody.Session) (*model.Workflow, error) {
	uuidI, ok := s.Get("uuid")
	if !ok {
		return nil, code.CanNotGetWorkflowUUIDErr
	}

	workflowUUID := uuidI.(uuid.UUID)
	return w.workflowStore.GetWorkflowByUUID(ctx, workflowUUID)
}

func (w *workflowImpl) OnWSConnect(ctx context.Context, s *melody.Session) error {
	uuidI, ok := s.Get("uuid")
	if !ok {
		return code.CanNotGetWorkflowUUIDErr
	}

	workflowUUID := uuidI.(uuid.UUID)
	exist, _ := w.workflowStore.IsExist(ctx, workflowUUID)
	if !exist {
		return code.WorkflowNotExistErr
	}
	return nil
}
