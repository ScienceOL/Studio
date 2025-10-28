package workflow

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olahol/melody"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/internal/config"
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
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo/tags"
	wfl "github.com/scienceol/studio/service/pkg/repo/workflow"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/datatypes"
)

type workflowImpl struct {
	workflowStore repo.WorkflowRepo
	labStore      repo.LaboratoryRepo
	materialStore repo.MaterialRepo
	tagsStore     repo.Tags
	rClient       *r.Client
	wsClient      *melody.Melody
	*schemaHelper
}

func New(ctx context.Context, wsClient *melody.Melody) workflow.Service {
	w := &workflowImpl{
		workflowStore: wfl.New(),
		labStore:      el.New(),
		wsClient:      wsClient,
		materialStore: mStore.NewMaterialImpl(),
		tagsStore:     tags.NewTag(),
		rClient:       redis.GetClient(),
		schemaHelper: &schemaHelper{
			materialStore: mStore.NewMaterialImpl(),
		},
	}
	if err := events.NewEvents().Registry(ctx, notify.WorkflowRun, w.HandleNotify); err != nil {
		logger.Errorf(ctx, "worflow Registry WorkflowRun fail err: %+v", err)
	}
	return w
}

func (w *workflowImpl) Create(ctx context.Context, data *workflow.CreateReq) (*workflow.CreateResp, error) {
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

	return &workflow.CreateResp{
		UUID:        d.UUID,
		Name:        d.Name,
		Description: d.Description,
	}, nil
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

func (w *workflowImpl) TemplateDetail(_ context.Context) {
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

func (w *workflowImpl) UpdateNodeTemplate(_ context.Context) {
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
		err = w.createGroup(ctx, s, b)
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
		err = w.batchSave(ctx, s, b)
	case workflow.RunWorkflow:
		data, err = w.runWorkflow(ctx, s, b)
	case workflow.StopWorkflow:
		data, err = w.stopWorkflow(ctx, s, b)
	case workflow.FetchWorkflowStatus:
		data, err = w.fetchWorkflowTask(ctx, s)
	case workflow.Dumplicate:
		data, err = w.duplicateNode(ctx, b)

	default:
		return common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, code.UnknownWSActionErr)
	}

	if err != nil {
		if err := common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, err); err != nil {
			logger.Errorf(ctx, "workflow ReplyWSErr err: %+v", err)
		}
		return err
	}

	if data != nil {
		return common.ReplyWSOk(s, msgType.Action, msgType.MsgUUID, data)
	}

	return common.ReplyWSOk(s, msgType.Action, msgType.MsgUUID)
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

	materialNodes := make([]*model.MaterialNode, 0, 1)
	if err := w.materialStore.FindDatas(ctx, &materialNodes, map[string]any{
		"lab_id": resp.Workflow.LabID,
	}, "id", "uuid", "name", "display_name", "type", "parent_id"); err != nil {
		return nil, err
	}

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
			schema, defaultValue := w.handleSchema(ctx, materialNodes, node.Action.Schema)
			data.Schema = schema
			if len(data.Param) == 0 && len(defaultValue) != 0 {
				data.Param, _ = json.Marshal(defaultValue)
			} else if len(data.Param) != 0 && len(defaultValue) != 0 {
				paramMap := make(map[string]any)
				if err := json.Unmarshal(data.Param, &paramMap); err == nil {
					maps.Copy(paramMap, defaultValue)
					data.Param, _ = json.Marshal(paramMap)
				}
			}

			data.TemplateUUID = node.Action.UUID
		}

		return data, true
	})

	edges := utils.FilterSlice(resp.Edges, func(edge *model.WorkflowEdge) (*workflow.WSEdge, bool) {
		return &workflow.WSEdge{
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

func (w *workflowImpl) createGroup(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[*workflow.WSGroup]{}
	err := json.Unmarshal(b, req)
	if err != nil || req.Data == nil {
		return code.ParamErr.WithErr(err)
	}

	reqData := req.Data
	if reqData == nil ||
		len(reqData.Children) == 0 ||
		reqData.Pose == datatypes.NewJSONType(model.Pose{}) {
		return code.ParamErr.WithMsg("data is empty")
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}

	wk, err := w.getWorkflow(ctx, s)
	if err != nil {
		return err
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
			ParentID: groupData.ID,
		}, []string{"parent_id"}); err != nil {
			return err
		}

		return nil
	})

	return err
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

	materialNodes := make([]*model.MaterialNode, 0, 1)
	if err := w.materialStore.FindDatas(ctx, &materialNodes, map[string]any{
		"lab_id": wk.LabID,
	}, "id", "uuid", "name", "display_name", "type", "parent_id"); err != nil {
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
		DeviceName:     w.materialStore.GetFirstDevice(ctx, deviceAction[0].ResourceNodeID),
		Param: utils.SafeValue(func() datatypes.JSON {
			if deviceAction[0].GoalDefault.String() != "" {
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

	schema, defaultValue := w.handleSchema(ctx, materialNodes, deviceAction[0].Schema)
	respData.Schema = schema
	if len(respData.Param) == 0 && len(defaultValue) != 0 {
		respData.Param, _ = json.Marshal(defaultValue)
	} else if len(respData.Param) != 0 && len(defaultValue) != 0 {
		paramMap := make(map[string]any)
		if err := json.Unmarshal(respData.Param, &paramMap); err == nil {
			maps.Copy(paramMap, defaultValue)
			res := maps.Values(paramMap)
			fmt.Println(res)

			respData.Param, _ = json.Marshal(paramMap)
		}
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
	req := &common.WSData[[]*workflow.WSEdge]{}
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

	edgeDatas := utils.FilterSlice(req.Data, func(edge *workflow.WSEdge) (*model.WorkflowEdge, bool) {
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

	respDatas := utils.FilterSlice(edgeDatas, func(data *model.WorkflowEdge) (*workflow.WSEdge, bool) {
		return &workflow.WSEdge{
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

	resp, err := w.workflowStore.DeleteWorkflowEdges(ctx, req.Data)
	if err != nil {
		return nil, code.UpsertWorkflowEdgeErr.WithErr(err)
	}

	return resp, nil
}

// 批量保存工作流节点
func (w *workflowImpl) batchSave(ctx context.Context, _ *melody.Session, b []byte) error {
	req := &common.WSData[*workflow.WSGraph]{}
	if err := json.Unmarshal(b, req); err != nil {
		return code.ParamErr.WithMsg(err.Error())
	}

	err := w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		// 保存节点
		if err := w.batchSaveNodes(txCtx, req.Data.Nodes); err != nil {
			return code.SaveWorkflowNodeErr.WithMsg(err.Error())
		}

		// 删除掉，边的关系是实时保存的
		// 保存边
		// if err := w.batchSaveEdge(txCtx, req.Data.Edges); err != nil {
		// 	return code.SaveWorkflowEdgeErr.WithMsg(err.Error())
		// }
		return nil
	})
	if err != nil {
		return err
	}

	return nil
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

	// FIXME: 修复提交工作流提前检测
	// nodes, err := w.workflowStore.GetWorkflowNodes(ctx, map[string]any{
	// 	"workflow_id": wk.ID,
	// }, "uuid", "name", "device_name")
	// if err != nil {
	// 	return nil, err
	// }
	//
	// utils.Range(nodes)

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
		conf := config.Global().Job
		data := engine.WorkflowInfo{
			Action:       engine.StartJob,
			TaskUUID:     task.UUID,
			WorkflowUUID: wk.UUID,
			LabUUID:      labUUID,
			UserID:       wk.UserID,
		}

		dataB, _ := json.Marshal(data)
		logger.Infof(ctx, "runWorkflow ============ data: %+v", data)

		ret := w.rClient.LPush(ctx, conf.JobQueueName, dataB)
		if ret.Err() != nil {
			logger.Errorf(ctx, "runWorkflow ============ send data error: %+v", ret.Err())
			return code.ParamErr.WithMsgf("push workflow redis msg err: %+v", ret.Err())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return taskUUID, nil
}

// HttpRunWorkflow 通过 HTTP 启动工作流（无鉴权）
func (w *workflowImpl) HttpRunWorkflow(ctx context.Context, req *workflow.RunReq) (uuid.UUID, error) {
	if req == nil || req.WorkflowUUID.IsNil() {
		return uuid.UUID{}, code.ParamErr.WithMsg("workflow uuid is empty")
	}

	wk, err := w.workflowStore.GetWorkflowByUUID(ctx, req.WorkflowUUID)
	if err != nil {
		return uuid.UUID{}, err
	}

	// 基于工作流记录的创建者作为 user_id（无 token 情况）
	userID := wk.UserID

	// 获取 lab uuid
	labMap := w.workflowStore.ID2UUID(ctx, &model.Laboratory{}, wk.LabID)
	labUUID, ok := labMap[wk.LabID]
	if !ok {
		return uuid.UUID{}, code.ParamErr.WithMsg("can not get lab uuid")
	}

	var taskUUID uuid.UUID
	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		task := &model.WorkflowTask{LabID: wk.LabID, WorkflowID: wk.ID, UserID: userID}
		if err := w.workflowStore.CreateWorkflowTask(txCtx, task); err != nil {
			return err
		}
		taskUUID = task.UUID

		conf := config.Global().Job
		data := engine.WorkflowInfo{
			Action:       engine.StartJob,
			TaskUUID:     task.UUID,
			WorkflowUUID: wk.UUID,
			LabUUID:      labUUID,
			UserID:       userID,
		}
		dataB, _ := json.Marshal(data)
		ret := w.rClient.LPush(ctx, conf.JobQueueName, dataB)
		if ret.Err() != nil {
			logger.Errorf(ctx, "http runWorkflow ============ send data error: %+v", ret.Err())
			return code.ParamErr.WithMsgf("push workflow redis msg err: %+v", ret.Err())
		}
		return nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	return taskUUID, nil
}

func (w *workflowImpl) stopWorkflow(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	req := &common.WSData[uuid.UUID]{}
	if err := json.Unmarshal(b, req); err != nil || req.Data.IsNil() {
		return nil, code.ParamErr
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
	case model.WorkflowTaskStatusCanceled, model.WorkflowTaskStatusFailed, model.WorkflowTaskStatusSuccessed:
		return nil, code.WorkflowTaskFinished
	default:
		return nil, code.WorkflowTaskStatusErr
	}

	conf := config.Global().Job
	data := engine.WorkflowInfo{
		Action:       engine.StopJob,
		TaskUUID:     req.Data,
		WorkflowUUID: wk.UUID,
		LabUUID:      labUUID,
		UserID:       wk.UserID,
	}

	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		task := tasks[0]
		task.Status = model.WorkflowTaskStatusCanceled
		task.UpdatedAt = time.Now()
		if err := w.workflowStore.UpdateData(txCtx, task, map[string]any{
			"uuid": req.Data,
		}, "status", "updated_at"); err != nil {
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

func (w *workflowImpl) duplicateNode(ctx context.Context, b []byte) (any, error) {
	req := &common.WSData[uuid.UUID]{}
	if err := json.Unmarshal(b, req); err != nil || req.Data.IsNil() {
		return nil, code.ParamErr
	}

	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin.WithMsg("can not get user info")
	}

	sourceNode := &model.WorkflowNode{}
	if err := w.workflowStore.GetData(ctx, sourceNode, map[string]any{
		"uuid": req.Data,
	}); err != nil {
		return nil, err
	}

	if sourceNode.Type == model.WorkflowNodeGroup && sourceNode.WorkflowNodeID <= 0 {
		return nil, code.ParamErr.WithMsgf("source node param err source uuid: %+s", req.Data)
	}

	tpl := &model.WorkflowNodeTemplate{}
	var handles []*model.WorkflowHandleTemplate
	var err error
	if sourceNode.WorkflowNodeID > 0 {
		if err = w.workflowStore.GetData(ctx, tpl, map[string]any{
			"id": sourceNode.WorkflowNodeID,
		}); err != nil {
			logger.Errorf(ctx, "duplicateNode can not get workflow template data id: %d, err: %+v", sourceNode.WorkflowNodeID, err)
		}
	}

	if tpl.ID > 0 {
		handles, err = w.workflowStore.GetWorkflowHandleTemplates(ctx, []int64{tpl.ID})
		if err != nil {
			logger.Errorf(ctx, "duplicateNode can not get workflow handle template id: %+d", tpl.ID)
		}
	}

	newNode := &model.WorkflowNode{
		WorkflowID:     sourceNode.WorkflowID,
		WorkflowNodeID: sourceNode.WorkflowNodeID,
		ParentID:       sourceNode.ParentID,
		Name:           sourceNode.Name,
		UserID:         userInfo.ID,
		Status:         sourceNode.Status,
		Type:           sourceNode.Type,
		LabNodeType:    sourceNode.LabNodeType,
		Icon:           sourceNode.Icon,
		Pose:           sourceNode.Pose,
		Param:          sourceNode.Param,
		Footer:         sourceNode.Footer,
		DeviceName:     sourceNode.DeviceName,
		ActionName:     sourceNode.ActionName,
		ActionType:     sourceNode.ActionType,
		Disabled:       false,
		Minimized:      false,
	}

	if err := w.workflowStore.CreateNode(ctx, newNode); err != nil {
		return nil, err
	}
	var parentUUID uuid.UUID

	resData := &workflow.WSNode{
		UUID:        newNode.UUID,
		Name:        newNode.Name,
		ParentUUID:  parentUUID,
		UserID:      newNode.UserID,
		Status:      newNode.Status,
		Type:        newNode.Type,
		Icon:        newNode.Icon,
		Pose:        newNode.Pose,
		Param:       newNode.Param,
		Footer:      newNode.Footer,
		DeviceName:  newNode.DeviceName,
		Disabled:    newNode.Disabled,
		Minimized:   newNode.Minimized,
		LabNodeType: newNode.LabNodeType,
		Handles: utils.FilterSlice(handles, func(h *model.WorkflowHandleTemplate) (*workflow.WSNodeHandle, bool) {
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
	if tpl.ID > 0 {
		resData.TemplateUUID = tpl.UUID
		resData.Schema = tpl.Schema
	}

	return resData, nil
}

func (w *workflowImpl) DuplicateWorkflow(ctx context.Context, req *workflow.DuplicateReq) (*workflow.DuplicateRes, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	sourceWorkflow, err := w.workflowStore.GetWorkflowByUUID(ctx, req.SourceUUID)
	if err != nil {
		return nil, err
	}

	nodes, err := w.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": sourceWorkflow.ID,
	})
	if err != nil {
		return nil, err
	}

	targetLabID := sourceWorkflow.LabID
	if !req.TargetLabUUID.IsNil() {
		targetLabID = w.workflowStore.UUID2ID(ctx, &model.Laboratory{}, req.TargetLabUUID)[req.TargetLabUUID]
		if targetLabID == 0 {
			return nil, code.LabNotFound
		}
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

	old2NewIDMap := make(map[int64]int64)
	old2NewUUIDMap := make(map[uuid.UUID]uuid.UUID)
	var newWK *model.Workflow

	sourceTargetTplIDMap := utils.Slice2Map(nodes, func(node *model.WorkflowNode) (int64, int64) {
		return node.WorkflowNodeID, node.WorkflowNodeID
	})

	edgeHandleUUIDMap := make(map[uuid.UUID]uuid.UUID)
	utils.Range(edges, func(_ int, edge *model.WorkflowEdge) bool {
		edgeHandleUUIDMap[edge.SourceHandleUUID] = edge.SourceHandleUUID
		edgeHandleUUIDMap[edge.TargetHandleUUID] = edge.TargetHandleUUID
		return true
	})

	var diff []*workflow.DuplicateError
	if sourceWorkflow.LabID != targetLabID {
		sourceTargetTplIDMap, diff, err = w.getMappingTemplate(ctx, sourceWorkflow.LabID, targetLabID, nodes)
		if err != nil {
			return nil, err
		}

		if len(diff) > 0 {
			return &workflow.DuplicateRes{
				Errors: diff,
			}, nil
		}

		edgeHandleUUIDMap, err = w.getEdgeMapping(ctx, sourceTargetTplIDMap, edges)
		if err != nil {
			return nil, err
		}
	}

	err = w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		newWK = &model.Workflow{
			Name:   utils.Or(req.Name, "Untitled"),
			UserID: userInfo.ID,
			LabID:  targetLabID,
		}
		if err := w.workflowStore.Create(txCtx, newWK); err != nil {
			return err
		}

		for _, nodes := range buildNodes {
			newNodes := utils.FilterSlice(nodes, func(oldNode *model.WorkflowNode) (*model.WorkflowNode, bool) {
				return &model.WorkflowNode{
					WorkflowID:     newWK.ID,
					WorkflowNodeID: sourceTargetTplIDMap[oldNode.WorkflowNodeID],
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
				SourceHandleUUID: edgeHandleUUIDMap[edge.SourceHandleUUID],
				TargetHandleUUID: edgeHandleUUIDMap[edge.TargetHandleUUID],
			}, true
		})

		return w.workflowStore.DuplicateEdge(txCtx, newEdges)
	})
	if err != nil {
		return nil, err
	}

	return &workflow.DuplicateRes{
		UUID: newWK.UUID,
		Name: newWK.Name,
	}, nil
}

func (w *workflowImpl) getEdgeMapping(ctx context.Context, stWorkflowTplMap map[int64]int64, edges []*model.WorkflowEdge) (map[uuid.UUID]uuid.UUID, error) {
	sourceHandleUUIDs := make([]uuid.UUID, 0, 2*len(edges))
	utils.Range(edges, func(_ int, edge *model.WorkflowEdge) bool {
		sourceHandleUUIDs = utils.AppendUniqSlice(sourceHandleUUIDs, edge.SourceHandleUUID, edge.TargetHandleUUID)
		return true
	})

	workflowHandleDatas := make([]*model.WorkflowHandleTemplate, 0, len(sourceHandleUUIDs))
	if err := w.workflowStore.FindDatas(ctx, &workflowHandleDatas, map[string]any{
		"uuid": sourceHandleUUIDs,
	}); err != nil {
		return nil, err
	}

	workflowNodeTplIDs := make([]int64, 0, len(workflowHandleDatas))
	handleKeys := make([]string, 0, len(workflowHandleDatas))
	ioTypes := make([]string, 0, len(workflowHandleDatas))
	sourceWorkflowHandleMap := make(map[string]*model.WorkflowHandleTemplate)
	utils.Range(workflowHandleDatas, func(_ int, handle *model.WorkflowHandleTemplate) bool {
		workflowNodeTplIDs = utils.AppendUniqSlice(workflowNodeTplIDs, stWorkflowTplMap[handle.WorkflowNodeID])
		handleKeys = utils.AppendUniqSlice(handleKeys, handle.HandleKey)
		ioTypes = utils.AppendUniqSlice(ioTypes, handle.IoType)
		sourceWorkflowHandleMap[fmt.Sprintf("%d-%s-%s",
			handle.WorkflowNodeID,
			handle.HandleKey, handle.IoType)] = handle
		return true
	})

	targetWorkflowHandleDatas := make([]*model.WorkflowHandleTemplate, 0, len(sourceHandleUUIDs))
	if err := w.workflowStore.FindDatas(ctx, &targetWorkflowHandleDatas, map[string]any{
		"workflow_node_id": workflowNodeTplIDs,
		"handle_key":       handleKeys,
		"io_type":          ioTypes,
	}); err != nil {
		return nil, err
	}

	targetWorkflowHandleMap := make(map[string]*model.WorkflowHandleTemplate)
	utils.Range(targetWorkflowHandleDatas, func(_ int, handle *model.WorkflowHandleTemplate) bool {
		targetWorkflowHandleMap[fmt.Sprintf("%d-%s-%s",
			handle.WorkflowNodeID,
			handle.HandleKey, handle.IoType)] = handle
		return true
	})

	sourceTargetUUIDMap := make(map[uuid.UUID]uuid.UUID)
	utils.RangeMap(sourceWorkflowHandleMap, func(key string, value *model.WorkflowHandleTemplate) bool {
		targetKey := fmt.Sprintf("%d-%s-%s",
			stWorkflowTplMap[value.WorkflowNodeID],
			value.HandleKey, value.IoType)

		targetHandle, ok := targetWorkflowHandleMap[targetKey]
		if !ok {
			logger.Warnf(ctx, "can not found target handle key: %s", targetKey)
			return true
		}
		sourceTargetUUIDMap[value.UUID] = targetHandle.UUID
		return true
	})

	return sourceTargetUUIDMap, nil
}

func (w *workflowImpl) getMappingTemplate(ctx context.Context, sourceLabID, targetLabID int64, nodes []*model.WorkflowNode) (map[int64]int64, []*workflow.DuplicateError, error) {
	sourceWorkflowTplIDs := utils.FilterUniqSlice(nodes, func(node *model.WorkflowNode) (int64, bool) {
		if node.WorkflowNodeID == 0 {
			return 0, false
		}
		return node.WorkflowNodeID, true
	})

	// 获取源节点的所有模板 id
	sourceWorkflowTplDatas := make([]*model.WorkflowNodeTemplate, 0, 1)
	if err := w.workflowStore.FindDatas(ctx, &sourceWorkflowTplDatas, map[string]any{
		"id": sourceWorkflowTplIDs,
	}, "id", "name", "resource_node_id"); err != nil {
		return nil, nil, err
	}

	sourceResourceNodeTplIDs := make([]int64, 0, 1)
	sourceWorkflowTplNames := make([]string, 0, 1)
	// 获取所有的资源 id 和节点模板名
	utils.Range(sourceWorkflowTplDatas, func(_ int, node *model.WorkflowNodeTemplate) bool {
		sourceResourceNodeTplIDs = utils.AppendUniqSlice(sourceResourceNodeTplIDs, node.ResourceNodeID)
		sourceWorkflowTplNames = utils.AppendUniqSlice(sourceWorkflowTplNames, node.Name)
		return true
	})

	// 获取所有的资源数据
	sourceResourceNodeTplDatas := make([]*model.ResourceNodeTemplate, 0, 1)
	if err := w.workflowStore.FindDatas(ctx, &sourceResourceNodeTplDatas, map[string]any{
		"id": sourceResourceNodeTplIDs,
	}, "id", "name"); err != nil {
		return nil, nil, err
	}

	// 比对目标节点是否有模板节点
	// 获取所有的资源模板名
	targetResourceNodeNames := utils.FilterSlice(sourceResourceNodeTplDatas, func(resTpl *model.ResourceNodeTemplate) (string, bool) {
		return resTpl.Name, true
	})
	// 获取所有的模板名对应的数据
	targetResourceNodeDatas := make([]*model.ResourceNodeTemplate, 0, 1)
	if err := w.workflowStore.FindDatas(ctx, &targetResourceNodeDatas, map[string]any{
		"lab_id": targetLabID,
		"name":   targetResourceNodeNames,
	}, "id", "name"); err != nil {
		return nil, nil, err
	}

	targetResourceNodeIDs := utils.FilterSlice(targetResourceNodeDatas, func(resTpl *model.ResourceNodeTemplate) (int64, bool) {
		return resTpl.ID, true
	})

	// 获取所有的目标节点名
	targetWorkflowTplDatas := make([]*model.WorkflowNodeTemplate, 0, 1)
	if err := w.workflowStore.FindDatas(ctx, &targetWorkflowTplDatas, map[string]any{
		"lab_id":           targetLabID,
		"resource_node_id": targetResourceNodeIDs,
		"name":             sourceWorkflowTplNames,
	}, "id", "name", "resource_node_id"); err != nil {
		return nil, nil, err
	}

	sourceResourceTplMap := utils.Slice2Map(sourceResourceNodeTplDatas, func(node *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return node.ID, node
	})

	sourceNameMap := make(map[string]*workflow.TplMapping)
	sourceIDMap := make(map[int64]*workflow.TplMapping)
	utils.Range(sourceWorkflowTplDatas, func(_ int, node *model.WorkflowNodeTemplate) bool {
		res, ok := sourceResourceTplMap[node.ResourceNodeID]
		if ok {
			sourceNameMap[res.Name+node.Name] = &workflow.TplMapping{
				WorkflowTpl: node,
				ResourceTpl: res,
			}
			sourceIDMap[node.ID] = &workflow.TplMapping{
				WorkflowTpl: node,
				ResourceTpl: res,
			}
		}
		return true
	})

	targetResourceTplMap := utils.Slice2Map(targetResourceNodeDatas, func(node *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return node.ID, node
	})

	targetNameMap := make(map[string]*workflow.TplMapping)
	targetIDMap := make(map[int64]*workflow.TplMapping)
	utils.Range(targetWorkflowTplDatas, func(_ int, node *model.WorkflowNodeTemplate) bool {
		res, ok := targetResourceTplMap[node.ResourceNodeID]
		if ok {
			targetNameMap[res.Name+node.Name] = &workflow.TplMapping{
				WorkflowTpl: node,
				ResourceTpl: res,
			}
			targetIDMap[node.ID] = &workflow.TplMapping{
				WorkflowTpl: node,
				ResourceTpl: res,
			}
		}
		return true
	})

	diffMap := utils.SetDifference(sourceNameMap, targetNameMap)
	if len(diffMap) > 0 {
		sourceLabData, _ := w.labStore.GetLabByID(ctx, sourceLabID, "id", "name")
		targetLabData, _ := w.labStore.GetLabByID(ctx, targetLabID, "id", "name")

		return nil, utils.MapToSlice(diffMap, func(key string, value *workflow.TplMapping) (*workflow.DuplicateError, bool) {
			sourceName := fmt.Sprintf("source lab name: %s resource name: %s action name: %s",
				sourceLabData.Name, value.ResourceTpl.Name, value.WorkflowTpl.Name)
			targetName := fmt.Sprintf("target lab name: %s resource name：%s action name: %s",
				targetLabData.Name, value.ResourceTpl.Name, value.WorkflowTpl.Name)
			return &workflow.DuplicateError{
				SourceTemplateName: sourceName,
				TargetTemplateName: targetName,
				Reason:             "target lab need " + value.WorkflowTpl.Name,
			}, true
		}), nil
	}

	sourceTargetIDMap := make(map[int64]int64)
	utils.RangeMap(sourceNameMap, func(name string, value *workflow.TplMapping) bool {
		v := targetNameMap[name]
		sourceTargetIDMap[value.WorkflowTpl.ID] = v.WorkflowTpl.ID
		return true
	})

	return sourceTargetIDMap, nil, nil
}

func (w *workflowImpl) batchSaveNodes(ctx context.Context, nodes []*workflow.WSNode) error {
	dbNodes, err := utils.FilterSliceWithErr(nodes, func(node *workflow.WSNode) ([]*model.WorkflowNode, bool, error) {
		if node.UUID.IsNil() {
			return nil, false, code.ParamErr.WithMsg("uuid is empty")
		}
		data := &model.WorkflowNode{
			// Icon:       node.Icon,
			Pose: node.Pose,
			// Param:      node.Param,
			// Footer:     node.Footer,
			// DeviceName: node.DeviceName,
			// Disabled:   node.Disabled,
			// Minimized:  node.Minimized,
		}
		data.UUID = node.UUID
		data.UpdatedAt = time.Now()
		return []*model.WorkflowNode{data}, true, nil
	})
	if err != nil {
		return err
	}

	return w.workflowStore.UpsertNodes(ctx, dbNodes, "pose", "updated_at")
}

func (w *workflowImpl) batchSaveEdge(ctx context.Context, edges []*workflow.WSEdge) error {
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

	workflowEdges := utils.FilterSlice(edges, func(edge *workflow.WSEdge) (*model.WorkflowEdge, bool) {
		return &model.WorkflowEdge{
			BaseModel:        model.BaseModel{UUID: edge.UUID},
			SourceNodeUUID:   edge.SourceNodeUUID,
			TargetNodeUUID:   edge.TargetNodeUUID,
			SourceHandleUUID: edge.SourceHandleUUID,
			TargetHandleUUID: edge.TargetHandleUUID,
		}, true
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
func (w *workflowImpl) GetWorkflowList(ctx context.Context, req *workflow.ListReq) (*workflow.ListResult, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	// 获取实验室ID
	var labID int64
	if !req.LabUUID.IsNil() {
		lab, err := w.labStore.GetLabByUUID(ctx, req.LabUUID)
		if err != nil {
			return nil, err
		}
		labID = lab.ID
	}

	// 从数据库获取工作流列表
	workflows, total, err := w.workflowStore.GetWorkflowList(ctx, "", labID, &req.PageReq)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	respList := utils.FilterSlice(workflows, func(wf *model.Workflow) (*workflow.ListResp, bool) {
		return &workflow.ListResp{
			UUID:        wf.UUID,
			Name:        wf.Name,
			Description: wf.Description,
			UserID:      wf.UserID,
			Published:   wf.Published,
			Tags:        []string(wf.Tags),
		}, true
	})

	hasMore := int64(req.Page*req.PageSize) < total
	return &workflow.ListResult{
		HasMore: hasMore,
		Data:    respList,
	}, nil
}

// GetWorkflowDetail 获取工作流详情
func (w *workflowImpl) GetWorkflowDetail(ctx context.Context, req *workflow.DetailReq) (*workflow.DetailResp, error) {
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

	id2UUIDMap := utils.Slice2Map(wfNodes, func(node *model.WorkflowNode) (int64, uuid.UUID) {
		return node.ID, node.UUID
	})

	nodeUUIDs := utils.FilterSlice(wfNodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		if node.Type == model.WorkflowNodeGroup {
			return uuid.UUID{}, false
		}
		return node.UUID, true
	})
	edges := make([]*model.WorkflowEdge, 0, 2*len(nodeUUIDs))
	if err := w.workflowStore.FindDatas(ctx, &edges, map[string]any{
		"source_node_uuid": nodeUUIDs,
		"target_node_uuid": nodeUUIDs,
	}); err != nil {
		return nil, err
	}

	return &workflow.DetailResp{
		UUID:        wf.UUID,
		Name:        wf.Name,
		Description: wf.Description,
		UserID:      wf.UserID,
		Nodes: utils.FilterSlice(wfNodes, func(node *model.WorkflowNode) (*workflow.WSNode, bool) {
			return &workflow.WSNode{
				UUID:       node.UUID,
				ParentUUID: id2UUIDMap[node.ParentID],
				Name:       node.Name,
				UserID:     node.UserID,
				Status:     node.Status,
				Type:       node.Type,
				Icon:       node.Icon,
				Pose:       node.Pose,
				Param:      node.Param,
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
		Edges: utils.FilterSlice(edges, func(edge *model.WorkflowEdge) (*workflow.WSEdge, bool) {
			return &workflow.WSEdge{
				UUID:             edge.UUID,
				SourceNodeUUID:   edge.SourceNodeUUID,
				TargetNodeUUID:   edge.TargetNodeUUID,
				SourceHandleUUID: edge.SourceHandleUUID,
				TargetHandleUUID: edge.TargetHandleUUID,
			}, true
		}),
	}, nil
}

// 导出工作流：包含节点与边，并携带模板/资源的名称以便跨实验室匹配
func (w *workflowImpl) ExportWorkflow(ctx context.Context, req *workflow.ExportReq) (*workflow.ExportData, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	wk, err := w.workflowStore.GetWorkflowByUUID(ctx, req.UUID)
	if err != nil {
		return nil, err
	}

	// 校验权限：只能导出自己的
	// if wk.UserID != userInfo.ID {
	// 	return nil, code.NoPermission
	// }

	nodes, err := w.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": wk.ID,
	})
	if err != nil {
		return nil, err
	}

	// 拉取模板与资源信息
	tplIDs := utils.FilterSlice(nodes, func(n *model.WorkflowNode) (int64, bool) {
		if n.WorkflowNodeID > 0 {
			return n.WorkflowNodeID, true
		}
		return 0, false
	})
	tplList, err := w.workflowStore.GetWorkflowNodeTemplate(ctx, map[string]any{
		"id": tplIDs,
	})
	if err != nil {
		return nil, err
	}
	tplMap := utils.Slice2Map(tplList, func(t *model.WorkflowNodeTemplate) (int64, *model.WorkflowNodeTemplate) {
		return t.ID, t
	})

	resIDs := utils.FilterSlice(tplList, func(t *model.WorkflowNodeTemplate) (int64, bool) {
		return t.ResourceNodeID, true
	})
	resNodes, err := w.labStore.GetResourceNodeTemplates(ctx, resIDs)
	if err != nil {
		return nil, err
	}
	resMap := utils.Slice2Map(resNodes, func(r *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return r.ID, r
	})

	// 句柄模板
	handles, err := w.workflowStore.GetWorkflowHandleTemplates(ctx, tplIDs)
	if err != nil {
		return nil, err
	}
	handleMap := utils.SliceToMapSlice(handles, func(h *model.WorkflowHandleTemplate) (int64, *model.WorkflowHandleTemplate, bool) {
		return h.WorkflowNodeID, h, true
	})

	// 构建导出节点
	id2uuid := utils.Slice2Map(nodes, func(n *model.WorkflowNode) (int64, uuid.UUID) { return n.ID, n.UUID })
	exportNodes := utils.FilterSlice(nodes, func(n *model.WorkflowNode) (*workflow.ExportNode, bool) {
		parentUUID := uuid.UUID{}
		if n.ParentID > 0 {
			parentUUID = id2uuid[n.ParentID]
		}
		tpl := tplMap[n.WorkflowNodeID]
		var tplUUID uuid.UUID
		var tplName string
		var resName string
		if tpl != nil {
			tplUUID = tpl.UUID
			tplName = tpl.Name
			if r := resMap[tpl.ResourceNodeID]; r != nil {
				resName = r.Name
			}
		}
		return &workflow.ExportNode{
			UUID:         n.UUID,
			ParentUUID:   parentUUID,
			Name:         n.Name,
			Type:         n.Type,
			Icon:         n.Icon,
			Pose:         n.Pose,
			Param:        n.Param,
			Footer:       n.Footer,
			DeviceName:   n.DeviceName,
			Disabled:     n.Disabled,
			Minimized:    n.Minimized,
			LabNodeType:  n.LabNodeType,
			TemplateUUID: tplUUID,
			TemplateName: tplName,
			ResourceName: resName,
		}, true
	})

	nodeUUIDs := utils.FilterSlice(nodes, func(n *model.WorkflowNode) (uuid.UUID, bool) { return n.UUID, true })
	edges, err := w.workflowStore.GetWorkflowEdges(ctx, nodeUUIDs)
	if err != nil {
		return nil, err
	}

	// 为边附上可匹配的句柄key与io
	// 需要模板->句柄映射: handleMap 已有；同时需要句柄UUID->(key, io)映射
	handleUUID2info := make(map[uuid.UUID]struct{ key, io string })
	for _, hs := range handleMap {
		for _, h := range hs {
			handleUUID2info[h.UUID] = struct{ key, io string }{key: h.HandleKey, io: h.IoType}
		}
	}
	exportEdges := utils.FilterSlice(edges, func(e *model.WorkflowEdge) (*workflow.ExportEdge, bool) {
		sh := handleUUID2info[e.SourceHandleUUID]
		th := handleUUID2info[e.TargetHandleUUID]
		return &workflow.ExportEdge{
			SourceNodeUUID:  e.SourceNodeUUID,
			TargetNodeUUID:  e.TargetNodeUUID,
			SourceHandleKey: sh.key,
			SourceHandleIO:  sh.io,
			TargetHandleKey: th.key,
			TargetHandleIO:  th.io,
		}, true
	})

	return &workflow.ExportData{
		WorkflowUUID: wk.UUID,
		WorkflowName: wk.Name,
		Tags:         []string(wk.Tags),
		Nodes:        exportNodes,
		Edges:        exportEdges,
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

func (w *workflowImpl) WorkflowTaskList(ctx context.Context,
	req *workflow.TaskReq) (*common.PageMoreResp[[]*workflow.TaskResp],
	error,
) {
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
			UserID:     "",
			LabID:      wk.LabID,
			WrokflowID: wk.ID,
		},
	})
	if err != nil {
		return nil, err
	}

	return &common.PageMoreResp[[]*workflow.TaskResp]{
		HasMore:  resp.HasMore,
		Page:     resp.Page,
		PageSize: resp.PageSize,
		Data: utils.FilterSlice(resp.Data, func(task *model.WorkflowTask) (*workflow.TaskResp, bool) {
			return &workflow.TaskResp{
				UUID:       task.UUID,
				Status:     task.Status,
				CreatedAt:  task.CreatedAt,
				FinishedAt: task.FinishedTime,
			}, true
		}),
	}, nil
}

func (w *workflowImpl) TaskDownload(ctx context.Context, req *workflow.TaskDownloadReq) (*bytes.Buffer, error) {
	taskID := w.workflowStore.UUID2ID(ctx, &model.WorkflowTask{}, req.UUID)[req.UUID]
	if taskID <= 0 {
		return nil, code.WorkflowTaskNotFoundErr
	}

	jobs := make([]*model.WorkflowNodeJob, 0, 2)
	if err := w.workflowStore.FindDatas(ctx, &jobs, map[string]any{
		"workflow_task_id": taskID,
	}); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入CSV头部
	header := []string{"ID", "状态", "数据", "更新时间", "创建时间"}
	if err := writer.Write(header); err != nil {
		return nil, code.FormatCSVTaskErr
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].ID < jobs[j].ID
	})

	for _, j := range jobs {
		returnInfoStr := ""
		if returnInfoBytes, err := json.Marshal(j.ReturnInfo); err == nil {
			returnInfoStr = string(returnInfoBytes)
		}
		if err := writer.Write([]string{
			strconv.FormatInt(j.ID, 10),
			string(j.Status),
			returnInfoStr,
			j.UpdatedAt.Format(time.DateTime),
			j.CreatedAt.Format(time.DateTime),
		}); err != nil {
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

	// if userInfo.ID != wk.UserID {
	// 	return code.NoPermission
	// }

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

	// 当发布为模板时，将工作流的 tags 写入 tags 表
	if req.Published != nil && *req.Published {
		if len(wk.Tags) > 0 && w.tagsStore != nil {
			tagModels := make([]*model.Tags, 0, len(wk.Tags))
			for _, name := range wk.Tags {
				tagModels = append(tagModels, &model.Tags{Type: model.WorkflowTemplateTag, Name: name})
			}
			if err := w.tagsStore.UpsertTags(ctx, tagModels); err != nil {
				return err
			}
		}
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

	// if wf.UserID != userInfo.ID {
	// 	return code.NoPermission
	// }

	return w.workflowStore.DelWorkflow(ctx, wf.ID)
}

func (w *workflowImpl) WorkflowTemplateList(ctx context.Context,
	req *workflow.TemplateListReq) (*common.PageResp[[]*workflow.TemplateListRes],
	error,
) {
	res, err := w.workflowStore.GetWorkflow(ctx, &common.PageReqT[*repo.QueryWorkflow]{
		PageReq: req.PageReq,
		Data: &repo.QueryWorkflow{
			Tags: req.Tags,
		},
	})
	if err != nil {
		return nil, err
	}

	return &common.PageResp[[]*workflow.TemplateListRes]{
		Total:    res.Total,
		Page:     res.Page,
		PageSize: res.PageSize,
		Data: utils.FilterSlice(res.Data, func(item *model.Workflow) (*workflow.TemplateListRes, bool) {
			return &workflow.TemplateListRes{
				UUID:      item.UUID,
				Name:      item.Name,
				Tags:      item.Tags,
				UserID:    item.UserID,
				CreatedAt: item.CreatedAt,
			}, true
		}),
	}, nil
}

func (w *workflowImpl) WorkflowTemplateTags(ctx context.Context) ([]string, error) {
	return w.workflowStore.GetTemplateTags(ctx, model.WorkflowTemplateTag)
}

// WorkflowTemplateTagsByLab 按实验室获取工作流模板标签
func (w *workflowImpl) WorkflowTemplateTagsByLab(ctx context.Context, req *workflow.TemplateTagsReq) ([]string, error) {
	labID := w.labStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
	if labID == 0 {
		return nil, code.CanNotGetLabIDErr
	}
	// 从 workflow.tags 聚合当前实验室下已发布工作流的标签
	return w.workflowStore.GetWorkflowTagsByLab(ctx, labID)
}

func (w *workflowImpl) ForkWrokflow(ctx context.Context, req *workflow.ForkReq) error {
	panic("not implemented") // TODO: Implement
}

// 导入工作流
func (w *workflowImpl) ImportWorkflow(ctx context.Context, req *workflow.ImportReq) (*workflow.CreateResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	if req.Data == nil || len(req.Data.Nodes) == 0 {
		return nil, code.ParamErr.WithMsg("import data is empty")
	}

	lab, err := w.labStore.GetLabByUUID(ctx, req.TargetLabUUID)
	if err != nil {
		return nil, err
	}

	resourceNames := utils.FilterUniqSlice(req.Data.Nodes, func(n *workflow.ExportNode) (string, bool) {
		return n.ResourceName, n.ResourceName != ""
	})

	resNodes := make([]*model.ResourceNodeTemplate, 0, 1)
	if err := w.labStore.FindDatas(ctx, &resNodes, map[string]any{
		"lab_id": lab.ID,
		"name":   resourceNames,
	}, "id", "name", "uuid"); err != nil {
		return nil, err
	}

	resName2ID := utils.Slice2Map(resNodes, func(r *model.ResourceNodeTemplate) (string, int64) { return r.Name, r.ID })
	// 只保留导入数据涉及到的资源ID集合
	resIDs := utils.FilterSlice(resNodes, func(r *model.ResourceNodeTemplate) (int64, bool) { return r.ID, true })
	nodeTpls, err := w.workflowStore.GetWorkflowNodeTemplate(ctx,
		map[string]any{
			"lab_id":           lab.ID,
			"resource_node_id": resIDs,
		})
	if err != nil {
		return nil, err
	}

	type tplKey struct {
		resID int64
		name  string
	}
	tplIndex := utils.Slice2Map(nodeTpls, func(t *model.WorkflowNodeTemplate) (tplKey, *model.WorkflowNodeTemplate) {
		return tplKey{resID: t.ResourceNodeID, name: t.Name}, t
	})

	// 预检：建立旧节点UUID -> 目标模板ID的映射，若缺失直接返回
	oldNodeUUID2TplID := make(map[uuid.UUID]*model.WorkflowNodeTemplate)
	for _, n := range req.Data.Nodes {
		if n.Type != model.WorkflowNodeGroup && n.ResourceName != "" && n.TemplateName != "" {
			resID, ok := resName2ID[n.ResourceName]
			if !ok || resID == 0 {
				return nil, code.TemplateNodeNotFoundErr.WithMsgf("节点 '%s' 导入失败: 资源 '%s' 在目标实验室中不存在，请确保目标实验室已配置该资源",
					n.Name, n.ResourceName)
			}
			if tpl := tplIndex[tplKey{resID: resID, name: n.TemplateName}]; tpl != nil {
				oldNodeUUID2TplID[n.UUID] = tpl
			} else {
				return nil, code.TemplateNodeNotFoundErr.WithMsgf("节点 '%s' 导入失败: 在资源 '%s' 中找不到模板 '%s'，请确保目标实验室已配置该模板",
					n.Name, n.ResourceName, n.TemplateName)
			}
		}
	}

	handles, err := w.workflowStore.GetWorkflowHandleTemplates(ctx, utils.FilterSlice(nodeTpls, func(t *model.WorkflowNodeTemplate) (int64, bool) { return t.ID, true }))
	if err != nil {
		return nil, err
	}
	handleIndex := utils.SliceToMapSlice(handles, func(h *model.WorkflowHandleTemplate) (int64, *model.WorkflowHandleTemplate, bool) {
		return h.WorkflowNodeID, h, true
	})

	newName := utils.Or(req.Data.WorkflowName, "Untitled")

	var resp *workflow.CreateResp
	if err := w.workflowStore.ExecTx(ctx, func(txCtx context.Context) error {
		wk := &model.Workflow{UserID: userInfo.ID, LabID: lab.ID, Name: newName}
		if req.Data.Published != nil {
			wk.Published = *req.Data.Published
		}
		// 写入导入的 tags 到 workflow.tags
		if len(req.Data.Tags) > 0 {
			wk.Tags = datatypes.NewJSONSlice(req.Data.Tags)
		}
		if err := w.workflowStore.Create(txCtx, wk); err != nil {
			return err
		}

		oldUUID2newUUID := make(map[uuid.UUID]uuid.UUID)
		oldUUID2newID := make(map[uuid.UUID]int64)
		remaining := make(map[uuid.UUID]*workflow.ExportNode)
		for _, n := range req.Data.Nodes {
			remaining[n.UUID] = n
		}

		cycleDetected := false
		var cycleNodes []string
		for len(remaining) > 0 {
			progressed := false
			for oldU, n := range remaining {
				if n.ParentUUID.IsNil() || oldUUID2newID[n.ParentUUID] != 0 {
					var tplID int64
					var actionName string
					var actionType string
					if tpl, ok := oldNodeUUID2TplID[n.UUID]; ok {
						tplID = tpl.ID
						actionName = tpl.Name
						actionType = tpl.Type
					}
					parentID := int64(0)
					if !n.ParentUUID.IsNil() {
						parentID = oldUUID2newID[n.ParentUUID]
					}
					node := &model.WorkflowNode{
						WorkflowID:     wk.ID,
						WorkflowNodeID: tplID,
						ParentID:       parentID,
						Name:           n.Name,
						UserID:         userInfo.ID,
						Status:         "draft",
						Type:           n.Type,
						LabNodeType:    n.LabNodeType,
						Icon:           n.Icon,
						Pose:           n.Pose,
						Param:          n.Param,
						Footer:         n.Footer,
						DeviceName:     n.DeviceName,
						Disabled:       n.Disabled,
						Minimized:      n.Minimized,
						ActionName:     actionName,
						ActionType:     actionType,
					}
					if err := w.workflowStore.CreateNode(txCtx, node); err != nil {
						return err
					}
					oldUUID2newUUID[oldU] = node.UUID
					oldUUID2newID[oldU] = node.ID
					delete(remaining, oldU)
					progressed = true
				} else {
					if !cycleDetected {
						cycleNodes = append(cycleNodes, fmt.Sprintf("节点 '%s' (UUID: %s)", n.Name, n.UUID))
					}
				}
			}
			if !progressed {
				cycleDetected = true
				if len(cycleNodes) > 0 {
					return code.ParamErr.WithMsgf("检测到循环依赖，以下节点可能存在循环引用关系: %s", strings.Join(cycleNodes, ", "))
				}
				return code.ParamErr.WithMsg("节点间存在循环依赖关系，请检查节点的父子关系")
			}
		}

		newNodes, err := w.workflowStore.GetWorkflowNodes(txCtx, map[string]any{"workflow_id": wk.ID})
		if err != nil {
			return err
		}
		newUUID2tplID := utils.Slice2Map(newNodes, func(n *model.WorkflowNode) (uuid.UUID, int64) { return n.UUID, n.WorkflowNodeID })

		edgesToCreate := make([]*model.WorkflowEdge, 0, len(req.Data.Edges))
		for _, e := range req.Data.Edges {
			sNew := oldUUID2newUUID[e.SourceNodeUUID]
			tNew := oldUUID2newUUID[e.TargetNodeUUID]
			if sNew.IsNil() || tNew.IsNil() {
				continue
			}
			sTplID := newUUID2tplID[sNew]
			tTplID := newUUID2tplID[tNew]
			var sHandleUUID, tHandleUUID uuid.UUID
			var sNodeName, tNodeName string

			for _, node := range newNodes {
				if node.UUID == sNew {
					sNodeName = node.Name
				}
				if node.UUID == tNew {
					tNodeName = node.Name
				}
			}

			for _, h := range handleIndex[sTplID] {
				if h.HandleKey == e.SourceHandleKey && h.IoType == e.SourceHandleIO {
					sHandleUUID = h.UUID
					break
				}
			}
			for _, h := range handleIndex[tTplID] {
				if h.HandleKey == e.TargetHandleKey && h.IoType == e.TargetHandleIO {
					tHandleUUID = h.UUID
					break
				}
			}

			if sHandleUUID.IsNil() {
				return code.ParamErr.WithMsgf("源节点 '%s' 的句柄匹配失败: handle_key='%s', io_type='%s'，目标实验室中可能不存在对应的句柄配置", sNodeName, e.SourceHandleKey, e.SourceHandleIO)
			}
			if tHandleUUID.IsNil() {
				return code.ParamErr.WithMsgf("目标节点 '%s' 的句柄匹配失败: handle_key='%s', io_type='%s'，目标实验室中可能不存在对应的句柄配置", tNodeName, e.TargetHandleKey, e.TargetHandleIO)
			}
			edgesToCreate = append(edgesToCreate, &model.WorkflowEdge{SourceNodeUUID: sNew, TargetNodeUUID: tNew, SourceHandleUUID: sHandleUUID, TargetHandleUUID: tHandleUUID})
		}
		if err := w.workflowStore.DuplicateEdge(txCtx, edgesToCreate); err != nil {
			return err
		}

		resp = &workflow.CreateResp{UUID: wk.UUID, Name: wk.Name, Description: wk.Description, Tags: []string(wk.Tags)}
		// Upsert 导入的 tags 到 tags 表，幂等
		if req.Data.Published != nil &&
			*req.Data.Published &&
			len(req.Data.Tags) > 0 {
			tagModels := make([]*model.Tags, 0, len(req.Data.Tags))
			for _, t := range req.Data.Tags {
				tagModels = append(tagModels, &model.Tags{Type: model.WorkflowTemplateTag, Name: t})
			}
			if err := w.tagsStore.UpsertTags(txCtx, tagModels); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return resp, nil
}
