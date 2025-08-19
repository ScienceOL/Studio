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

	// resp, err := w.workflowStore.GetWorkflowTemplatePage(ctx, req.LabUUID, &req.PageReq)
	// if err != nil {
	// 	return nil, err
	// }
	// _ = resp

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
	case workflow.CreateGroup:
		return w.createGroup(ctx, s, b)
	case workflow.CreateNode:
		return w.createNode(ctx, s, b)
	case workflow.UpdateNode: // 批量更新节点
		return w.upateNode(ctx, s, b)
	case workflow.BatchDelGroupNode: // 批量删除节点
		return w.batchDelGroupNode(ctx, s, b)
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
			Name: node.Action.Name,
			TemplateUUID: utils.SafeValue(func() uuid.UUID {
				return node.Action.UUID
			}, uuid.UUID{}),
			ParentUUID: nodeIDUUIDMap[node.Node.ParentID],
			UserID:     node.Node.UserID,
			Status:     node.Node.Status,
			Type:       node.Node.Type,
			Icon:       node.Node.Icon,
			Pose:       node.Node.Pose,
			Footer:     utils.Or(node.Node.Footer, node.Action.Class),
			Param:      node.Node.Param,
			Schema:     node.Action.GoalSchema,
			Handles: utils.FilterSlice(node.Handles, func(h *model.ActionHandleTemplate) (*workflow.WSNodeHandle, bool) {
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

	resp, err := w.workflowStore.GetDeviceAction(ctx, map[string]any{
		"lab_id": data.LabID,
	})

	if err != nil {
		common.ReplyWSErr(s, string(workflow.FetchTemplate), msgUUID, err)
		return err
	}

	respResNodeIDs := utils.FilterUniqSlice(resp, func(item *model.DeviceAction) (int64, bool) {
		return item.ResNodeID, true
	})

	resNodes, err := w.labStore.GetResourceNodeTemplates(ctx, respResNodeIDs)

	if err != nil || len(resNodes) != len(respResNodeIDs) {
		common.ReplyWSErr(s, string(workflow.FetchTemplate), msgUUID, err)
		return err
	}

	resMap := utils.SliceToMap(resNodes, func(item *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return item.ID, item
	})

	actionIDs := utils.FilterSlice(resp, func(item *model.DeviceAction) (int64, bool) {
		return item.ID, true
	})

	respHandles, err := w.workflowStore.GetDeviceActionHandles(ctx, actionIDs)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.FetchTemplate), msgUUID, err)
		return err
	}

	respActionMap := utils.SliceToMapSlice(resp, func(item *model.DeviceAction) (int64, *model.DeviceAction, bool) {
		return item.ResNodeID, item, true
	})

	respHandleMap := utils.SliceToMapSlice(respHandles, func(item *model.ActionHandleTemplate) (int64, *model.ActionHandleTemplate, bool) {
		return item.ActionID, item, true
	})

	templates := make([]*workflow.WSNodeTpl, 0, len(resMap))
	for id, resNode := range resMap {
		templates = append(templates, &workflow.WSNodeTpl{
			Name: resNode.Name,
			UUID: resNode.UUID,
			HandleTemplates: utils.FilterSlice(respActionMap[id], func(item *model.DeviceAction) (*workflow.WSTemplateHandles, bool) {
				return &workflow.WSTemplateHandles{

					Template: &workflow.WSTemplate{
						UUID:          item.UUID,
						Name:          item.Name,
						DisplayName:   item.Name,
						Header:        item.Class,
						Footer:        &item.Name,
						Schema:        item.Schema,
						ExecuteScript: "",
						NodeType:      "DeviceTemplate",
						Icon:          item.Icon,
					},
					Handles: utils.FilterSlice(respHandleMap[item.ID], func(h *model.ActionHandleTemplate) (*workflow.WSNodeHandle, bool) {
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
			}),
		})
	}

	return common.ReplyWSOk(s, string(workflow.FetchTemplate), msgUUID, &workflow.WSTemplates{
		Templates: templates,
	})
}

func (w *workflowImpl) createGroup(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[*workflow.WSGroup]{}
	err := json.Unmarshal(b, req)
	if err != nil || req.Data == nil {
		common.ReplyWSErr(s, string(workflow.CreateGroup), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}

	reqData := req.Data
	if reqData == nil ||
		len(reqData.Children) == 0 ||
		reqData.Pose == datatypes.NewJSONType(model.Pose{}) {
		common.ReplyWSErr(s, string(workflow.CreateGroup), req.MsgUUID, code.ParamErr.WithMsg("data is empty"))
		return err
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		common.ReplyWSErr(s, string(workflow.CreateGroup), req.MsgUUID, code.UnLogin)
		return nil
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.CreateGroup), req.MsgUUID, nil)
		return err
	}

	groupData := &model.WorkflowNode{
		WorkflowID: wk.ID,
		ActionID:   0,
		ParentID:   0,
		Name:       "group",
		UserID:     userInfo.ID,
		Status:     "draft",
		Type:       "Group",
		Icon:       "",
		Pose:       reqData.Pose,
		Param:      datatypes.JSON{},
		Footer:     "",
	}

	w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		if err := w.workflowStore.CreateNode(txCtx, groupData); err != nil {
			common.ReplyWSErr(s, string(workflow.CreateGroup), req.MsgUUID, err)
			return err
		}

		if err := w.workflowStore.UpdateWorkflowNodes(txCtx, reqData.Children, &model.WorkflowNode{
			ParentID: groupData.ID}, []string{"parent_id"}); err != nil {
			common.ReplyWSErr(s, string(workflow.CreateGroup), req.MsgUUID, err)
			return err
		}

		return nil
	})

	return nil
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
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, nil)
		return err
	}

	reqData := req.Data
	if reqData.TemplateUUID.IsNil() {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr)
		return err
	}

	deviceAction, err := w.workflowStore.GetDeviceAction(ctx, map[string]any{
		"uuid": reqData.TemplateUUID,
	})

	if err != nil || len(deviceAction) != 1 {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}

	actionHandles, err := w.workflowStore.GetDeviceActionHandles(ctx, []int64{deviceAction[0].ID})
	if err != nil {
		common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}

	parentID := int64(0)
	if !reqData.ParentUUID.IsNil() {
		if parentID = w.workflowStore.UUID2ID(ctx, &model.WorkflowNode{}, reqData.ParentUUID)[reqData.ParentUUID]; parentID == 0 {
			common.ReplyWSErr(s, string(workflow.CreateNode), req.MsgUUID, code.ParamErr.WithMsg("can not get parent node info"))
			return err
		}
	}

	nodeData := &model.WorkflowNode{
		WorkflowID: wk.ID,
		ActionID:   deviceAction[0].ID,
		ParentID:   parentID,
		UserID:     userInfo.ID,
		Name:       utils.Or(reqData.Name, deviceAction[0].Name),
		Status:     "draft",
		Type:       utils.Or(reqData.Type, "ILab"),
		Icon:       utils.Or(deviceAction[0].Icon, reqData.Icon),
		Pose:       reqData.Pose,
		Param: utils.SafeValue(func() datatypes.JSON {
			if deviceAction[0].GoalDefault.String() == "" {
				return deviceAction[0].GoalDefault
			}
			return deviceAction[0].Goal
		}, deviceAction[0].Goal),
		Footer: utils.Or(reqData.Footer, deviceAction[0].Class),
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
		Name:         utils.Or(reqData.Name, deviceAction[0].Name),
		Type:         nodeData.Type,
		Icon:         nodeData.Icon,
		Pose:         nodeData.Pose,
		Param:        nodeData.Param,
		Schema:       deviceAction[0].GoalSchema,
		Footer:       nodeData.Footer,
		Handles: utils.FilterSlice(actionHandles, func(h *model.ActionHandleTemplate) (*workflow.WSNodeHandle, bool) {
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
	if reqData.ParentUUID != nil && !(reqData.ParentUUID.IsNil()) {
		d.ParentID = w.workflowStore.UUID2ID(ctx,
			&model.WorkflowNode{},
			*reqData.ParentUUID)[*reqData.ParentUUID]
		keys = append(keys, "parent_id")
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

	if reqData.Name != nil {
		d.Name = *reqData.Name
		keys = append(keys, "name")
	}

	if reqData.Footer != nil {
		d.Footer = *reqData.Footer
		keys = append(keys, "footer")
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
func (w *workflowImpl) batchDelGroupNode(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[[]uuid.UUID]{}
	if err := json.Unmarshal(b, &req); err != nil {

		common.ReplyWSErr(s, string(workflow.BatchDelGroupNode), req.MsgUUID, code.ParamErr.WithMsg(err.Error()))
		return err
	}

	if len(req.Data) == 0 {
		common.ReplyWSOk(s, string(workflow.BatchDelGroupNode), req.MsgUUID, &workflow.WSDelNodes{})
		return nil
	}

	resp, err := w.workflowStore.DeleteWorkflowGroupNodes(ctx, req.Data)
	if err != nil {
		common.ReplyWSErr(s, string(workflow.BatchDelGroupNode), req.MsgUUID, err)
		return err
	}

	return common.ReplyWSOk(s, string(workflow.BatchDelGroupNode), req.MsgUUID, &workflow.WSDelNodes{
		NodeUUIDs: resp.NodeUUIDs,
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

	// count, err = w.workflowStore.Count(ctx, &model.WorkflowHandleTemplate{}, map[string]any{"uuid": handleUUIDs})
	// if err != nil || count != int64(len(handleUUIDs)) {
	// 	common.ReplyWSErr(s, string(workflow.BatchCreateEdge), req.MsgUUID, code.ParamErr.WithMsg("handle templet uuid not exist"))
	// 	return code.ParamErr.WithMsg("handle templet not exist")
	// }

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

// GetWorkflowList 获取工作流列表
func (w *workflowImpl) GetWorkflowList(ctx context.Context, req *workflow.WorkflowListReq) (*workflow.WorkflowListResult, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	// 获取实验室ID
	var labID int64 = 0
	if !req.LabUUID.IsNil() {
		lab, err := w.labStore.GetLabByUUID(ctx, req.LabUUID)
		if err != nil {
			return nil, err
		}
		labID = lab.ID
	}

	// 从数据库获取工作流列表
	workflows, total, err := w.workflowStore.GetWorkflowList(ctx, userInfo.ID, labID, &req.PageReq)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	respList := utils.FilterSlice(workflows, func(wf *model.Workflow) (*workflow.WorkflowListResp, bool) {
		return &workflow.WorkflowListResp{
			UUID:        wf.UUID,
			Name:        wf.Name,
			Description: wf.Description,
			UserID:      wf.UserID,
		}, true
	})

	hasMore := int64(req.Page*req.PageSize) < total
	return &workflow.WorkflowListResult{
		HasMore: hasMore,
		Data:    respList,
	}, nil
}

// GetWorkflowDetail 获取工作流详情
func (w *workflowImpl) GetWorkflowDetail(ctx context.Context, workflowUUID uuid.UUID) (*workflow.WorkflowDetailResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	// 从数据库获取工作流详情
	wf, err := w.workflowStore.GetWorkflowByUUID(ctx, workflowUUID)
	if err != nil {
		return nil, err
	}

	// 检查权限（只能查看自己的工作流）
	if wf.UserID != userInfo.ID {
		return nil, code.PermissionDenied
	}

	return &workflow.WorkflowDetailResp{
		UUID:        wf.UUID,
		Name:        wf.Name,
		Description: wf.Description,
		UserID:      wf.UserID,
	}, nil
}
