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

func (w *workflowImpl) Create(ctx context.Context, data *workflow.WorkflowReq) (*workflow.WorkflowResp, error) {
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

func (w *workflowImpl) NodeTemplateList(ctx context.Context, req *workflow.TplPageReq) (*common.PageResp[[]*workflow.TemplateNodeResp], error) {
	if req.LabUUID.IsNil() {
		return nil, code.ParamErr.WithMsg("lab uuid is empty")
	}

	resp, err := w.workflowStore.GetWorkflowTemplatePage(ctx, req.LabUUID, &req.PageReq)
	if err != nil {
		return nil, err
	}
	_ = resp

	// tplNodes := utils.FilterSlice(resp, func(item *repo.WorkflowTemplate) (*workflow.TemplateNodeResp, bool) {
	// 	return &workflow.TemplateNodeResp{
	// 		UUID: item.Template.UUID,
	// 		Type: item.Template.NodeType,
	// 		Icon: item.Template.Icon,
	// 		Name: item.Template.Name,
	// 		TemplateHandles: utils.FilterSlice(item.Handles, func(h *model.WorkflowHandleTemplate) (*workflow.TemplateHandle, bool) {
	// 			return &workflow.TemplateHandle{
	// 				HandleKey: h.HandleKey,
	// 				IoType:    h.IoType,
	// 			}, true
	// 		}),
	// 	}, true
	// })
	// _ = tplNodes

	// return &common.PageResp[*workflow.TemplateNodeResp]{
	// 	Total   :tplNode
	// Page    :
	// PageSize:
	// }, nil
	return nil, nil
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
	case workflow.FetchGraph: // 首次获取组态图
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
		common.ReplyWSErr(s, string(workflow.FetchGraph), msgUUID, err)
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

	return common.ReplyWSOk(s, string(workflow.FetchGraph), msgUUID, wsResp)
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
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr)
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
		Icon:       utils.Or(tplNode.Template.Icon, reqData.Icon),
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
	req := &common.WSData[*workflow.WSUpdateNode]{}
	if err := json.Unmarshal(b, req); err != nil {
		common.ReplyWSErr(s, string(workflow.UpdateNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}

	reqData := req.Data

	if reqData == nil ||
		reqData.UUID.IsNil() {
		common.ReplyWSErr(s, string(workflow.UpdateNode), req.MsgUUID, code.ParamErr)
		return code.ParamErr.WithMsg("empty param")
	}

	d := &model.WorkflowNode{}
	keys := make([]string, 0, 1)
	if reqData.ParentUUID != nil || !(reqData.ParentUUID.IsNil()) {
		d.ParentID = w.workflowStore.UUID2ID(ctx,
			&model.WorkflowNode{},
			*reqData.ParentUUID)[*reqData.ParentUUID]
	}

	if reqData.Status != nil {
		d.Status = *reqData.Status
		keys = append(keys, "status")
	}

	if reqData.Type != nil {
		d.Type = *reqData.Type
		keys = append(keys, "type")
	}

	if reqData.Icon != nil {
		d.Icon = *reqData.Icon
		keys = append(keys, "icon")
	}

	if reqData.Pose != nil {
		d.Pose = *reqData.Pose
		keys = append(keys, "pose")
	}

	if reqData.Param != nil {
		d.Param = *reqData.Param
		keys = append(keys, "param")
	}

	if len(keys) == 0 {
		common.ReplyWSOk(s, string(workflow.UpdateNode), reqData.UUID)
		return nil
	}

	if err := w.workflowStore.UpdateWorkflowNode(ctx, reqData.UUID, d, keys); err != nil {
		common.ReplyWSErr(s, string(workflow.UpdateNode), reqData.UUID, err)
		return err
	}

	return common.ReplyWSOk(s, string(workflow.UpdateNode), reqData.UUID, reqData)
}

// 批量删除工作流节点
func (w *workflowImpl) batchDelNode(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[[]uuid.UUID]{}
	if err := json.Unmarshal(b, &req); err != nil {

		common.ReplyWSErr(s, string(workflow.BatchDelNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}

	if len(req.Data) == 0 {
		common.ReplyWSOk(s, string(workflow.BatchDelNode), req.MsgUUID, &workflow.WSDelNodes{})
		return nil
	}

	resp, err := w.workflowStore.DeleteWorkflowNodes(ctx, req.Data)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.BatchDelNode), req.MsgUUID, err)
		return err
	}

	return common.ReplyWSOk(s, string(workflow.BatchDelNode), req.MsgUUID, &workflow.WSDelNodes{
		NodeUUIDs: resp.NodeUUIDs,
		EdgeUUIDs: resp.EdgeUUIDs,
	})
}

// 批量创建边
func (w *workflowImpl) batchCreateEdge(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[[]*workflow.WSWorkflowEdge]{}
	if err := json.Unmarshal(b, req); err != nil {
		common.ReplyWSErr(s, string(workflow.BatchCreateEdge), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return code.ParamErr.WithMsg(err.Error())
	}

	if len(req.Data) == 0 {
		common.ReplyWSErr(s, string(workflow.BatchCreateEdge), req.MsgUUID, code.ParamErr.WithMsg("edge is empty"))
		return code.ParamErr.WithMsg("edge is empty")
	}

	nodeUUIDs := make([]uuid.UUID, 0, 2*len(req.Data))
	handleUUIDs := make([]uuid.UUID, 0, 2*len(req.Data))
	for _, edge := range req.Data {
		if edge.SourceHandleUUID.IsNil() ||
			edge.TargetHandleUUID.IsNil() ||
			edge.SourceNodeUUID.IsNil() ||
			edge.TargetNodeUUID.IsNil() {

			common.ReplyWSErr(s, string(workflow.BatchCreateEdge), req.MsgUUID, code.ParamErr.WithMsg("uuid is empty"))
			return code.ParamErr.WithMsg("uuid is empty")
		}

		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, edge.SourceNodeUUID, edge.TargetNodeUUID)
		handleUUIDs = utils.AppendUniqSlice(handleUUIDs, edge.SourceHandleUUID, edge.TargetHandleUUID)
	}

	count, err := w.workflowStore.Count(ctx, &model.WorkflowNode{}, map[string]any{"uuid": nodeUUIDs})
	if err != nil || count != int64(len(nodeUUIDs)) {
		common.ReplyWSErr(s, string(workflow.BatchCreateEdge), req.MsgUUID, code.ParamErr.WithMsg("node uuid not exist"))
		return code.ParamErr.WithMsg("node uuid not exist")
	}

	count, err = w.workflowStore.Count(ctx, &model.WorkflowHandleTemplate{}, map[string]any{"uuid": handleUUIDs})
	if err != nil || count != int64(len(handleUUIDs)) {
		common.ReplyWSErr(s, string(workflow.BatchCreateEdge), req.MsgUUID, code.ParamErr.WithMsg("handle templet uuid not exist"))
		return code.ParamErr.WithMsg("handle templet not exist")
	}

	edgeDatas := utils.FilterSlice(req.Data, func(edge *workflow.WSWorkflowEdge) (*model.WorkflowEdge, bool) {
		return &model.WorkflowEdge{
			SourceNodeUUID:   edge.SourceNodeUUID,
			TargetNodeUUID:   edge.TargetNodeUUID,
			SourceHandleUUID: edge.SourceHandleUUID,
			TargetHandleUUID: edge.TargetHandleUUID,
		}, true

	})

	if err := w.workflowStore.UpsertWorkflowEdge(ctx, edgeDatas); err != nil {
		common.ReplyWSErr(s, string(workflow.BatchCreateEdge), req.MsgUUID, code.ParamErr.WithMsg("node uuid not exist"))
		return code.UpsertWorkflowEdgeErr.WithErr(err)
	}

	respDatas := utils.FilterSlice(edgeDatas, func(data *model.WorkflowEdge) (*workflow.WSWorkflowEdge, bool) {
		return &workflow.WSWorkflowEdge{
			UUID:             data.UUID,
			SourceNodeUUID:   data.SourceNodeUUID,
			TargetNodeUUID:   data.TargetNodeUUID,
			SourceHandleUUID: data.SourceHandleUUID,
			TargetHandleUUID: data.TargetHandleUUID,
		}, true
	})

	return common.ReplyWSOk(s, string(workflow.BatchCreateEdge), req.MsgUUID, respDatas)
}

// 批量删除边
func (w *workflowImpl) batchDelEdge(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[[]uuid.UUID]{}
	if err := json.Unmarshal(b, req); err != nil {
		common.ReplyWSErr(s, string(workflow.BatchDelEdge), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return code.ParamErr.WithMsg(err.Error())
	}

	if resp, err := w.workflowStore.DeleteWorkflowEdges(ctx, req.Data); err != nil {
		common.ReplyWSErr(s, string(workflow.BatchDelEdge), req.MsgUUID, code.ParamErr.WithErr(err))
		return code.UpsertWorkflowEdgeErr.WithErr(err)
	} else {
		return common.ReplyWSOk(s, string(workflow.BatchDelEdge), req.MsgUUID, resp)
	}
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
	count, err := w.workflowStore.Count(ctx, &model.Workflow{}, map[string]any{
		"uuid": workflowUUID,
	})
	if err != nil || count == 0 {
		return code.WorkflowNotExistErr
	}
	return nil
}
