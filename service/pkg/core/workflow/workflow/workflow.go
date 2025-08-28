package workflow

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"strconv"
	"time"

	"github.com/olahol/melody"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/internal/configs/schedule"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/core/workflow"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
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
	rClient       *r.Client
	wsClient      *melody.Melody
}

func New(ctx context.Context, wsClient *melody.Melody) workflow.Service {
	w := &workflowImpl{
		workflowStore: wfl.New(),
		labStore:      el.New(),
		wsClient:      wsClient,
		materialStore: mStore.NewMaterialImpl(),
		rClient:       redis.GetClient(),
	}
	events.NewEvents().Registry(ctx, notify.WorkflowRun, w.HandleNotify)
	return w
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

func (w *workflowImpl) NodeTemplateDetail(ctx context.Context, templateUUID uuid.UUID) (*workflow.NodeTemplateDetailResp, error) {
	if templateUUID.IsNil() {
		return nil, code.ParamErr.WithMsg("template uuid is empty")
	}

	// 获取节点模板详情
	template, err := w.workflowStore.GetNodeTemplateByUUID(ctx, templateUUID)
	if err != nil {
		return nil, err
	}

	// 获取实验室信息 - 根据template的lab_id获取实验室
	lab, err := w.labStore.GetLabByID(ctx, template.LabID)
	if err != nil {
		return nil, err
	}

	// 获取模板的handle列表
	handles, err := w.workflowStore.GetWorkflowHandleTemplates(ctx, []int64{template.ID})
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	handleList := utils.FilterSlice(handles, func(h *model.WorkflowHandleTemplate) (*workflow.NodeHandle, bool) {
		return &workflow.NodeHandle{
			UUID:        h.UUID,
			HandleKey:   h.HandleKey,
			IoType:      h.IoType,
			DisplayName: h.DisplayName,
			Type:        h.Type,
			DataSource:  h.DataSource,
			DataKey:     h.DataKey,
		}, true
	})

	return &workflow.NodeTemplateDetailResp{
		UUID:        template.UUID,
		Name:        template.Name,
		Class:       template.Class,
		Type:        template.Type,
		Icon:        template.Icon,
		Schema:      template.Schema,
		Goal:        template.Goal,
		GoalDefault: template.GoalDefault,
		Feedback:    template.Feedback,
		Result:      template.Result,
		LabName:     lab.Name,
		CreatedAt:   template.CreatedAt.Format("2006-01-02 15:04:05"),
		Handles:     handleList,
		Header:      template.Header,
		Footer:      template.Footer,
	}, nil
}

func (w *workflowImpl) TemplateDetail(ctx context.Context) {
	// TODO: 未实现
}

func (w *workflowImpl) TemplateList(ctx context.Context, req *workflow.TplPageReq) (*common.PageResp[[]*workflow.TemplateListResp], error) {
	if req.LabUUID.IsNil() {
		return nil, code.ParamErr.WithMsg("lab uuid is empty")
	}

	// 获取实验室信息
	lab, err := w.labStore.GetLabByUUID(ctx, req.LabUUID)
	if err != nil {
		return nil, err
	}

	resNodeIDs := []int64{}
	if len(req.Tags) > 0 {
		resNodes := make([]*model.ResourceNodeTemplate, 0, len(req.Tags))
		if err := w.labStore.FindDatas(ctx, &resNodes, map[string]any{
			"name": req.Tags,
		}, "id"); err != nil {
			return nil, err
		}
		resNodeIDs = utils.FilterSlice(resNodes, func(t *model.ResourceNodeTemplate) (int64, bool) {
			return t.ID, true
		})
	}

	// 获取模板列表
	templates, err := w.workflowStore.GetTemplateList(ctx, &common.PageReqT[*repo.QueryTemplage]{
		PageReq: req.PageReq,
		Data: &repo.QueryTemplage{
			Name:            req.Name,
			ResourceNodeIDs: resNodeIDs,
			LabID:           lab.ID,
		},
	})
	if err != nil {
		return nil, err
	}

	// 获取所有模板的handle数量
	templateIDs := utils.FilterSlice(templates.Data, func(t *model.WorkflowNodeTemplate) (int64, bool) {
		return t.ID, true
	})

	handles, err := w.workflowStore.GetWorkflowHandleTemplates(ctx, templateIDs)
	if err != nil {
		return nil, err
	}

	handleMap := utils.SliceToMapSlice(handles, func(h *model.WorkflowHandleTemplate) (int64, *model.WorkflowHandleTemplate, bool) {
		return h.WorkflowNodeID, h, true
	})

	// 转换为响应格式
	respList := utils.FilterSlice(templates.Data, func(t *model.WorkflowNodeTemplate) (*workflow.TemplateListResp, bool) {
		hs := utils.FilterSlice(handleMap[t.ID], func(h *model.WorkflowHandleTemplate) (*workflow.TemplateHandleResp, bool) {
			return &workflow.TemplateHandleResp{
				HandleKey:   h.HandleKey,
				IoType:      h.IoType,
				DisplayName: h.DisplayName,
				Type:        h.Type,
			}, true
		})

		return &workflow.TemplateListResp{
			UUID:      t.UUID,
			Name:      t.Name,   // 模板名称（从device_action name字段取）
			LabName:   lab.Name, // 实验室名字
			Header:    t.Header, // 头部信息
			Footer:    t.Footer, // 底部信息
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
			Handles:   hs,
		}, true
	})

	return &common.PageResp[[]*workflow.TemplateListResp]{
		Data:     respList,
		Total:    templates.Total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (w *workflowImpl) TemplateTags(ctx context.Context, req *workflow.TemplateTagsReq) ([]string, error) {
	labID := w.labStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
	if labID == 0 {
		return nil, code.CanNotGetLabIDErr
	}

	resourceNames := w.labStore.GetAllResourceName(ctx, labID)

	return resourceNames, nil
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

	var data any

	switch workflow.ActionType(msgType.Action) {
	case workflow.FetchGraph: // 首次获取组态图
		data, err = w.fetchGraph(ctx, s)
	case workflow.FetchTemplate: // 首次获取模板
		data, err = w.fetchNodeTemplate(ctx, s)
	case workflow.FetchDevice:
		data, err = w.fetchDevices(ctx, s, b)
	case workflow.CreateGroup:
		data, err = w.createGroup(ctx, s, b)
	case workflow.CreateNode:
		data, err = w.createNode(ctx, s, b)
	case workflow.UpdateNode: // 批量更新节点
		data, err = w.upateNode(ctx, s, b)
	case workflow.BatchDelNode: // 批量删除节点
		data, err = w.batchDelNodes(ctx, s, b)
	case workflow.BatchCreateEdge: // 批量创建边
		data, err = w.batchCreateEdge(ctx, s, b)
	case workflow.BatchDelEdge: // 批量删除边
		data, err = w.batchDelEdge(ctx, s, b)
	case workflow.SaveWorkflow:
		data, err = w.batchSave(ctx, s, b)
	case workflow.RunWorkflow:
		data, err = w.runWorkflow(ctx, s, b)
	case workflow.StopWorkflow:
		data, err = w.stopWorkflow(ctx, s, b)
	case workflow.FetchWorkflowStatus:
		data, err = w.fetchWorkflowTask(ctx, s)
	case workflow.Dumplicate:
		data, err = w.duplicateWorkflow(ctx, s, b)

	default:
		return common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, code.UnknownWSActionErr)
	}

	if err != nil {
		common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, err)
		return err
	}

	if data != nil {
		return common.ReplyWSOk(s, msgType.Action, msgType.MsgUUID, data)
	} else {
		return common.ReplyWSOk(s, msgType.Action, msgType.MsgUUID)
	}

}

// 获取工作流 dag 图
func (w *workflowImpl) fetchGraph(ctx context.Context, s *melody.Session) (any, error) {
	uuidI, _ := s.Get("uuid")
	workflowUUID := uuidI.(uuid.UUID)
	userInfo := auth.GetCurrentUser(ctx)
	resp, err := w.workflowStore.GetWorkflowGraph(ctx, userInfo.ID, workflowUUID)
	if err != nil {
		return nil, err
	}

	nodeIDUUIDMap := utils.Slice2Map(resp.Nodes, func(node *repo.WorkflowNodeInfo) (int64, uuid.UUID) {
		return node.Node.ID, node.Node.UUID
	})

	nodes := utils.FilterSlice(resp.Nodes, func(node *repo.WorkflowNodeInfo) (*workflow.WSNode, bool) {
		data := &workflow.WSNode{
			UUID:        node.Node.UUID,
			ParentUUID:  nodeIDUUIDMap[node.Node.ParentID],
			UserID:      node.Node.UserID,
			Status:      node.Node.Status,
			Type:        node.Node.Type,
			Icon:        node.Node.Icon,
			Pose:        node.Node.Pose,
			Footer:      utils.Or(node.Node.Footer, ""),
			Param:       node.Node.Param,
			DeviceName:  node.Node.DeviceName,
			LabNodeType: node.Node.LabNodeType,
			Disabled:    node.Node.Disabled,
			Minimized:   node.Node.Minimized,
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

		if node.Action != nil {
			data.Name = utils.Or(node.Node.Name, node.Action.Name)
			data.Schema = node.Action.Schema
			data.TemplateUUID = node.Action.UUID
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

	return wsResp, nil
}

// 获取实验所有节点模板
func (w *workflowImpl) fetchNodeTemplate(ctx context.Context, s *melody.Session) (any, error) {
	data, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	resp, err := w.workflowStore.GetWorkflowNodeTemplate(ctx, map[string]any{
		"lab_id": data.LabID,
	})

	if err != nil {
		return nil, err
	}

	respResNodeIDs := utils.FilterUniqSlice(resp, func(item *model.WorkflowNodeTemplate) (int64, bool) {
		return item.ResourceNodeID, true
	})

	resNodes, err := w.labStore.GetResourceNodeTemplates(ctx, respResNodeIDs)

	if err != nil || len(resNodes) != len(respResNodeIDs) {
		return nil, code.TemplateNodeNotFoundErr
	}

	resMap := utils.Slice2Map(resNodes, func(item *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return item.ID, item
	})

	actionIDs := utils.FilterSlice(resp, func(item *model.WorkflowNodeTemplate) (int64, bool) {
		return item.ID, true
	})

	respHandles, err := w.workflowStore.GetWorkflowHandleTemplates(ctx, actionIDs)
	if err != nil {
		return nil, err
	}

	respActionMap := utils.SliceToMapSlice(resp, func(item *model.WorkflowNodeTemplate) (int64, *model.WorkflowNodeTemplate, bool) {
		return item.ResourceNodeID, item, true
	})

	respHandleMap := utils.SliceToMapSlice(respHandles, func(item *model.WorkflowHandleTemplate) (int64, *model.WorkflowHandleTemplate, bool) {
		return item.WorkflowNodeID, item, true
	})

	templates := make([]*workflow.WSNodeTpl, 0, len(resMap))
	for id, resNode := range resMap {
		templates = append(templates, &workflow.WSNodeTpl{
			Name: resNode.Name,
			UUID: resNode.UUID,
			HandleTemplates: utils.FilterSlice(respActionMap[id], func(item *model.WorkflowNodeTemplate) (*workflow.WSTemplateHandles, bool) {
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
					Handles: utils.FilterSlice(respHandleMap[item.ID], func(h *model.WorkflowHandleTemplate) (*workflow.WSNodeHandle, bool) {
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

	return &workflow.WSTemplates{
		Templates: templates,
	}, nil
}

func (w *workflowImpl) fetchDevices(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	// 根据 工作流模板节点 uuid 获取到所有的设备名称, 返回一个 list
	req := &common.WSData[uuid.UUID]{}
	err := json.Unmarshal(b, req)
	if err != nil || req.Data.IsNil() {
		return nil, code.ParamErr.WithMsg("workflow template uuid is empty")
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	workflowNodes := make([]*model.WorkflowNodeTemplate, 0, 1)
	if err := w.materialStore.FindDatas(ctx, &workflowNodes, map[string]any{
		"uuid": req.Data,
	}, "resource_node_id"); err != nil {
		return nil, err
	}

	if len(workflowNodes) != 1 {
		return nil, code.ParamErr.WithMsgf("can not found resource node uuid: %s", req.Data)
	}

	materialNodes := make([]*model.MaterialNode, 0, 1)
	if err := w.materialStore.FindDatas(ctx, &materialNodes, map[string]any{
		"lab_id":           wk.LabID,
		"resource_node_id": workflowNodes[0].ResourceNodeID,
	}, "name"); err != nil {
		return nil, err
	}

	deviceNames := utils.FilterSlice(materialNodes, func(node *model.MaterialNode) (string, bool) {
		return node.Name, true
	})

	return deviceNames, nil
}

func (w *workflowImpl) createGroup(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	req := &common.WSData[*workflow.WSGroup]{}
	err := json.Unmarshal(b, req)
	if err != nil || req.Data == nil {
		return nil, code.ParamErr.WithErr(err)
	}

	reqData := req.Data
	if reqData == nil ||
		len(reqData.Children) == 0 ||
		reqData.Pose == datatypes.NewJSONType(model.Pose{}) {
		return nil, code.ParamErr.WithMsg("data is empty")
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	groupData := &model.WorkflowNode{
		WorkflowID:     wk.ID,
		WorkflowNodeID: 0,
		ParentID:       0,
		Name:           "group",
		UserID:         userInfo.ID,
		Status:         "draft",
		Type:           "Group",
		Icon:           "",
		Pose:           reqData.Pose,
		Param:          datatypes.JSON{},
		Footer:         "",
	}
	groupData.UUID = reqData.UUID

	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		if err := w.workflowStore.CreateNode(txCtx, groupData); err != nil {
			return err
		}

		if err := w.workflowStore.UpdateWorkflowNodes(txCtx, reqData.Children, &model.WorkflowNode{
			ParentID: groupData.ID}, []string{"parent_id"}); err != nil {
			return err
		}

		return nil
	})

	return nil, err
}

// 创建工作流节点
func (w *workflowImpl) createNode(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	req := &common.WSData[*workflow.WSCreateNode]{}
	err := json.Unmarshal(b, req)
	if err != nil || req.Data == nil {
		return nil, code.ParamErr.WithMsg(err.Error())
	}
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	reqData := req.Data
	if reqData.TemplateUUID.IsNil() {
		return nil, code.ParamErr
	}

	deviceAction, err := w.workflowStore.GetWorkflowNodeTemplate(ctx, map[string]any{
		"uuid": reqData.TemplateUUID,
	})

	if err != nil || len(deviceAction) != 1 {
		return nil, code.WorkflowTemplateNotFoundErr
	}

	actionHandles, err := w.workflowStore.GetWorkflowHandleTemplates(ctx, []int64{deviceAction[0].ID})
	if err != nil {
		return nil, err
	}

	parentID := int64(0)
	if !reqData.ParentUUID.IsNil() {
		if parentID = w.workflowStore.UUID2ID(ctx, &model.WorkflowNode{}, reqData.ParentUUID)[reqData.ParentUUID]; parentID == 0 {
			return nil, code.WorkflowNodeNotFoundErr.WithMsgf("parent uuid: %s", reqData.ParentUUID)
		}
	}

	nodeData := &model.WorkflowNode{
		WorkflowID:     wk.ID,
		WorkflowNodeID: deviceAction[0].ID,
		ParentID:       parentID,
		UserID:         userInfo.ID,
		ActionName:     deviceAction[0].Name,
		ActionType:     deviceAction[0].Type,
		Name:           utils.Or(reqData.Name, deviceAction[0].Name),
		Status:         "draft",
		Type:           utils.Or(reqData.Type, model.WorkflowNodeILab),
		Icon:           utils.Or(deviceAction[0].Icon, reqData.Icon),
		Pose:           reqData.Pose,
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
		return nil, err
	}

	respData := &workflow.WSNode{
		UUID:         nodeData.UUID,
		TemplateUUID: reqData.TemplateUUID,
		ParentUUID:   reqData.ParentUUID,
		UserID:       nodeData.UserID,
		Status:       nodeData.Status,
		Name:         utils.Or(reqData.Name, deviceAction[0].Name),
		Type:         nodeData.Type,
		Icon:         nodeData.Icon,
		Pose:         nodeData.Pose,
		Param:        nodeData.Param,
		Schema:       deviceAction[0].Schema,
		Footer:       nodeData.Footer,
		Handles: utils.FilterSlice(actionHandles, func(h *model.WorkflowHandleTemplate) (*workflow.WSNodeHandle, bool) {
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

	return respData, nil
}

// 更新工作流节点
func (w *workflowImpl) upateNode(ctx context.Context, _ *melody.Session, b []byte) (any, error) {
	req := &common.WSData[*workflow.WSUpdateNode]{}
	if err := json.Unmarshal(b, req); err != nil {
		return nil, code.ParamErr.WithMsg(err.Error())
	}

	reqData := req.Data
	if reqData == nil ||
		reqData.UUID.IsNil() {
		return nil, code.ParamErr.WithMsg("empty uuid")
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

	if reqData.Minimized != nil {
		d.Minimized = *reqData.Minimized
		keys = append(keys, "minimized")
	}

	if reqData.Disabled != nil {
		d.Disabled = *reqData.Disabled
		keys = append(keys, "disabled")
	}

	if reqData.DeviceName != nil {
		d.DeviceName = reqData.DeviceName
		keys = append(keys, "device_name")
	}

	if len(keys) == 0 {
		return nil, nil
	}

	if err := w.workflowStore.UpdateWorkflowNode(ctx, reqData.UUID, d, keys); err != nil {
		return nil, err
	}

	return reqData, nil
}

// 批量删除工作流节点
func (w *workflowImpl) batchDelNodes(ctx context.Context, _ *melody.Session, b []byte) (any, error) {
	req := &common.WSData[[]uuid.UUID]{}
	if err := json.Unmarshal(b, &req); err != nil {

		return nil, code.ParamErr.WithMsg(err.Error())
	}

	if len(req.Data) == 0 {
		return &workflow.WSDelNodes{}, nil
	}

	resp, err := w.workflowStore.DeleteWorkflowNodes(ctx, req.Data)
	if err != nil {
		return nil, err
	}

	return &workflow.WSDelNodes{
		NodeUUIDs: resp.NodeUUIDs,
	}, nil
}

// 批量创建边
func (w *workflowImpl) batchCreateEdge(ctx context.Context, _ *melody.Session, b []byte) (any, error) {
	req := &common.WSData[[]*workflow.WSWorkflowEdge]{}
	if err := json.Unmarshal(b, req); err != nil {
		return nil, code.ParamErr.WithMsg(err.Error())
	}

	if len(req.Data) == 0 {
		return nil, code.ParamErr.WithMsg("edge is empty")
	}

	nodeUUIDs := make([]uuid.UUID, 0, 2*len(req.Data))
	handleUUIDs := make([]uuid.UUID, 0, 2*len(req.Data))
	for _, edge := range req.Data {
		if edge.SourceHandleUUID.IsNil() ||
			edge.TargetHandleUUID.IsNil() ||
			edge.SourceNodeUUID.IsNil() ||
			edge.TargetNodeUUID.IsNil() {

			return nil, code.ParamErr.WithMsg("uuid is empty")
		}

		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, edge.SourceNodeUUID, edge.TargetNodeUUID)
		handleUUIDs = utils.AppendUniqSlice(handleUUIDs, edge.SourceHandleUUID, edge.TargetHandleUUID)
	}

	count, err := w.workflowStore.Count(ctx, &model.WorkflowNode{}, map[string]any{"uuid": nodeUUIDs})
	if err != nil || count != int64(len(nodeUUIDs)) {

		return nil, code.ParamErr.WithMsg("node uuid not exist")
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
		return nil, code.UpsertWorkflowEdgeErr.WithErr(err)
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

	return respDatas, nil
}

// 批量删除边
func (w *workflowImpl) batchDelEdge(ctx context.Context, _ *melody.Session, b []byte) (any, error) {
	req := &common.WSData[[]uuid.UUID]{}
	if err := json.Unmarshal(b, req); err != nil {
		return nil, code.ParamErr.WithMsg(err.Error())
	}

	if resp, err := w.workflowStore.DeleteWorkflowEdges(ctx, req.Data); err != nil {
		return nil, code.UpsertWorkflowEdgeErr.WithErr(err)
	} else {
		return resp, nil
	}
}

// 批量保存工作流节点
func (w *workflowImpl) batchSave(ctx context.Context, _ *melody.Session, b []byte) (any, error) {
	req := &common.WSData[*workflow.WSGraph]{}
	if err := json.Unmarshal(b, req); err != nil {
		return nil, code.ParamErr.WithMsg(err.Error())
	}

	err := w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		// 保存节点
		if err := w.batchSaveNodes(txCtx, req.Data.Nodes); err != nil {
			return code.SaveWorkflowNodeErr.WithMsg(err.Error())
		}

		// 保存边
		if err := w.batchSaveEdge(txCtx, req.Data.Edges); err != nil {
			return code.SaveWorkflowEdgeErr.WithMsg(err.Error())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (w *workflowImpl) runWorkflow(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	req := &common.WSData[any]{}
	if err := json.Unmarshal(b, req); err != nil {
		return nil, code.ParamErr.WithMsg(err.Error())
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin.WithMsg("can not get user info")
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	labMap := w.workflowStore.ID2UUID(ctx, &model.Laboratory{}, wk.LabID)

	labUUID, ok := labMap[wk.LabID]
	if !ok {
		return nil, code.ParamErr.WithMsg("can not get lab uuid")
	}

	var taskUUID uuid.UUID

	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		task := &model.WorkflowTask{
			LabID:      wk.LabID,
			WorkflowID: wk.ID,
			UserID:     userInfo.ID,
		}
		if err := w.workflowStore.CreateWorkflowTask(txCtx, task); err != nil {
			return err
		}
		taskUUID = task.UUID
		conf := webapp.Config().Job
		data := engine.WorkflowInfo{
			Action:       engine.StartJob,
			TaskUUID:     task.UUID,
			WorkflowUUID: wk.UUID,
			LabUUID:      labUUID,
			UserID:       wk.UserID,
		}

		dataB, _ := json.Marshal(data)

		ret := w.rClient.LPush(ctx, conf.JobQueueName, dataB)
		if ret.Err() != nil {
			return code.ParamErr.WithMsgf("push workflow redis msg err: %+v", ret.Err())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return taskUUID, nil
}

func (w *workflowImpl) stopWorkflow(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	req := &common.WSData[uuid.UUID]{}
	if err := json.Unmarshal(b, req); err != nil || req.Data.IsNil() {
		return nil, code.ParamErr.WithMsg(err.Error())
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin.WithMsg("can not get user info")
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	labMap := w.workflowStore.ID2UUID(ctx, &model.Laboratory{}, wk.LabID)

	labUUID, ok := labMap[wk.LabID]
	if !ok {
		return nil, code.ParamErr.WithMsg("can not get lab info")
	}

	tasks := make([]*model.WorkflowTask, 0, 1)
	if err := w.workflowStore.FindDatas(ctx, &tasks, map[string]any{
		"uuid": req.Data,
	}, "status"); err != nil || len(tasks) != 1 {
		return nil, code.WorkflowTaskNotFoundErr.WithMsg("can not get workflow task")
	}

	task := tasks[0]
	switch task.Status {
	case model.WorkflowTaskStatusPending, model.WorkflowTaskStatusRunnig:
	case model.WorkflowTaskStatusStoped, model.WorkflowTaskStatusFiled, model.WorkflowTaskStatusSuccessed:
		return nil, code.WorkflowTaskFinished
	default:
		return nil, code.WorkflowTaskStatusErr
	}

	conf := schedule.Config().Job
	data := engine.WorkflowInfo{
		Action:       engine.StopJob,
		TaskUUID:     req.Data,
		WorkflowUUID: wk.UUID,
		LabUUID:      labUUID,
		UserID:       wk.UserID,
	}

	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		task := tasks[0]
		task.Status = model.WorkflowTaskStatusStoped
		if err := w.workflowStore.UpdateData(ctx, task, map[string]any{
			"uuid": req.Data,
		}, "status"); err != nil {
			return err
		}

		dataB, _ := json.Marshal(data)
		ret := w.rClient.LPush(ctx, conf.JobQueueName, dataB)
		if ret.Err() != nil {
			return code.ParamErr.WithMsgf("push workflow redis msg err: %+v", ret.Err())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data.TaskUUID, nil
}

func (w *workflowImpl) fetchWorkflowTask(ctx context.Context, s *melody.Session) (any, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	tasks := make([]*model.WorkflowTask, 0, 10)
	err = w.workflowStore.FindDatas(ctx, &tasks, map[string]any{
		"lab_id":      wk.LabID,
		"workflow_id": wk.ID,
		"user_id":     userInfo.ID,
		"status":      []string{"pending", "running"},
	}, "id", "uuid")
	if err != nil {
		return nil, err
	}

	return utils.FilterSlice(tasks, func(task *model.WorkflowTask) (uuid.UUID, bool) {
		return task.UUID, true
	}), nil
}

func (w *workflowImpl) duplicateWorkflow(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	req := &common.WSData[*workflow.DuplicateWorkflow]{}
	if err := json.Unmarshal(b, req); err != nil {
		return nil, code.ParamErr.WithMsg(err.Error())
	}

	if req.Data == nil || req.Data.SourceUUID.IsNil() {
		return nil, code.ParamErr.WithMsg("workflow source uuid is empty")
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	wk, err := w.workflowStore.GetWorkflowByUUID(ctx, req.Data.SourceUUID)
	if err != nil {
		return nil, err
	}

	nodes, err := w.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": wk.ID,
	})

	if err != nil {
		return nil, err
	}

	nodeUUIDs := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		return node.UUID, true
	})

	edges, err := w.workflowStore.GetWorkflowEdges(ctx, nodeUUIDs)
	if err != nil {
		return nil, err
	}

	preBuildNodes := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (*utils.Node[int64, *model.WorkflowNode], bool) {
		return &utils.Node[int64, *model.WorkflowNode]{
			Name:   node.ID,
			Parent: node.ParentID,
			Data:   node,
		}, true
	})

	buildNodes, err := utils.BuildHierarchy(preBuildNodes)
	if err != nil {
		return nil, err
	}
	newWK, err := w.getWorkflow(ctx, s)
	if err != nil {
		return nil, err
	}

	old2NewIDMap := make(map[int64]int64)
	old2NewUUIDMap := make(map[uuid.UUID]uuid.UUID)
	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range buildNodes {
			newNodes := utils.FilterSlice(nodes, func(oldNode *model.WorkflowNode) (*model.WorkflowNode, bool) {
				return &model.WorkflowNode{
					WorkflowID:     newWK.ID,
					WorkflowNodeID: oldNode.WorkflowNodeID,
					ParentID:       old2NewIDMap[oldNode.ParentID],
					Name:           oldNode.Name,
					UserID:         userInfo.ID,
					Status:         oldNode.Status,
					Type:           oldNode.Type,
					LabNodeType:    oldNode.LabNodeType,
					Icon:           oldNode.Icon,
					Pose:           oldNode.Pose,
					Param:          oldNode.Param,
					Footer:         oldNode.Footer,
					DeviceName:     oldNode.DeviceName,
					ActionName:     oldNode.ActionName,
					ActionType:     oldNode.ActionType,
					Disabled:       oldNode.Disabled,
					Minimized:      oldNode.Minimized,

					OldNode: oldNode,
				}, true
			})

			if err := w.workflowStore.UpsertNodes(txCtx, newNodes); err != nil {
				return err
			}

			utils.Range(newNodes, func(_ int, newNode *model.WorkflowNode) bool {
				old2NewIDMap[newNode.OldNode.ID] = newNode.ID
				old2NewUUIDMap[newNode.OldNode.UUID] = newNode.UUID
				return true
			})
		}

		newEdges := utils.FilterSlice(edges, func(edge *model.WorkflowEdge) (*model.WorkflowEdge, bool) {
			sourceNodeUUID, ok := old2NewUUIDMap[edge.SourceNodeUUID]
			if !ok || sourceNodeUUID.IsNil() {
				logger.Warnf(txCtx, "can not duplicate edge source node uuid: %s", edge.SourceNodeUUID)
				return nil, false
			}

			targetNodeUUID, ok := old2NewUUIDMap[edge.TargetNodeUUID]
			if !ok || targetNodeUUID.IsNil() {
				logger.Warnf(txCtx, "can not duplicate edge target node uuid: %s", edge.TargetNodeUUID)
				return nil, false
			}

			return &model.WorkflowEdge{
				SourceNodeUUID:   sourceNodeUUID,
				TargetNodeUUID:   targetNodeUUID,
				SourceHandleUUID: edge.SourceHandleUUID,
				TargetHandleUUID: edge.TargetHandleUUID,
			}, true
		})

		return w.workflowStore.DuplicateEdge(txCtx, newEdges)
	})

	if err != nil {
		return nil, err
	}

	return w.fetchGraph(ctx, s)
}

func (w *workflowImpl) batchSaveNodes(ctx context.Context, nodes []*workflow.WSNode) error {
	dbNodes, err := utils.FilterSliceWithErr(nodes, func(node *workflow.WSNode) ([]*model.WorkflowNode, bool, error) {
		if node.UUID.IsNil() {
			return nil, false, code.ParamErr.WithMsg("uuid is empty")
		}
		data := &model.WorkflowNode{
			Icon:       node.Icon,
			Pose:       node.Pose,
			Param:      node.Param,
			Footer:     node.Footer,
			DeviceName: node.DeviceName,
			Disabled:   node.Disabled,
			Minimized:  node.Minimized,
		}
		data.UUID = node.UUID
		data.UpdatedAt = time.Now()
		return []*model.WorkflowNode{data}, true, nil
	})

	if err != nil {
		return err
	}

	return w.workflowStore.UpsertNodes(ctx, dbNodes)
}

func (w *workflowImpl) batchSaveEdge(ctx context.Context, edges []*workflow.WSWorkflowEdge) error {
	nodeUUIDs := make([]uuid.UUID, 0, 2*len(edges))
	handleUUIDs := make([]uuid.UUID, 0, 2*len(edges))
	for _, edge := range edges {
		if edge.UUID.IsNil() ||
			edge.SourceNodeUUID.IsNil() ||
			edge.TargetNodeUUID.IsNil() ||
			edge.SourceHandleUUID.IsNil() ||
			edge.TargetHandleUUID.IsNil() {
			return code.ParamErr.WithMsg("edge uuid is empty")
		}

		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, edge.SourceNodeUUID)
		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, edge.TargetNodeUUID)

		handleUUIDs = utils.AppendUniqSlice(handleUUIDs, edge.SourceHandleUUID)
		handleUUIDs = utils.AppendUniqSlice(handleUUIDs, edge.TargetHandleUUID)
	}

	nodeCount, err := w.workflowStore.Count(ctx, &model.WorkflowNode{}, map[string]any{
		"uuid": nodeUUIDs,
	})

	if err != nil {
		return err
	}
	if int(nodeCount) != len(nodeUUIDs) {
		return code.ParamErr.WithMsg("node not exist")
	}

	handleCount, err := w.workflowStore.Count(ctx, &model.WorkflowHandleTemplate{}, map[string]any{
		"uuid": handleUUIDs,
	})
	if err != nil {
		return err
	}
	if int(handleCount) != len(handleUUIDs) {
		return code.ParamErr.WithMsg("node not exist")
	}

	workflowEdges := utils.FilterSlice(edges, func(edge *workflow.WSWorkflowEdge) (*model.WorkflowEdge, bool) {
		return &model.WorkflowEdge{}, true
	})

	return w.workflowStore.UpsertEdge(ctx, workflowEdges)
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
func (w *workflowImpl) GetWorkflowDetail(ctx context.Context, req *workflow.DetailReq) (*workflow.WorkflowDetailResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	// 从数据库获取工作流详情
	wf, err := w.workflowStore.GetWorkflowByUUID(ctx, req.UUID)
	if err != nil {
		return nil, err
	}

	wfNodes, err := w.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": wf.ID,
	})

	if err != nil {
		return nil, err
	}

	tplNodeIDs := utils.FilterSlice(wfNodes, func(node *model.WorkflowNode) (int64, bool) {
		return node.WorkflowNodeID, node.WorkflowNodeID > 0
	})

	handles, err := w.workflowStore.GetWorkflowHandleTemplates(ctx, tplNodeIDs)
	if err != nil {
		return nil, err
	}

	handleMap := utils.SliceToMapSlice(handles, func(h *model.WorkflowHandleTemplate) (int64, *model.WorkflowHandleTemplate, bool) {
		return h.WorkflowNodeID, h, true
	})

	return &workflow.WorkflowDetailResp{
		UUID:        wf.UUID,
		Name:        wf.Name,
		Description: wf.Description,
		UserID:      wf.UserID,
		Nodes: utils.FilterSlice(wfNodes, func(node *model.WorkflowNode) (*workflow.WSNode, bool) {
			if node.Type == model.WorkflowNodeGroup {
				return nil, false
			}
			return &workflow.WSNode{
				UUID:   node.UUID,
				Name:   node.Name,
				UserID: node.UserID,
				Status: node.Status,
				Type:   node.Type,
				Icon:   node.Icon,
				Pose:   node.Pose,
				Param:  node.Param,
				Handles: utils.FilterSlice(handleMap[node.WorkflowNodeID], func(h *model.WorkflowHandleTemplate) (*workflow.WSNodeHandle, bool) {
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
				Footer:      node.Footer,
				DeviceName:  node.DeviceName,
				Disabled:    node.Disabled,
				Minimized:   node.Minimized,
				LabNodeType: node.LabNodeType,
			}, true

		}),
	}, nil
}

func (w *workflowImpl) HandleNotify(ctx context.Context, msg string) error {
	notifyData := &notify.SendMsg{}
	if err := json.Unmarshal([]byte(msg), notifyData); err != nil {
		logger.Errorf(ctx, "HandleNotify unmarshal data err: %+v", err)
		return err
	}

	d := &common.Resp{
		Code: code.Success,
		Data: &common.WSData[any]{
			WsMsgType: common.WsMsgType{
				Action:  string(workflow.WorkflowUpdate),
				MsgUUID: notifyData.UUID,
			},
			Data: notifyData.Data,
		},
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(d)
	return w.wsClient.BroadcastFilter(data, func(s *melody.Session) bool {
		sessionValue, ok := s.Get("uuid")
		if !ok {
			return false
		}

		if sessionValue.(uuid.UUID) == notifyData.WorkflowUUID {
			return true
		}

		return false
	})
}

func (w *workflowImpl) WorkflowTaskList(ctx context.Context, req *workflow.WorkflowTaskReq) (*common.PageMoreResp[[]*workflow.WorkflowTaskResp], error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	wk, err := w.workflowStore.GetWorkflowByUUID(ctx, req.UUID)
	if err != nil {
		return nil, code.CanNotGetworkflowErr
	}

	resp, err := w.workflowStore.GetWorkflowTasks(ctx, &common.PageReqT[*repo.TaskReq]{
		PageReq: req.PageReq,
		Data: &repo.TaskReq{
			UserID:     userInfo.ID,
			LabID:      wk.LabID,
			WrokflowID: wk.ID,
		},
	})
	if err != nil {
		return nil, err
	}

	return &common.PageMoreResp[[]*workflow.WorkflowTaskResp]{
		HasMore:  resp.HasMore,
		Page:     resp.Page,
		PageSize: resp.PageSize,
		Data: utils.FilterSlice(resp.Data, func(task *model.WorkflowTask) (*workflow.WorkflowTaskResp, bool) {
			return &workflow.WorkflowTaskResp{
				UUID:       task.UUID,
				Status:     task.Status,
				CreatedAt:  task.CreatedAt,
				FinishedAt: task.FinishedTime,
			}, true
		}),
	}, nil
}

func (w *workflowImpl) TaskDownload(ctx context.Context, req *workflow.WorkflowTaskDownloadReq) (*bytes.Buffer, error) {
	taskID := w.workflowStore.UUID2ID(ctx, &model.WorkflowTask{}, req.UUID)[req.UUID]
	if taskID <= 0 {
		return nil, code.WorkflowTaskNotFoundErr
	}

	jobs := make([]*model.WorkflowNodeJob, 0, 2)
	w.workflowStore.FindDatas(ctx, &jobs, map[string]any{
		"workflow_task_id": taskID,
	})

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入CSV头部
	header := []string{"ID", "状态", "数据", "更新时间", "创建时间"}
	if err := writer.Write(header); err != nil {
		return nil, code.FormatCSVTaskErr
	}

	for _, j := range jobs {
		if err := writer.Write([]string{strconv.FormatInt(j.ID, 10),
			string(j.Status),
			string(j.Data),
			j.UpdatedAt.Format(time.DateTime),
			j.CreatedAt.Format(time.DateTime)}); err != nil {
			return nil, err
		}
	}
	writer.Flush()

	return &buf, nil

}

func (w *workflowImpl) UpdateWorkflow(ctx context.Context, req *workflow.UpdateReq) error {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}

	if req.UUID.IsNil() {
		return code.ParamErr.WithMsg("workflow uuid is empty")
	}

	wk := &model.Workflow{}
	if err := w.workflowStore.GetData(ctx, wk, map[string]any{
		"uuid": req.UUID,
	}); err != nil {
		return err
	}

	if userInfo.ID != wk.UserID {
		return code.NoPermission
	}

	keys := make([]string, 0, 3)
	if req.Name != nil {
		wk.Name = utils.Or(*req.Name, "Untitled")
		keys = append(keys, "name")
	}

	if req.Published != nil {
		wk.Published = utils.Or(*req.Published, false)
		keys = append(keys, "published")
	}

	if req.Description != nil {
		wk.Description = utils.Or(req.Description, wk.Description)
		keys = append(keys, "description")
	}

	if len(keys) == 0 {
		return nil
	}

	if err := w.workflowStore.UpdateData(ctx, wk, map[string]any{
		"uuid": req.UUID,
	}, keys...); err != nil {
		return err
	}

	return nil
}

func (w *workflowImpl) DelWorkflow(ctx context.Context, req *workflow.DelReq) error {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}
	// 检查权限
	wf := &model.Workflow{}
	if err := w.workflowStore.GetData(ctx, wf, map[string]any{
		"uuid": req.UUID,
	}); err != nil {
		return err
	}

	if wf.UserID != userInfo.ID {
		return code.NoPermission
	}

	return w.workflowStore.DelWorkflow(ctx, wf.ID)
}

func (w *workflowImpl) WorkflowTemplateList(ctx context.Context, req *workflow.WorkflowTemplateListReq) (*common.PageResp[[]*workflow.WorkflowTemplateListRes], error) {
	res, err := w.workflowStore.GetWorkflow(ctx, &common.PageReqT[*repo.QueryWorkflow]{
		PageReq: req.PageReq,
		Data: &repo.QueryWorkflow{
			Tags: req.Tags,
		},
	})

	if err != nil {
		return nil, err
	}

	return &common.PageResp[[]*workflow.WorkflowTemplateListRes]{
		Total:    res.Total,
		Page:     res.Page,
		PageSize: res.PageSize,
		Data: utils.FilterSlice(res.Data, func(item *model.Workflow) (*workflow.WorkflowTemplateListRes, bool) {
			return &workflow.WorkflowTemplateListRes{
				UUID:      item.UUID,
				Name:      item.Name,
				UserID:    item.UserID,
				CreatedAt: item.CreatedAt,
			}, true
		}),
	}, nil
}

func (w *workflowImpl) WorkflowTemplateTags(ctx context.Context) ([]string, error) {
	return w.workflowStore.GetTemplateTags(ctx, model.WorkflowTemplateTag)
}

func (w *workflowImpl) ForkWrokflow(ctx context.Context, req *workflow.ForkReq) error {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}

	// 1. 验证源工作流模板是否存在且已发布
	sourceWorkflow, err := w.workflowStore.GetWorkflowByUUID(ctx, req.SourceWorkflowUUID)
	if err != nil {
		return err
	}

	// 检查工作流是否已发布（所有实验室可见）
	if !sourceWorkflow.Published {
		return code.NoPermission.WithMsg("source workflow is not published")
	}

	// 2. 验证目标实验室是否存在
	targetLab, err := w.labStore.GetLabByUUID(ctx, req.TargetLabUUID)
	if err != nil {
		return err
	}

	// 3. 检查目标实验室是否已有同名工作流
	existingWorkflow := &model.Workflow{}
	err = w.workflowStore.GetData(ctx, existingWorkflow, map[string]any{
		"lab_id":  targetLab.ID,
		"name":    sourceWorkflow.Name,
		"user_id": userInfo.ID,
	})
	if err == nil {
		// 如果存在同名工作流，返回错误
		return code.CreateDataErr.WithMsg("workflow with same name already exists in target lab")
	}

	// 4. 获取源工作流的所有节点
	nodes, err := w.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": sourceWorkflow.ID,
	})
	if err != nil {
		return err
	}

	// 5. 验证目标实验室是否有所有必需的模板（非阻断）
	if templateValidationErr := w.validateTemplatesInLab(ctx, nodes, targetLab.ID); templateValidationErr != nil {
		logger.Warnf(ctx, "fork workflow validate templates warn: %v", templateValidationErr)
	}

	// 6. 获取节点UUIDs用于获取边
	nodeUUIDs := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		return node.UUID, true
	})

	// 7. 获取工作流的边
	edges, err := w.workflowStore.GetWorkflowEdges(ctx, nodeUUIDs)
	if err != nil {
		return err
	}

	// 8. 构建节点层次结构
	preBuildNodes := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (*utils.Node[int64, *model.WorkflowNode], bool) {
		return &utils.Node[int64, *model.WorkflowNode]{
			Name:   node.ID,
			Parent: node.ParentID,
			Data:   node,
		}, true
	})

	buildNodes, err := utils.BuildHierarchy(preBuildNodes)
	if err != nil {
		return err
	}

	// 9. 创建新的工作流
	newWorkflow := &model.Workflow{
		UserID:      userInfo.ID,
		Name:        sourceWorkflow.Name,
		Description: sourceWorkflow.Description,
		LabID:       targetLab.ID,
		Published:   false, // fork的工作流默认未发布
		Tags:        sourceWorkflow.Tags,
	}

	// 10. 执行事务：创建工作流并复制所有节点和边
	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		// 创建新工作流
		if err := w.workflowStore.Create(txCtx, newWorkflow); err != nil {
			return err
		}

		old2NewIDMap := make(map[int64]int64)
		old2NewUUIDMap := make(map[uuid.UUID]uuid.UUID)

		// 复制所有节点
		for _, nodes := range buildNodes {
			newNodes := utils.FilterSlice(nodes, func(oldNode *model.WorkflowNode) (*model.WorkflowNode, bool) {
				return &model.WorkflowNode{
					WorkflowID:     newWorkflow.ID,
					WorkflowNodeID: oldNode.WorkflowNodeID,
					ParentID:       old2NewIDMap[oldNode.ParentID],
					Name:           oldNode.Name,
					UserID:         userInfo.ID,
					Status:         oldNode.Status,
					Type:           oldNode.Type,
					LabNodeType:    oldNode.LabNodeType,
					Icon:           oldNode.Icon,
					Pose:           oldNode.Pose,
					Param:          oldNode.Param,
					Footer:         oldNode.Footer,
					DeviceName:     oldNode.DeviceName,
					ActionName:     oldNode.ActionName,
					ActionType:     oldNode.ActionType,
					Disabled:       oldNode.Disabled,
					Minimized:      oldNode.Minimized,
					OldNode:        oldNode,
				}, true
			})

			if err := w.workflowStore.UpsertNodes(txCtx, newNodes); err != nil {
				return err
			}

			// 更新ID映射
			utils.Range(newNodes, func(_ int, newNode *model.WorkflowNode) bool {
				old2NewIDMap[newNode.OldNode.ID] = newNode.ID
				old2NewUUIDMap[newNode.OldNode.UUID] = newNode.UUID
				return true
			})
		}

		// 复制所有边
		newEdges := utils.FilterSlice(edges, func(edge *model.WorkflowEdge) (*model.WorkflowEdge, bool) {
			sourceNodeUUID, ok := old2NewUUIDMap[edge.SourceNodeUUID]
			if !ok || sourceNodeUUID.IsNil() {
				logger.Warnf(txCtx, "can not fork edge source node uuid: %s", edge.SourceNodeUUID)
				return nil, false
			}

			targetNodeUUID, ok := old2NewUUIDMap[edge.TargetNodeUUID]
			if !ok || targetNodeUUID.IsNil() {
				logger.Warnf(txCtx, "can not fork edge target node uuid: %s", edge.TargetNodeUUID)
				return nil, false
			}

			return &model.WorkflowEdge{
				SourceNodeUUID:   sourceNodeUUID,
				TargetNodeUUID:   targetNodeUUID,
				SourceHandleUUID: edge.SourceHandleUUID,
				TargetHandleUUID: edge.TargetHandleUUID,
			}, true
		})

		return w.workflowStore.DuplicateEdge(txCtx, newEdges)
	})

	if err != nil {
		return err
	}

	logger.Infof(ctx, "Successfully forked workflow %s to lab %s", sourceWorkflow.UUID, targetLab.UUID)
	return nil
}

// validateTemplatesInLab 验证目标实验室是否包含所有必需的模板
func (w *workflowImpl) validateTemplatesInLab(ctx context.Context, nodes []*model.WorkflowNode, targetLabID int64) error {
	// 收集所有需要的模板ID
	templateIDs := make([]int64, 0)
	for _, node := range nodes {
		if node.WorkflowNodeID > 0 {
			templateIDs = append(templateIDs, node.WorkflowNodeID)
		}
	}

	if len(templateIDs) == 0 {
		return nil // 没有模板依赖
	}

	// 查询目标实验室中的模板
	templates, err := w.workflowStore.GetWorkflowNodeTemplate(ctx, map[string]any{
		"id":     templateIDs,
		"lab_id": targetLabID,
	})
	if err != nil {
		return err
	}

	// 检查是否所有模板都存在
	templateMap := utils.Slice2Map(templates, func(t *model.WorkflowNodeTemplate) (int64, *model.WorkflowNodeTemplate) {
		return t.ID, t
	})

	missingTemplates := make([]int64, 0)
	for _, templateID := range templateIDs {
		if _, exists := templateMap[templateID]; !exists {
			missingTemplates = append(missingTemplates, templateID)
		}
	}

	if len(missingTemplates) > 0 {
		return code.WorkflowTemplateNotFoundErr.WithMsgf("missing templates in target lab: %v", missingTemplates)
	}

	return nil
}
