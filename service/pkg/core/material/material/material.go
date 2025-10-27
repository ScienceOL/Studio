package material

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/olahol/melody"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/material"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"

	// machineImpl "github.com/scienceol/studio/service/pkg/repo/machine"
	"github.com/scienceol/studio/service/pkg/model"
	mStore "github.com/scienceol/studio/service/pkg/repo/material"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"
)

type materialImpl struct {
	envStore      repo.LaboratoryRepo
	materialStore repo.MaterialRepo
	wsClient      *melody.Melody
	msgCenter     notify.MsgCenter
	// machine       repo.Machine // deprecated
	rClient *r.Client
}

func NewMaterial(ctx context.Context, wsClient *melody.Melody) material.Service {
	m := &materialImpl{
		envStore:      eStore.New(),
		materialStore: mStore.NewMaterialImpl(),
		wsClient:      wsClient,
		msgCenter:     events.NewEvents(),
		// machine:       machineImpl.NewMachine(), // deprecated
		rClient: redis.GetClient(),
	}
	if err := events.NewEvents().Registry(ctx, notify.MaterialModify, m.OnMaterialNotify); err != nil {
		logger.Errorf(ctx, "Registry MaterialModify fail err: %+v", err)
	}

	return m
}

func (m *materialImpl) CreateMaterial(ctx context.Context, req *material.GraphNodeReq) error {
	labUser := auth.GetCurrentUser(ctx)
	if labUser == nil {
		return code.UnLogin
	}

	if len(req.Nodes) == 0 {
		return nil
	}

	labData, err := m.envStore.GetLabByAkSk(ctx, labUser.AccessKey, labUser.AccessSecret)
	if err != nil {
		return err
	}
	if err := m.createNodes(ctx, labData, req); err != nil {
		return err
	}

	_ = m.addEdges(ctx, labData.ID, req.Edges, false)
	return nil
}

func (m *materialImpl) SaveMaterial(ctx context.Context, req *material.SaveGrapReq) error {
	labID := m.materialStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
	if labID == 0 {
		return code.LabNotFound
	}

	return m.saveAllMaterial(ctx, labID, &req.Graph)
}

func (m *materialImpl) LabMaterial(ctx context.Context, req *material.MaterialReq) ([]*material.MaterialResp, error) {
	labUser := auth.GetCurrentUser(ctx)
	if labUser == nil {
		return nil, code.UnLogin
	}

	var allNodes []*model.MaterialNode
	nodeMap := make(map[int64]*model.MaterialNode)
	if req.ID == "" {
		if err := m.materialStore.FindDatas(ctx, &allNodes, map[string]any{
			"lab_id": labUser.LabID,
		}); err != nil {
			return nil, err
		}
		nodeMap = utils.Slice2Map(allNodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
			return node.ID, node
		})
	} else {
		names := strings.Split(strings.Trim(req.ID, "/"), "/")
		if len(names) == 0 {
			return nil, nil
		}
		if slices.Contains(names, "") {
			logger.Errorf(ctx, "LabMaterial id contains empty name")
			return nil, code.PathHasEmptyName.WithMsg(req.ID)
		}

		nodes, err := m.materialStore.GetMaterialNodeByPath(ctx, labUser.LabID, names)
		if err != nil {
			return nil, err
		}

		if len(nodes) != len(names) {
			return nil, code.CanNotFoundTargetNode
		}

		allNodes = append(allNodes, nodes[len(names)-1])
		if req.WithChildren {
			childrens, err := m.materialStore.GetDescendants(ctx, nodes[len(names)-1].LabID, nodes[len(names)-1].ID)
			if err != nil {
				return nil, err
			}
			allNodes = append(allNodes, childrens...)
			nodeMap = utils.Slice2Map(append(nodes, childrens...), func(node *model.MaterialNode) (int64, *model.MaterialNode) {
				return node.ID, node
			})
		} else {
			nodeMap = utils.Slice2Map(nodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
				return node.ID, node
			})
		}
	}

	return m.formatMaterialNode(ctx, nodeMap, allNodes, true)
}

// FIXME: 这是一段垃圾代码，逻辑不严谨，需要 edge 侧配合更改逻辑, 临时使用。该函数会创建或者更新节点数据
func (m *materialImpl) BatchUpdateMaterial(ctx context.Context, req *material.UpdateMaterialReq) error {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}
	// 默认所有 parent 都是 根节点名字， FIXME: 这个逻辑不可能
	parentNames := utils.FilterUniqSlice(req.Nodes, func(n *material.Node) (string, bool) {
		if n.Parent != "" {
			return n.Parent, true
		}
		return "", false
	})
	parentNodes := make([]*model.MaterialNode, 0, len(parentNames))
	if err := m.materialStore.FindDatas(ctx, &parentNodes, map[string]any{
		"name":      parentNodes,
		"parent_id": 0,
		"lab_id":    userInfo.LabID,
	}, "id", "name"); err != nil {
		return err
	}

	tplNames := utils.FilterUniqSlice(req.Nodes, func(n *material.Node) (string, bool) {
		if n.Class != "" {
			return n.Class, true
		}

		return "", false
	})

	resourceNodes := make([]*model.ResourceNodeTemplate, 0, len(tplNames))
	if err := m.materialStore.FindDatas(ctx, &resourceNodes, map[string]any{
		"name":   tplNames,
		"lab_id": userInfo.LabID,
	}, "id", "name", "config_schema"); err != nil {
		return err
	}

	resourceNameMap := utils.Slice2Map(resourceNodes, func(r *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
		return r.Name, r
	})

	parentMap := utils.Slice2Map(parentNodes, func(n *model.MaterialNode) (string, int64) {
		return n.Name, n.ID
	})

	nodes := utils.FilterSlice(req.Nodes, func(n *material.Node) (*model.MaterialNode, bool) {
		d := &model.MaterialNode{
			LabID:       userInfo.LabID,
			Name:        n.DeviceID,
			DisplayName: n.Name,
			ParentID:    parentMap[n.Parent],
			Data:        n.Data,
		}

		tpl, ok := resourceNameMap[n.Class]
		if ok && tpl != nil {
			d.ResourceNodeID = tpl.ID
			d.InitParamData = tpl.ConfigSchema
		}

		return d, true
	})

	rets, err := m.materialStore.UpsertMaterialNode(ctx, nodes, []string{"lab_id", "name", "parent_id"},
		[]string{"uuid", "data"}, "data", "resource_node_id")
	if err != nil {
		return err
	}

	// FIXME: 不要了
	data := utils.FilterSlice(rets, func(n *model.MaterialNode) (*material.WSNode, bool) {
		return &material.WSNode{
			UUID:            uuid.UUID{},
			ParentUUID:      uuid.UUID{},
			Name:            "",
			DisplayName:     "",
			Description:     new(string),
			Type:            "",
			ResTemplateUUID: uuid.UUID{},
			ResTemplateName: "",
			InitParamData:   datatypes.JSON{},
			Schema:          datatypes.JSON{},
			Data:            datatypes.JSON{},
			Status:          "",
			Header:          "",
			Pose:            datatypes.JSONType[model.Pose]{},
			Model:           datatypes.JSON{},
			Icon:            "",
			Handles:         []*material.WSHandle{},
		}, true
	})

	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		LabUUID: userInfo.LabUUID,
		UUID:    uuid.NewV4(),
		Data: material.UpdateMaterialResNotify{
			Action: string(material.UpdateNodeData),
			Data:   data,
		},
	}); err != nil {
		logger.Errorf(ctx, "BatchUpdateMaterial Broadcast msg fail err: %+v", err)
	}

	return nil
}

// 默认所有的物料名字都不一样
func (m *materialImpl) BatchUpdateUniqueName(ctx context.Context, req *material.UpdateMaterialReq) error {
	if len(req.Nodes) == 0 {
		return nil
	}

	// 包含如下几个事件， 层级结构变更，创建物料，更新数据
	labUser := auth.GetCurrentUser(ctx)
	// 默认所有 name 都是唯一的，不会出现重复
	// 检测上传重复名字
	uniqueNodes, duplicateNodes := utils.FindDuplicates(req.Nodes, func(n *material.Node) string {
		return n.Name
	})

	if len(duplicateNodes) > 0 {
		duplicateNames := utils.FilterSlice(duplicateNodes, func(n *material.Node) (string, bool) {
			return n.Name, true
		})
		logger.Warnf(ctx, "BatchUpdateUniqueName lab id: %d, name slice: %+v", labUser.LabID, duplicateNames)
	}

	if len(uniqueNodes) == 0 {
		logger.Warnf(ctx, "BatchUpdateUniqueName no update node lab id: %d", labUser.LabID)
		return nil
	}

	nodeNames := utils.FilterSlice(uniqueNodes, func(n *material.Node) (string, bool) {
		return n.Name, true
	})

	nodeDatas := make([]*model.MaterialNode, 0, len(nodeNames))
	if err := m.materialStore.FindDatas(ctx, &nodeDatas, map[string]any{
		"lab_id": labUser.LabID,
		"name":   nodeNames,
	}); err != nil {
		return err
	}

	// 检测数据库重复名字
	uniqueNodeDatas, duplicateNodeDatas := utils.FindDuplicates(nodeDatas, func(n *model.MaterialNode) string {
		return n.Name
	})

	// 数据库内有重复的 name 不更新数据
	if len(duplicateNodeDatas) > 0 {
		duplicateNodeDataMap := utils.Slice2Map(duplicateNodeDatas, func(n *model.MaterialNode) (string, bool) {
			return n.Name, true
		})
		logger.Warnf(ctx, "BatchUpdateUniqueName duplicate db node lab id: %d, name slice: %+v", labUser.LabID,
			utils.FilterSlice(duplicateNodeDatas, func(n *model.MaterialNode) (string, bool) {
				return n.Name, true
			}))

		// 数据库内有重复的数据，不更新
		updateNodes := make([]*material.Node, 0, len(uniqueNodes))
		for _, reqNode := range uniqueNodes {
			if _, ok := duplicateNodeDataMap[reqNode.Name]; ok {
				continue
			}
			updateNodes = append(updateNodes, reqNode)
		}
		uniqueNodes = updateNodes
	}

	if len(uniqueNodes) == 0 {
		return nil
	}

	if len(uniqueNodes) < len(uniqueNodeDatas) {
		// 不会存在更新的节点比数据库索引出来的还多
		logger.Warnf(ctx, "BatchUpdateUniqueName no this scene")
		return nil
	}

	uniqueNodeMap := utils.Slice2Map(uniqueNodes, func(n *material.Node) (string, *material.Node) {
		return n.Name, n
	})

	uniqueNodeDataMap := utils.Slice2Map(uniqueNodeDatas, func(n *model.MaterialNode) (string, *model.MaterialNode) {
		return n.Name, n
	})

	createNodes := make([]*material.Node, 0, len(uniqueNodes))
	updateNodes := make([]*material.UpdatePair, 0, len(uniqueNodes))
	for key, n := range uniqueNodeMap {
		if dbNode, ok := uniqueNodeDataMap[key]; !ok {
			createNodes = append(createNodes, n)
		} else {
			updateNodes = append(updateNodes, &material.UpdatePair{
				ReqNode: n,
				DBNode:  dbNode,
			})
		}
	}
	if err := m.updateMaterialNode(ctx, updateNodes); err != nil {
		logger.Errorf(ctx, "BatchUpdateUniqueName update node fail err: %+v", err)
	}
	if err := m.createMaterialNode(ctx, createNodes); err != nil {
		logger.Errorf(ctx, "BatchUpdateUniqueName create node fail err: %+v", err)
	}

	return nil
}

func (m *materialImpl) createMaterialNode(ctx context.Context, reqNodes []*material.Node) error {
	if len(reqNodes) == 0 {
		return nil
	}
	labUser := auth.GetCurrentUser(ctx)
	resourceNames := make([]string, 0, len(reqNodes))
	parentNames := utils.FilterSlice(reqNodes, func(n *material.Node) (string, bool) {
		if n.Class != "" {
			resourceNames = utils.AppendUniqSlice(resourceNames, n.Class)
		}
		return n.Parent, n.Parent != ""
	})

	parentNodes := make([]*model.MaterialNode, 0, len(parentNames))
	if err := m.materialStore.FindDatas(ctx, &parentNodes, map[string]any{
		"lab_id": labUser.LabID,
		"name":   parentNames,
	}, "name", "id", "uuid"); err != nil {
		logger.Errorf(ctx, "createMaterialNode find parent node fail lab id: %d, names: %v", labUser.LabID, parentNames)
	}

	resNodes := make([]*model.ResourceNodeTemplate, 0, len(resourceNames))
	if err := m.materialStore.FindDatas(ctx, &resNodes, map[string]any{
		"lab_id": labUser.LabID,
		"name":   resourceNames,
	}, "name", "id", "uuid", "icon"); err != nil {
		logger.Errorf(ctx, "createMaterialNode find resource node fail lab id: %d, names: %v", labUser.LabID, resourceNames)
	}
	parentNodeMap := utils.Slice2Map(parentNodes, func(n *model.MaterialNode) (string, *model.MaterialNode) {
		return n.Name, n
	})

	parentNodeIDMap := utils.Slice2Map(parentNodes, func(n *model.MaterialNode) (int64, *model.MaterialNode) {
		return n.ID, n
	})

	resNodeMap := utils.Slice2Map(resNodes, func(n *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
		return n.Name, n
	})

	datas := utils.FilterSlice(reqNodes, func(n *material.Node) (*model.MaterialNode, bool) {
		var parentID int64
		var resID int64
		var icon string
		parentNode, ok := parentNodeMap[n.Parent]
		if ok {
			parentID = parentNode.ID
		}

		resNode, ok := resNodeMap[n.Class]
		if ok {
			resID = resNode.ID
			icon = resNode.Icon
		}

		config := &material.InnerBaseConfig{}
		_ = json.Unmarshal([]byte(n.Config), config)

		pose := model.Pose{
			Position: n.Position,
			Size: model.Size{
				Width:  config.SizeX,
				Height: config.SizeY,
				Depth:  config.SizeZ,
			},
			Layout:   "2d",
			Rotation: config.Rotation,
			Scale: model.Scale{
				X: 1,
				Y: 1,
				Z: 1,
			},
		}

		node := &model.MaterialNode{
			ParentID:       parentID,
			LabID:          labUser.LabID,
			Name:           n.DeviceID,
			DisplayName:    n.Name,
			Description:    n.Description,
			Status:         "idle",
			Type:           n.Type,
			ResourceNodeID: resID,
			Class:          n.Class,
			InitParamData:  n.Config,
			Schema:         n.Schema,
			Data:           n.Data,
			Pose:           datatypes.NewJSONType(pose),
			Model:          n.Model,
			Icon:           icon,
		}
		return node, true
	})

	if _, err := m.materialStore.UpsertMaterialNode(ctx, datas, nil, nil); err != nil {
		return err
	}

	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		LabUUID: labUser.LabUUID,
		UUID:    uuid.NewV4(),
		Data: material.UpdateMaterialResNotify{
			Action: string(material.UpdateNodeCreate),
			Data: utils.FilterSlice(datas, func(n *model.MaterialNode) (*material.WSNode, bool) {
				var resUUID uuid.UUID
				var resName string
				var parentUUID uuid.UUID
				resNode, ok := resNodeMap[n.Class]
				if ok {
					resName = resNode.Name
					resUUID = resNode.UUID
				}

				parentNode, ok := parentNodeIDMap[n.ParentID]
				if ok {
					parentUUID = parentNode.UUID
				}

				return &material.WSNode{
					UUID:            n.UUID,
					ParentUUID:      parentUUID,
					Name:            n.Name,
					DisplayName:     n.DisplayName,
					Description:     n.Description,
					Type:            n.Type,
					ResTemplateUUID: resUUID,
					ResTemplateName: resName,
					InitParamData:   n.InitParamData,
					Schema:          n.Schema,
					Data:            n.Data,
					Status:          n.Status,
					Header:          utils.Or(n.DisplayName, n.Name),
					Pose:            n.Pose,
					Model:           n.Model,
					Icon:            n.Icon,
					// Handles:         []*material.WSHandle{},
				}, true
			}),
		},
		// UserID:       labUser.ID,
		Timestamp:    time.Now().Unix(),
		WorkflowUUID: uuid.UUID{},
		TaskUUID:     uuid.UUID{},
	}); err != nil {
		logger.Errorf(ctx, "BatchUpdateMaterial Broadcast msg fail err: %+v", err)
	}

	return nil
}

func (m *materialImpl) updateMaterialNode(ctx context.Context, reqNodes []*material.UpdatePair) error {
	if len(reqNodes) == 0 {
		return nil
	}

	labUser := auth.GetCurrentUser(ctx)
	parentNames := utils.FilterSlice(reqNodes, func(n *material.UpdatePair) (string, bool) {
		return n.ReqNode.Parent, n.ReqNode.Parent != ""
	})

	parentNodes := make([]*model.MaterialNode, 0, len(parentNames))
	if err := m.materialStore.FindDatas(ctx, &parentNodes, map[string]any{
		"lab_id": labUser.LabID,
		"name":   parentNames,
	}, "name", "id", "uuid"); err != nil {
		logger.Errorf(ctx, "createMaterialNode find parent node fail lab id: %d, names: %v", labUser.LabID, parentNames)
	}

	parentNodeMap := utils.Slice2Map(parentNodes, func(n *model.MaterialNode) (string, *model.MaterialNode) {
		return n.Name, n
	})

	parentNodeIDMap := utils.Slice2Map(parentNodes, func(n *model.MaterialNode) (int64, *model.MaterialNode) {
		return n.ID, n
	})

	// 更新 postion、data 、 parent
	datas := utils.FilterSlice(reqNodes, func(n *material.UpdatePair) (*model.MaterialNode, bool) {
		var parentID int64
		parentNode, ok := parentNodeMap[n.ReqNode.Parent]
		if ok {
			parentID = parentNode.ID
		}

		config := &material.InnerBaseConfig{}
		_ = json.Unmarshal([]byte(n.ReqNode.Config), config)

		pose := model.Pose{
			Position: n.ReqNode.Position,
			Size: model.Size{
				Width:  config.SizeX,
				Height: config.SizeY,
				Depth:  config.SizeZ,
			},
			Layout:   "2d",
			Rotation: config.Rotation,
			Scale: model.Scale{
				X: 1,
				Y: 1,
				Z: 1,
			},
		}

		data := &model.MaterialNode{
			ParentID:      parentID,
			InitParamData: n.ReqNode.Config,
			Data:          n.ReqNode.Data,
			Pose:          datatypes.NewJSONType(pose),
		}

		data.ID = n.DBNode.ID
		data.UUID = n.DBNode.UUID
		return data, true
	})

	if _, err := m.materialStore.UpsertMaterialNode(ctx, datas,
		[]string{"id"},
		nil,
		[]string{"parent_id", "init_param_data", "data", "pose"}...); err != nil {
		return err
	}

	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		LabUUID: labUser.LabUUID,
		UUID:    uuid.NewV4(),
		Data: material.UpdateMaterialResNotify{
			Action: string(material.UpdateNodeResData),
			Data: utils.FilterSlice(datas, func(n *model.MaterialNode) (*material.WSNode, bool) {
				var parentUUID uuid.UUID
				parentNode, ok := parentNodeIDMap[n.ParentID]
				if ok {
					parentUUID = parentNode.UUID
				}

				return &material.WSNode{
					UUID:          n.UUID,
					ParentUUID:    parentUUID,
					InitParamData: n.InitParamData,
					Data:          n.Data,
					Pose:          n.Pose,
				}, true
			}),
		},
		// UserID:    labUser.ID,
		Timestamp: time.Now().Unix(),
	}); err != nil {
		logger.Errorf(ctx, "BatchUpdateMaterial Broadcast msg fail err: %+v", err)
	}

	return nil
}

func (m *materialImpl) formatMaterialNode(ctx context.Context,
	nodeMap map[int64]*model.MaterialNode,
	nodes []*model.MaterialNode,
	edgeFormat bool,
) ([]*material.MaterialResp, error) {
	resIDs := utils.FilterUniqSlice(nodes, func(node *model.MaterialNode) (int64, bool) {
		return node.ResourceNodeID, node.ResourceNodeID > 0
	})

	resHandleMap, err := m.envStore.GetResourceHandleTemplates(ctx, resIDs)
	if err != nil {
		return nil, err
	}

	resps := utils.FilterSlice(nodes, func(node *model.MaterialNode) (*material.MaterialResp, bool) {
		params := make([]*material.DataParam, 0, 2)
		if len(node.Data) != 0 {
			params = append(params, &material.DataParam{
				ParamDataKey:   "data",
				ParamType:      "DEFAULT",
				Title:          "Data",
				ParamInputData: node.Data,
				SelectChoices:  "",
				Attachment:     "",
			})
		}

		if len(node.InitParamData) != 0 {
			params = append(params, &material.DataParam{
				ParamDataKey:   "config",
				ParamType:      "DEFAULT",
				Title:          "Configuration",
				ParamInputData: node.InitParamData,
				SelectChoices:  "",
				Attachment:     "",
			})
		}

		resp := &material.MaterialResp{
			ID:        node.Name,
			CloudUUID: node.UUID,
			Type:      utils.Or(string(node.Type), string(model.MATERIALCONTAINER)),
			Data: &material.MaterialData{
				Header: utils.Or(node.DisplayName, node.Name),
				Handles: utils.FilterSlice(resHandleMap[node.ResourceNodeID], func(h *model.ResourceHandleTemplate) (*material.HandleData, bool) {
					return &material.HandleData{
						ID:           h.Name,
						Type:         h.IOType,
						HasConnected: true,
						Required:     true,
					}, true
				}),
				Params:       params,
				Executors:    []any{},
				Footer:       "",
				NodeCardIcon: node.Icon,
			},
			Position:      node.Pose.Data().Position,
			Status:        node.Status,
			Minimized:     false,
			Disabled:      false,
			Version:       "1.0.0",
			DragHandle:    ".drag-handle", // FIXME: 这个字段作用是什么
			DeviceID:      node.Name,
			Name:          node.DisplayName,
			ExperimentEnv: node.LabID,
			Description:   node.Description,
			Collapsed:     false,
			Width:         float32(node.Pose.Data().Size.Width),
			Height:        float32(node.Pose.Data().Size.Height),
			ChildNodesUUID: utils.FilterSlice(nodes, func(n *model.MaterialNode) (string, bool) {
				return n.Name, n.ParentID == node.ID
			}),
			ParentNodeUUID: func() string {
				pNode, ok := nodeMap[node.ParentID]
				if ok {
					return pNode.Name
				}
				return ""
			}(),
			EqType: node.Class,
			Dirs:   m.getDirs(ctx, node, nodeMap, 10000),
		}

		if edgeFormat {
			resp.Data = utils.Ternary(len(node.Data) != 0, node.Data, datatypes.JSON{})
			resp.Config = node.InitParamData
			resp.Parent = resp.ParentNodeUUID
			resp.Children = resp.ChildNodesUUID
			resp.Class = node.Class
		}
		return resp, true
	})

	return resps, nil
}

func (m *materialImpl) getDirs(ctx context.Context, node *model.MaterialNode, nodes map[int64]*model.MaterialNode, maxDeep int) []string {
	current := node
	res := make([]string, 0, 10)
	visited := make(map[int64]bool) // 防止循环引用
	for current != nil && maxDeep > 0 {
		res = append(res, current.Name)
		if _, ok := visited[current.ID]; ok {
			logger.Errorf(ctx, "getDirs has circular id: %d", current.ID)
			return []string{}
		}
		visited[current.ID] = true
		current, _ = nodes[current.ParentID]
		maxDeep--
	}

	slices.Reverse(res)
	return res
}

func (m *materialImpl) RecalculatePosition(_ context.Context, req *material.GraphNodeReq) {
	index := 0
	_ = utils.FilterSlice(req.Nodes, func(n *material.Node) (*material.Node, bool) {
		nodeType := string(n.Type)
		innerConfig := &material.InnerBaseConfig{}
		_ = json.Unmarshal(n.Config, innerConfig)
		if nodeType == string(model.MATERIALDEVICE) &&
			string(nodeType) != innerConfig.Type &&
			innerConfig.Type != "" {
			nodeType = innerConfig.Type
		}
		nodeType = strings.ToLower(string(nodeType))
		n.Type = model.DEVICETYPE(nodeType)

		pose := n.Pose.Data()
		pose.Layout = utils.Or(pose.Layout, "2d")
		pose.Position = model.Position{
			X: utils.Or(n.Position.X, 0),
			Y: utils.Or(n.Position.Y, 0),
			Z: utils.Or(n.Position.Z, 0),
		}
		pose.Size = model.Size{
			Width:  innerConfig.SizeX,
			Height: innerConfig.SizeY,
			Depth:  innerConfig.SizeZ,
		}
		pose.Scale = model.Scale{
			X: utils.Or(pose.Scale.X, 1),
			Y: utils.Or(pose.Scale.Y, 1),
			Z: utils.Or(pose.Scale.Z, 1),
		}
		pose.Rotation = model.Rotation{
			X: utils.Or(innerConfig.Rotation.X, 0),
			Y: utils.Or(innerConfig.Rotation.Y, 0),
			Z: utils.Or(innerConfig.Rotation.Z, 0),
		}

		if n.Position3D == nil {
			pose.Postion3D = model.Position{
				X: n.Position.X,
				Y: n.Position.Y,
				Z: n.Position.Z,
			}
		} else {
			pose.Postion3D = model.Position{
				X: n.Position3D.X,
				Y: n.Position3D.Y,
				Z: n.Position3D.Z,
			}
		}

		pose.CrossSectionType = utils.Or(n.Pose.Data().CrossSectionType, "rectangle")

		// pose.Position.X = float32((pose.Size.Width + 20) * (index % 10))
		// pose.Position.Y = float32(2 * pose.Size.Height * (index / 10))

		n.Pose = datatypes.NewJSONType(pose)
		index++
		return nil, false
	})
}

func (m *materialImpl) createNodes(ctx context.Context, labData *model.Laboratory, req *material.GraphNodeReq) error {
	resTplNames := make([]string, 0, len(req.Nodes))
	for _, data := range req.Nodes {
		if data.Class == "" {
			continue
		}
		resTplNames = utils.AppendUniqSlice(resTplNames, data.Class)
	}

	resourceNodes := make([]*model.ResourceNodeTemplate, 0, len(resTplNames))
	err := m.envStore.FindDatas(ctx, &resourceNodes, map[string]any{
		"lab_id": labData.ID,
		"name":   resTplNames,
	}, "id", "name", "icon", "model")
	if err != nil {
		return err
	}

	// 强制校验资源模板是否存在
	if len(resourceNodes) != len(resTplNames) {
		return code.ResNotExistErr
	}

	resMap := utils.Slice2Map(resourceNodes, func(item *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
		return item.Name, item
	})

	nodeNames := utils.FilterSlice(req.Nodes, func(item *material.Node) (*utils.Node[string, *material.Node], bool) {
		if item.Name == "" {
			return nil, false
		}

		return &utils.Node[string, *material.Node]{
			Name:   item.DeviceID,
			Parent: item.Parent,
			Data:   item,
		}, true
	})

	m.RecalculatePosition(ctx, req)
	levelNodes, err := utils.BuildHierarchy(nodeNames)
	if err != nil {
		return code.InvalidDagErr.WithMsg(err.Error())
	}

	nodeMap := make(map[string]*model.MaterialNode)
	if err = db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range levelNodes {
			datas := make([]*model.MaterialNode, 0, len(nodes))
			for _, n := range nodes {
				data := &model.MaterialNode{
					ParentID:       0,
					LabID:          labData.ID,
					Name:           n.DeviceID,
					DisplayName:    n.Name,
					Description:    n.Description,
					Class:          n.Class,
					Type:           n.Type,
					ResourceNodeID: 0,
					InitParamData:  n.Config,
					Data:           n.Data,
					Pose:           n.Pose,
					Icon:           "",
					Schema:         n.Schema,
				}
				if node := nodeMap[n.Parent]; node != nil {
					data.ParentID = node.ID
				}

				if resInfo := resMap[n.Class]; resInfo != nil {
					data.ResourceNodeID = resInfo.ID
					data.Icon = resInfo.Icon
					data.Model = resInfo.Model
				}

				datas = append(datas, data)
				nodeMap[data.Name] = data
			}

			if _, err := m.materialStore.UpsertMaterialNode(txCtx, datas, nil, nil); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *materialImpl) CreateEdge(ctx context.Context, req *material.GraphEdge) error {
	labUser := auth.GetCurrentUser(ctx)
	if labUser == nil {
		return code.UnLogin
	}
	labData, err := m.envStore.GetLabByAkSk(ctx, labUser.AccessKey, labUser.AccessSecret)
	if err != nil {
		return err
	}

	return m.addEdges(ctx, labData.ID, req.Edges, true)
}

func (m *materialImpl) addEdges(ctx context.Context, labID int64, edges []*material.Edge, checkLink bool) error {
	nodeNames := make([]string, 0, 2*len(edges))
	handleNames := make([]string, 0, 2*len(edges))

	for _, edge := range edges {
		nodeNames = utils.AppendUniqSlice(nodeNames, edge.Source)
		nodeNames = utils.AppendUniqSlice(nodeNames, edge.Target)

		handleNames = utils.AppendUniqSlice(handleNames, edge.SourceHandle)
		handleNames = utils.AppendUniqSlice(handleNames, edge.TargetHandle)
	}

	edgeInfo, err := m.materialStore.GetNodeHandles(ctx, labID, nodeNames, handleNames)
	if err != nil {
		return err
	}
	edgeDatas := make([]*model.MaterialEdge, 0, len(edges))
	for _, edge := range edges {
		sourceNode, ok := edgeInfo[edge.Source]
		if !ok && checkLink {
			logger.Errorf(ctx, "CreateEdge source not exist lab id: %d, source node name: %s", labID, edge.Source)
			return code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("lab id: %d, source node name: %s", labID, edge.Source))
		} else if !ok {
			logger.Infof(ctx, "CreateEdge source not exist lab id: %d, source node name: %s", labID, edge.Source)
			continue
		}
		sourceHandle, ok := sourceNode[edge.SourceHandle]
		if !ok && checkLink {
			logger.Errorf(ctx, "CreateEdge source handle not exist lab id: %d, source node name: %s, handle name: %s",
				labID, edge.Source, edge.SourceHandle)
			return code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("lab id: %d, source node name: %s, handle name: %s",
				labID, edge.Source, edge.SourceHandle))
		} else if !ok {
			logger.Infof(ctx, "CreateEdge source handle not exist lab id: %d, source node name: %s, handle name: %s",
				labID, edge.Source, edge.SourceHandle)
			continue
		}

		targetNode, ok := edgeInfo[edge.Target]
		if !ok && checkLink {
			logger.Errorf(ctx, "CreateEdge target not exist lab id: %d, target node name: %s", labID, edge.Target)
			return code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("lab id: %d, target node name: %s", labID, edge.Target))
		} else if !ok {
			logger.Infof(ctx, "CreateEdge target not exist lab id: %d, target node name: %s", labID, edge.Target)
			continue
		}

		targetHandle, ok := targetNode[edge.TargetHandle]
		if !ok && checkLink {
			logger.Errorf(ctx, "CreateEdge target handle not exist lab id: %d, node name: %s", labID, edge.Target)
			return code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("lab id: %d, target node name: %s, handle name: %s",
				labID, edge.Target, edge.TargetHandle))
		} else if !ok {
			logger.Infof(ctx, "CreateEdge target handle not exist lab id: %d, node name: %s", labID, edge.Target)
			continue
		}

		edgeDatas = append(edgeDatas, &model.MaterialEdge{
			SourceNodeUUID:   sourceHandle.NodeUUID,
			SourceHandleUUID: sourceHandle.HandleUUID,
			TargetNodeUUID:   targetHandle.NodeUUID,
			TargetHandleUUID: targetHandle.HandleUUID,
		})
	}

	if err := m.materialStore.UpsertMaterialEdge(ctx, edgeDatas); err != nil {
		return err
	}

	return nil
}

func (m *materialImpl) OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error {
	msgType := &common.WsMsgType{}
	err := json.Unmarshal(b, msgType)
	if err != nil {
		return err
	}

	var data any
	var msgErr error

	switch material.ActionType(msgType.Action) {
	case material.FetchGraph: // 首次获取组态图
		data, msgErr = m.fetchGraph(ctx, s)
	case material.TemplateDetail:
		data, msgErr = m.templateDetail(ctx, s, b)
	case material.FetchTemplate: // 首次获取模板
		data, msgErr = m.fetchDeviceTemplate(ctx, s)
	case material.SaveGraph:
		data, msgErr = nil, nil
	case material.CreateNode:
		data, msgErr = m.createNode(ctx, s, b)
	case material.UpdateNode: // 批量更新节点
		msgErr = m.upateNode(ctx, s, b)
	case material.BatchDelNode: // 批量删除节点
		data, msgErr = m.batchDelNode(ctx, s, b)
	case material.BatchCreateEdge: // 批量创建边
		data, msgErr = m.batchCreateEdge(ctx, s, b)
	case material.BatchDelEdge: // 批量删除边
		data, msgErr = m.batchDelEdge(ctx, s, b)
	default:
		return common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, code.UnknownWSActionErr)
	}

	if msgErr != nil {
		return common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, msgErr)
	}

	if data != nil {
		return common.ReplyWSOk(s, msgType.Action, msgType.MsgUUID, data)
	}

	return common.ReplyWSOk(s, msgType.Action, msgType.MsgUUID)
}

// 获取组态图
func (m *materialImpl) fetchGraph(ctx context.Context, s *melody.Session) (any, error) {
	// 获取所有组态图信息
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return nil, err
	}
	nodes, err := m.materialStore.GetNodesByLabID(ctx, labData.ID)
	if err != nil {
		return nil, err
	}

	resTplIDs := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (int64, bool) {
		if nodeItem.ResourceNodeID == 0 {
			return 0, false
		}
		return nodeItem.ResourceNodeID, true
	})

	resTplIDs = utils.RemoveDuplicates(resTplIDs)
	nodesMap := utils.Slice2Map(nodes, func(item *model.MaterialNode) (int64, *model.MaterialNode) {
		return item.ID, item
	})

	resHandlesMap, err := m.envStore.GetResourceHandleTemplates(ctx, resTplIDs)
	if err != nil {
		return nil, err
	}

	nodeUUIDs := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (uuid.UUID, bool) {
		return nodeItem.UUID, true
	})

	edges, err := m.materialStore.GetEdgesByNodeUUID(ctx, nodeUUIDs)
	if err != nil {
		return nil, err
	}
	resNodes, err := m.envStore.GetResourceNodeTemplates(ctx, resTplIDs)
	if err != nil {
		return nil, err
	}

	resNodeTplMap := utils.Slice2Map(resNodes, func(item *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return item.ID, item
	})

	respNodes := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (*material.WSNode, bool) {
		var parentUUID uuid.UUID
		parentNode, ok := nodesMap[nodeItem.ParentID]
		if ok {
			parentUUID = parentNode.UUID
		}

		var tplUUID uuid.UUID
		var tplName string
		if resNodeTpl, ok := resNodeTplMap[nodeItem.ResourceNodeID]; ok {
			tplUUID = resNodeTpl.UUID
			tplName = resNodeTpl.Name
		}
		handles := resHandlesMap[nodeItem.ResourceNodeID]
		return &material.WSNode{
			UUID:            nodeItem.UUID,
			ParentUUID:      parentUUID,
			Name:            nodeItem.Name,
			DisplayName:     nodeItem.DisplayName,
			Description:     nodeItem.Description,
			Type:            nodeItem.Type,
			ResTemplateUUID: tplUUID,
			ResTemplateName: tplName,
			InitParamData:   nodeItem.InitParamData,
			Schema:          nodeItem.Schema,
			Data:            nodeItem.Data,
			Status:          nodeItem.Status,
			Pose:            nodeItem.Pose,
			Model:           nodeItem.Model,
			Icon:            nodeItem.Icon,
			Handles: utils.FilterSlice(handles, func(handleItem *model.ResourceHandleTemplate) (*material.WSHandle, bool) {
				return &material.WSHandle{
					UUID:        handleItem.UUID,
					Name:        handleItem.Name,
					Side:        handleItem.Side,
					DisplayName: handleItem.DisplayName,
					Type:        handleItem.Type,
					IOType:      handleItem.IOType,
					Source:      handleItem.Source,
					Key:         handleItem.Key,
				}, true
			}),
		}, true
	})

	respEdges := utils.FilterSlice(edges, func(item *model.MaterialEdge) (*material.WSEdge, bool) {
		return &material.WSEdge{
			UUID:             item.UUID,
			SourceNodeUUID:   item.SourceNodeUUID,
			TargetNodeUUID:   item.TargetNodeUUID,
			SourceHandleUUID: item.SourceHandleUUID,
			TargetHandleUUID: item.TargetHandleUUID,
			Type:             "step",
		}, true
	})

	resp := &material.WSGraph{
		Nodes: respNodes,
		Edges: respEdges,
	}

	return resp, nil
}

func (m *materialImpl) getLab(ctx context.Context, s *melody.Session) (*model.Laboratory, error) {
	sessionValue, ok := s.Get("lab_uuid")
	if !ok {
		return nil, code.CanNotGetLabIDErr
	}
	labUUID, _ := sessionValue.(uuid.UUID)
	if labUUID.IsNil() {
		return nil, code.CanNotGetLabIDErr
	}

	labData, err := m.envStore.GetLabByUUID(ctx, labUUID)
	if err != nil {
		return nil, err
	}
	return labData, nil
}

// func (m *materialImpl) buildTplNode(ctx context.Context, nodes []*model.ResourceNodeTemplate) []*model.ResourceNodeTemplate {
// 	nodeMap := utils.Slice2Map(nodes, func(node *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
// 		return node.ID, node
// 	})
//
// 	rootNodes := make([]*model.ResourceNodeTemplate, 0, len(nodes))
//
// 	for _, n := range nodes {
// 		if n.ParentID != 0 {
// 			node, ok := nodeMap[n.ParentID]
// 			if ok {
// 				node.ConfigInfo = append(node.ConfigInfo, n)
// 			} else {
// 				logger.Errorf(ctx, "buildTplNode can not found parent node id: %d, parent id: %d", n.ID, n.ParentID)
// 			}
// 		} else {
// 			rootNodes = append(rootNodes, n)
// 		}
// 	}
// 	return rootNodes
// }

// func (m *materialImpl) getChildren(ctx context.Context, node *model.ResourceNodeTemplate, maxDeep int) ([]*model.ResourceNodeTemplate, error) {
// 	if maxDeep <= 0 {
// 		logger.Errorf(ctx, "getChildren reach max deep")
// 		return nil, code.MaxTplNodeDeepErr
// 	}
//
// 	children := make([]*model.ResourceNodeTemplate, 0, len(node.ConfigInfo))
// 	for _, child := range node.ConfigInfo {
// 		if child == nil {
// 			continue
// 		}
//
// 		if len(child.ConfigInfo) > 0 {
// 			deepChildren, err := m.getChildren(ctx, child, maxDeep-1)
// 			if err != nil {
// 				return nil, err
// 			}
// 			children = append(children, deepChildren...)
// 		}
//
// 		children = append(children, child)
// 	}
// 	return children, nil
// }

// 获取指定模板详情
func (m *materialImpl) templateDetail(_ context.Context, _ *melody.Session, b []byte) (any, error) {
	req := &common.WSData[uuid.UUID]{}
	err := json.Unmarshal(b, req)
	if err != nil {
		return nil, code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	data := req.Data
	if data.UUID.IsNil() {
		return nil, code.ParamErr.WithMsg("update node uuid is empyt")
	}

	// tplDatas := make([]*model.ResourceNodeTemplatejh)
	// m.materialStore.FindDatas(ctx, )

	return nil, nil
}

// 获取设备模板
func (m *materialImpl) fetchDeviceTemplate(ctx context.Context, s *melody.Session) (any, error) {
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return nil, err
	}

	tplNodes, err := m.envStore.GetAllResourceTemplateByLabID(ctx, labData.ID, "uuid", "name", "resource_type", "tags")
	if err != nil {
		return nil, err
	}
	// tplIDs := utils.FilterSlice(tplNodes, func(item *model.ResourceNodeTemplate) (int64, bool) {
	// 	return item.ID, true
	// })

	// tplHandles, err := m.envStore.GetResourceHandleTemplates(ctx, tplIDs)
	// if err != nil {
	// 	return nil, err
	// }

	// tplNodeMap := utils.Slice2Map(tplNodes, func(item *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
	// 	return item.ID, item
	// })

	// rootNode := m.buildTplNode(ctx, tplNodes)

	tplDatas, err := utils.FilterSliceWithErr(tplNodes, func(nodeItem *model.ResourceNodeTemplate) ([]*material.ResourceTemplate, bool, error) {
		// childrenNodes, err := m.getChildren(ctx, nodeItem, 5)
		// if err != nil {
		// 	return nil, false, err
		// }

		return []*material.ResourceTemplate{{
			UUID:         nodeItem.UUID,
			Name:         nodeItem.Name,
			Tags:         nodeItem.Tags,
			ResourceType: nodeItem.ResourceType,
		}}, true, nil
	})
	if err != nil {
		return nil, err
	}

	resData := &material.ResourceTemplates{
		Templates: tplDatas,
	}

	return resData, nil
}

func (m *materialImpl) saveAllMaterial(ctx context.Context, labID int64, datas *material.WSGraph) error {
	nodeUUIDs := make([]uuid.UUID, 0, len(datas.Nodes))
	tplUUIDs := make([]uuid.UUID, 0, len(datas.Nodes))
	for _, n := range datas.Nodes {
		if n.UUID.IsNil() {
			return code.ParamErr.WithMsg("saveGraph check node uuid id empty")
		}

		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, n.UUID)
		if !n.ResTemplateUUID.IsNil() {
			tplUUIDs = utils.AppendUniqSlice(tplUUIDs, n.ResTemplateUUID)
		}
	}

	mUUID2IDMap := m.materialStore.UUID2ID(ctx, &model.MaterialNode{}, nodeUUIDs...)
	resUUID2IDMap := m.materialStore.UUID2ID(ctx, &model.ResourceNodeTemplate{}, tplUUIDs...)

	nodes, err := utils.FilterSliceWithErr(datas.Nodes, func(item *material.WSNode) ([]*model.MaterialNode, bool, error) {
		if item.UUID.IsNil() || item.Name == "" {
			return nil, false, code.ParamErr.WithMsg("saveGraph node uuid is empty")
		}
		data := &model.MaterialNode{
			ParentID:       mUUID2IDMap[item.ParentUUID],
			LabID:          labID,
			Name:           item.Name,
			DisplayName:    item.DisplayName,
			Description:    item.Description,
			Type:           item.Type,
			ResourceNodeID: resUUID2IDMap[item.ResTemplateUUID],
			InitParamData:  item.InitParamData,
			Data:           item.Data,
			Pose:           item.Pose,
			Model:          item.Model,
			Icon:           utils.GetFilenameFromURL(item.Icon),
			Schema:         item.Schema,
			// Class:                  item.Class,
		}
		data.UUID = item.UUID
		return []*model.MaterialNode{data}, true, nil
	})
	if err != nil {
		return err
	}

	keys := []string{
		"parent_id",
		"name",
		"display_name",
		"description",
		"init_param_data",
		"schema",
		"data",
		"pose",
		"model",
		"icon",
		"updated_at",
	}

	if _, err := m.materialStore.UpsertMaterialNode(ctx, nodes, []string{"uuid"}, nil, keys...); err != nil {
		return err
	}
	return nil
}

// 批量创建节点
func (m *materialImpl) createNode(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	req := &common.WSData[*material.WSNode]{}
	err := json.Unmarshal(b, req)
	if err != nil {
		return nil, code.UnmarshalWSDataErr.WithMsg(err.Error())
	}
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return nil, err
	}
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	labUUIDI, _ := s.Get("lab_uuid")
	labUUID, _ := labUUIDI.(uuid.UUID)
	if labUUID.IsNil() {
		return nil, code.UnLogin
	}

	reqData := req.Data
	mData := &model.MaterialNode{}

	if reqData.ResTemplateUUID.IsNil() {
		return nil, code.TemplateNodeNotFoundErr
	}

	resNodeTpl, childTpl, resNodeChildrenTpl := m.getResourceTemplates(ctx, reqData.ResTemplateUUID)
	if resNodeTpl == nil {
		return nil, code.TemplateNodeNotFoundErr
	}
	mData.ResourceNodeID = resNodeTpl.ID
	mData.Icon = resNodeTpl.Icon
	mData.Class = resNodeTpl.Name

	if !reqData.ParentUUID.IsNil() {
		nodeID, ok := m.materialStore.UUID2ID(ctx,
			&model.MaterialNode{},
			reqData.ParentUUID)[reqData.ParentUUID]
		if !ok {
			return nil, code.CanNotFoundMaterialNodeErr
		}

		mData.ParentID = nodeID
	}

	mData.LabID = labData.ID
	mData.Name = reqData.Name
	mData.DisplayName = reqData.DisplayName
	mData.Description = utils.Or(reqData.Description, resNodeTpl.Description)
	mData.Type = utils.Or(func() model.DEVICETYPE {
		if childTpl != nil {
			return model.DEVICETYPE(childTpl.ResourceType)
		}

		return model.DEVICETYPE(resNodeTpl.ResourceType)
	}(), reqData.Type) // 只有设备类型
	mData.InitParamData = func() datatypes.JSON {
		if childTpl != nil && len(childTpl.ConfigSchema) != 0 {
			return childTpl.ConfigSchema
		}
		return reqData.InitParamData
	}()
	mData.Schema = reqData.Schema
	mData.Data = func() datatypes.JSON {
		if childTpl != nil && len(childTpl.DataSchema) != 0 {
			return childTpl.DataSchema
		}
		return reqData.Data
	}()
	mData.Pose = utils.Or(func() datatypes.JSONType[model.Pose] {
		if childTpl != nil {
			return childTpl.Pose
		}
		return reqData.Pose
	}(), datatypes.JSONType[model.Pose]{})
	mData.Model = utils.Ternary(len(reqData.Model) != 0, reqData.Model, resNodeTpl.Model)

	if _, err := m.materialStore.UpsertMaterialNode(ctx, []*model.MaterialNode{mData}, nil, nil); err != nil {
		return nil, err
	}
	reqData.UUID = mData.UUID

	childrenMaterialNode := utils.FilterSlice(resNodeChildrenTpl, func(tpl *model.ResourceNodeTemplate) (*model.MaterialNode, bool) {
		data := &model.MaterialNode{
			ResourceNodeID: tpl.ID,
			Icon:           tpl.Icon,
			ParentID:       mData.ID,
			LabID:          labData.ID,
			// Class:          tpl.Name,
			Name:        tpl.Name,
			DisplayName: tpl.Name,
			Type:        model.DEVICETYPE(tpl.ResourceType),
			// InitParamData:        tpl.DataSchema,
			// Schema:               tpl.ConfigSchema,
			InitParamData: tpl.ConfigSchema,
			Schema:        datatypes.JSON{},
			Data: func() datatypes.JSON {
				if len(reqData.PlateWellDatas[tpl.Name]) > 0 {
					return reqData.PlateWellDatas[tpl.Name]
				}

				return tpl.DataSchema
			}(),
			Pose:                 tpl.Pose,
			ResourceNodeTemplate: tpl,
		}
		return data, true
	})

	if _, err := m.materialStore.UpsertMaterialNode(ctx, childrenMaterialNode, nil, nil); err != nil {
		return nil, err
	}
	// 获取 handle
	handles := make([]*model.ResourceHandleTemplate, 0, 1)
	if err := m.materialStore.FindDatas(ctx, &handles, map[string]any{
		"resource_node_id": resNodeTpl.ID,
	}); err != nil {
		return nil, err
	}

	resDatas := make([]*material.WSNode, 0, 1+len(childrenMaterialNode))
	resDatas = append(resDatas, &material.WSNode{
		UUID:            mData.UUID,
		Name:            mData.Name,
		DisplayName:     mData.DisplayName,
		Description:     mData.Description,
		Type:            mData.Type,
		ResTemplateUUID: reqData.ResTemplateUUID,
		ResTemplateName: resNodeTpl.Name,
		InitParamData:   mData.InitParamData,
		Schema:          mData.Schema,
		Data:            mData.Data,
		Status:          mData.Status,
		Header:          mData.DisplayName,
		Pose:            mData.Pose,
		Model:           mData.Model,
		Icon:            mData.Icon,
		Handles: utils.FilterSlice(handles, func(h *model.ResourceHandleTemplate) (*material.WSHandle, bool) {
			return &material.WSHandle{
				UUID:        h.UUID,
				Name:        h.Name,
				Side:        h.Side,
				DisplayName: h.DisplayName,
				Type:        h.Type,
				IOType:      h.IOType,
				Source:      h.Source,
				Key:         h.Key,
			}, true
		}),
		ParentUUID: uuid.UUID{},
	})
	childrenDatas := utils.FilterSlice(childrenMaterialNode, func(node *model.MaterialNode) (*material.WSNode, bool) {
		return &material.WSNode{
			UUID:            node.UUID,
			ParentUUID:      reqData.UUID,
			Name:            node.Name,
			DisplayName:     node.DisplayName,
			Description:     node.Description,
			Type:            node.Type,
			ResTemplateUUID: node.ResourceNodeTemplate.UUID,
			ResTemplateName: node.ResourceNodeTemplate.Name,
			InitParamData:   node.InitParamData,
			Schema:          node.Schema,
			Data:            node.Data,
			Status:          node.Status,
			Header:          node.DisplayName,
			Pose:            node.Pose,
			Model:           node.Model,
			Icon:            node.Icon,
		}, true
	})

	resDatas = append(resDatas, childrenDatas...)

	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labUUID,
		UUID:    uuid.NewV4(),
		Data:    req,
	}); err != nil {
		logger.Errorf(ctx, "updateNode notify fail err: %+v", err)
	}

	// FIXME: 后续再优化
	m.notifyEdgeAdd(ctx, labUUID, resDatas)

	return resDatas, nil
}

// 暂时根据 uuid 查询，后续优化
func (m *materialImpl) notifyEdgeAdd(ctx context.Context, labUUID uuid.UUID, nodes []*material.WSNode) {
	if len(nodes) == 0 {
		return
	}
	ancestorNodes, err := m.materialStore.GetAncestors(ctx, nodes[0].UUID)
	if err != nil {
		return
	}

	slices.Reverse(ancestorNodes)
	deviceID := ""
	deviceUUID := uuid.UUID{}
	for _, n := range ancestorNodes {
		if n.Type == model.MATERIALDEVICE {
			deviceID = n.Name
			deviceUUID = n.UUID
			break
		}
	}

	addMaterials := utils.FilterSlice(nodes, func(n *material.WSNode) (*engine.MaterialUpdate, bool) {
		return &engine.MaterialUpdate{
			UUID:       n.UUID,
			DeviceUUID: deviceUUID,
			DeviceID:   deviceID,
		}, true
	})

	// 通知调度器
	m.notify(ctx, labUUID, engine.AddMaterial, addMaterials)
}

func (m *materialImpl) notify(ctx context.Context, labUUID uuid.UUID, action engine.WorkflowAction, uuids []*engine.MaterialUpdate) {
	if len(uuids) == 0 {
		return
	}

	data := engine.WorkflowInfo{
		Action:  action,
		LabUUID: labUUID,
		Data:    uuids,
	}

	dataB, _ := json.Marshal(data)
	conf := config.Global().Job
	ret := m.rClient.LPush(ctx, conf.JobQueueName, dataB)
	if ret.Err() != nil {
		logger.Errorf(ctx, "notify material ============ send data error: %+v", ret.Err())
	}
}

func (m *materialImpl) getResourceTemplates(ctx context.Context,
	resourceNodeUUID uuid.UUID) (*model.ResourceNodeTemplate,
	*model.ResourceNodeTemplate, []*model.ResourceNodeTemplate,
) {
	res := &model.ResourceNodeTemplate{}
	if err := m.envStore.GetData(ctx, res, map[string]any{
		"uuid": resourceNodeUUID,
	}, "id", "uuid", "icon", "name", "resource_type", "data_schema", "config_schema", "pose", "config_info", "model"); err != nil {
		logger.Errorf(ctx, "getResourceTemplate fail err: %+v", err)
		return nil, nil, nil
	}

	childRes := utils.FilterSlice(res.ConfigInfo, func(r model.ResourceConfig) (*model.ResourceNodeTemplate, bool) {
		innerConfig := &material.InnerBaseConfig{}
		if err := json.Unmarshal(r.Config, innerConfig); err != nil {
			logger.Errorf(ctx, "getResourceTemplate unmarshal innerConfig err: %+v", err)
		}

		return &model.ResourceNodeTemplate{
			BaseModel: model.BaseModel{
				ID:   res.ID,
				UUID: res.UUID,
			},
			Name:         r.Name,
			LabID:        res.LabID,
			UserID:       res.UserID,
			Header:       res.Header,
			Footer:       res.Footer,
			Icon:         "",
			Description:  res.Description,
			Model:        res.Model,
			Module:       res.Module,
			ResourceType: r.Type,
			Language:     res.Language,
			StatusTypes:  datatypes.JSON{},
			Tags:         res.Tags,
			DataSchema:   r.Data,
			ConfigSchema: r.Config,
			Pose: datatypes.NewJSONType(model.Pose{
				Layout:   "2d",
				Position: r.Position,
				Size: model.Size{
					Width:  innerConfig.SizeX,
					Height: innerConfig.SizeY,
					Depth:  innerConfig.SizeZ,
				},
				Rotation: model.Rotation{
					X: innerConfig.Rotation.X,
					Y: innerConfig.Rotation.Y,
					Z: innerConfig.Rotation.Z,
				},
			}),
			Version:    res.Version,
			ParentName: res.Name,
			ConfigInfo: datatypes.JSONSlice[model.ResourceConfig]{},
			ParentNode: &model.ResourceNodeTemplate{},
		}, true
	})

	// childRes := make([]*model.ResourceNodeTemplate, 0, 1)
	// if err := m.envStore.FindDatas(ctx, &childRes, map[string]any{
	// 	"parent_id": res[0].ID,
	// }, "id", "uuid", "parent_id", "icon", "name", "resource_type", "data_schema", "config_schema", "pose"); err != nil {
	// 	logger.Warnf(ctx, "getResourceTemplate fail err: %+v", err)
	// 	return res[0], nil, nil
	// }
	//
	// if len(childRes) != 1 {
	// 	return res[0], nil, nil
	// }
	// childrenRes := make([]*model.ResourceNodeTemplate, 0, 1)
	// if err := m.envStore.FindDatas(ctx, &childrenRes, map[string]any{
	// 	"parent_id": childRes[0].ID,
	// }, "id", "uuid", "parent_id", "icon", "name", "resource_type", "data_schema", "config_schema", "pose"); err != nil {
	// 	logger.Warnf(ctx, "getResourceTemplate fail err: %+v", err)
	// 	return res[0], nil, nil
	// }

	return res,
		utils.TernaryLazy(len(childRes) > 0, func() *model.ResourceNodeTemplate {
			return childRes[0]
		}, func() *model.ResourceNodeTemplate {
			return nil
		}),
		utils.TernaryLazy(len(childRes) > 1, func() []*model.ResourceNodeTemplate {
			return childRes[1:]
		}, func() []*model.ResourceNodeTemplate {
			return nil
		})
}

// 批量更新节点
func (m *materialImpl) upateNode(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[*material.WSUpdateNode]{}
	err := json.Unmarshal(b, req)
	if err != nil {
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	data := req.Data
	if data.UUID.IsNil() {
		return code.ParamErr.WithMsg("update node uuid is empyt")
	}

	deviceOldID := ""
	deviceOldUUID := uuid.UUID{}
	if ancestorNode, err := m.materialStore.GetAncestors(ctx, req.Data.UUID); err == nil {
		slices.Reverse(ancestorNode)
		for _, n := range ancestorNode {
			if n.Type == model.MATERIALDEVICE {
				deviceOldID = n.Name
				deviceOldUUID = n.UUID
				break
			}
		}
	}

	keys := make([]string, 0, 7)
	materialData := &model.MaterialNode{
		BaseModel: model.BaseModel{
			UUID: data.UUID,
		},
	}
	// 父节点 uuid 不为空且父节点 uuid 不等于自身 uuid
	if data.ParentUUID != nil && !(*data.ParentUUID).IsNil() && *data.ParentUUID != data.UUID {
		parentID, err := m.materialStore.GetNodeIDByUUID(ctx, *data.ParentUUID)
		if err != nil {
			return code.ParamErr.WithMsgf("update node uuid is empyt uuid: %s", *data.ParentUUID)
		}
		keys = append(keys, "parent_id")
		materialData.ParentID = parentID
	}

	if data.DisplayName != nil {
		keys = append(keys, "display_name")
		materialData.DisplayName = *data.DisplayName
	}
	if data.Description != nil {
		keys = append(keys, "description")
		materialData.Description = data.Description
	}
	if data.InitParamData != nil {
		keys = append(keys, "init_param_data")
		materialData.InitParamData = *data.InitParamData
	}
	if data.Data != nil {
		keys = append(keys, "data")
		materialData.Data = *data.Data
	}
	if data.Pose != nil {
		keys = append(keys, "pose")
		materialData.Pose = *data.Pose
	}
	if data.Schema != nil {
		keys = append(keys, "schema")
		materialData.Schema = *data.Schema
	}
	if data.Extra != nil {
		keys = append(keys, "extra")
		materialData.Extra = *data.Extra
	}

	err = m.materialStore.UpdateNodeByUUID(ctx, materialData, keys...)
	if err != nil {
		return err
	}

	// 广播
	labUUID, _ := s.Get("lab_uuid")
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		logger.Warnf(ctx, "batchDelNode broadcast can not get user info")
		return code.UnLogin
	}

	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labUUID.(uuid.UUID),
		UUID:    uuid.NewV4(),
		Data:    req,
	}); err != nil {
		logger.Errorf(ctx, "updateNode notify fail err: %+v", err)
	}

	if ancestorNode, err := m.materialStore.GetAncestors(ctx, req.Data.UUID); err == nil {
		slices.Reverse(ancestorNode)
		deviceID := ""
		deviceUUID := uuid.UUID{}
		for _, n := range ancestorNode {
			if n.Type == model.MATERIALDEVICE {
				deviceID = n.Name
				deviceUUID = n.UUID
				break
			}
		}

		m.notify(ctx, labUUID.(uuid.UUID), engine.UpdateMaterial, []*engine.MaterialUpdate{
			{
				UUID:          req.Data.UUID,
				DeviceOldUUID: deviceOldUUID,
				DeviceOldID:   deviceOldID,
				DeviceID:      deviceID,
				DeviceUUID:    deviceUUID,
			},
		})
	}

	return nil
}

// 批量删除节点
func (m *materialImpl) batchDelNode(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	data := &common.WSData[[]uuid.UUID]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelNode unmarshal data err: %+v", err)
		return nil, code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	labUUID, _ := s.Get("lab_uuid")
	labID := m.materialStore.UUID2ID(ctx, &model.Laboratory{}, labUUID.(uuid.UUID))[labUUID.(uuid.UUID)]

	allNodes := make([]*model.MaterialNode, 0, 10)
	_ = m.materialStore.FindDatas(ctx, &allNodes, map[string]any{
		"lab_id": labID,
	}, "id", "uuid", "name", "type", "parent_id")

	allNodeIDMap := utils.Slice2Map(allNodes, func(n *model.MaterialNode) (int64, *model.MaterialNode) {
		return n.ID, n
	})

	allNodeUUIDMap := utils.Slice2Map(allNodes, func(n *model.MaterialNode) (uuid.UUID, *model.MaterialNode) {
		return n.UUID, n
	})

	res, err := m.materialStore.DelNodes(ctx, data.Data)
	if err != nil {
		return nil, err
	}

	// 广播
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		logger.Warnf(ctx, "batchDelNode broadcast can not get user info")
		return nil, code.UnLogin
	}

	resData := &common.WSData[*repo.DelNodeInfo]{
		WsMsgType: common.WsMsgType{Action: data.Action, MsgUUID: data.MsgUUID},
		Data:      res,
	}
	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labUUID.(uuid.UUID),
		Data:    resData,
	}); err != nil {
		logger.Errorf(ctx, "batchDelEdge notify fail err: %+v", err)
	}

	updates := make([]*engine.MaterialUpdate, 0, len(res.NodeUUIDs))
	for _, nodeUUID := range res.NodeUUIDs {
		node, ok := allNodeUUIDMap[nodeUUID]
		if !ok {
			continue
		}

		parentNode, ok := allNodeIDMap[node.ParentID]
		deviceUUID, deviceID := uuid.UUID{}, ""
		if ok {
			deviceUUID, deviceID = getDeviceParent(ctx, parentNode.ID, allNodeIDMap, 1000)
		}

		updates = append(updates, &engine.MaterialUpdate{
			UUID:       nodeUUID,
			DeviceUUID: deviceUUID,
			DeviceID:   deviceID,
		})
	}

	m.notify(ctx, labUUID.(uuid.UUID), engine.RemoveMaterial, updates)

	return res, nil
}

func getDeviceParent(ctx context.Context, nodeID int64, nodeMap map[int64]*model.MaterialNode, deep int) (uuid.UUID, string) {
	deep = deep - 1
	if deep <= 0 {
		return uuid.UUID{}, ""
	}

	node, ok := nodeMap[nodeID]
	if !ok {
		return uuid.UUID{}, ""
	}

	if node.Type == model.MATERIALDEVICE {
		return node.UUID, node.Name
	}

	if node.ParentID == 0 {
		return uuid.UUID{}, ""
	}

	return getDeviceParent(ctx, node.ParentID, nodeMap, deep)
}

// 批量创建 edge
func (m *materialImpl) batchCreateEdge(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	userInfo := auth.GetCurrentUser(ctx)
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return nil, err
	}
	data := &common.WSData[[]material.WSEdge]{}
	err = json.Unmarshal(b, data)
	if err != nil {
		return nil, code.UnmarshalWSDataErr.WithMsg(err.Error())
	}
	resDatas, err := m.addWSEdges(ctx, data.Data)
	if err != nil {
		return nil, err
	}

	wsData := &common.WSData[[]material.WSEdge]{
		WsMsgType: common.WsMsgType{
			Action:  data.Action,
			MsgUUID: data.MsgUUID,
		},
		Data: resDatas,
	}

	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labData.UUID,
		Data:    wsData,
	}); err != nil {
		logger.Errorf(ctx, "batchCreateEdge broadcast err: %+v", err)
	}

	return resDatas, nil
}

func (m *materialImpl) batchDelEdge(ctx context.Context, s *melody.Session, b []byte) (any, error) {
	data := &common.WSData[[]uuid.UUID]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		return nil, code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	if err := m.materialStore.DelEdges(ctx, data.Data); err != nil {
		return nil, err
	}

	userInfo := auth.GetCurrentUser(ctx)
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return nil, err
	}

	resData := &common.WSData[[]uuid.UUID]{
		WsMsgType: common.WsMsgType{
			Action:  data.Action,
			MsgUUID: data.MsgUUID,
		},
		Data: data.Data,
	}
	if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labData.UUID,
		Data:    resData,
	}); err != nil {
		logger.Errorf(ctx, "batchDelEdge fail err: %+v", err)
	}

	return data.Data, nil
}

// 接受到 redis 广播通知消息
func (m *materialImpl) OnMaterialNotify(ctx context.Context, msg string) error {
	notifyData := &notify.SendMsg{}
	if err := json.Unmarshal([]byte(msg), notifyData); err != nil {
		logger.Errorf(ctx, "HandleNotify unmarshal data err: %+v", err)
		return err
	}

	d := &common.Resp{
		Code:      code.Success,
		Data:      notifyData.Data,
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(d)
	return m.wsClient.BroadcastFilter(data, func(s *melody.Session) bool {
		sessionValue, ok := s.Get("lab_uuid")
		if !ok {
			return false
		}

		userInfo, ok := s.Get(auth.USERKEY)
		if !ok {
			return false
		}

		if sessionValue.(uuid.UUID) != notifyData.LabUUID {
			return false
		}

		u, ok := userInfo.(*model.UserData)
		if !ok || u == nil {
			return false
		}

		// edge 侧的数据，需要所有用户都更新
		if notifyData.UserID == "" {
			return true
		}

		// 发送给除自己外的所有用户
		return notifyData.UserID != u.ID
	})
}

func (m *materialImpl) checkWSConnet(ctx context.Context, s *melody.Session) error {
	labIDUUIDStr, ok := s.Get("lab_uuid")
	if !ok {
		return code.CanNotGetLabIDErr
	}

	labUUID, ok := labIDUUIDStr.(uuid.UUID)
	if !ok {
		return code.CanNotGetLabIDErr
	}

	_, err := m.envStore.GetLabByUUID(ctx, labUUID, "id")
	if err != nil {
		return code.CanNotGetLabIDErr
	}
	return nil
}

func (m *materialImpl) OnWSConnect(ctx context.Context, s *melody.Session) error {
	// TODO: 检查用户是否有权限
	err := m.checkWSConnet(ctx, s)
	if err == nil {
		return nil
	}

	d := &common.Resp{
		Code: code.UnDefineErr,
		Error: &common.Error{
			Msg: err.Error(),
		},
		Timestamp: time.Now().Unix(),
	}
	b, _ := json.Marshal(d)
	return s.CloseWithMsg(b)
}

func (m *materialImpl) addWSEdges(ctx context.Context, edges []material.WSEdge) ([]material.WSEdge, error) {
	nodeUUIDs := make([]uuid.UUID, 0, 2*len(edges))
	for _, e := range edges {
		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, e.SourceNodeUUID)
		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, e.TargetNodeUUID)
	}

	edgeInfo, err := m.materialStore.GetNodeHandlesByUUID(ctx, nodeUUIDs)
	if err != nil {
		return nil, err
	}
	edgeDatas := make([]*model.MaterialEdge, 0, len(edges))
	for _, edge := range edges {
		sourceNode, ok := edgeInfo[edge.SourceNodeUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges source not exist source node uuid: %s", edge.SourceNodeUUID)
			return nil, code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("source node uuid: %s", edge.SourceNodeUUID))
		}

		sourceHandle, ok := sourceNode[edge.SourceHandleUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges source handle not exist source uuid: %s",
				edge.SourceHandleUUID)
			return nil, code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("source handle uuid: %s",
				edge.SourceHandleUUID))
		}

		targetNode, ok := edgeInfo[edge.TargetNodeUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges target not exist target node uuid: %s", edge.TargetNodeUUID)
			return nil, code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("target node uuid: %s", edge.TargetNodeUUID))
		}

		targetHandle, ok := targetNode[edge.TargetHandleUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges target handle not exist uuid: %s", edge.TargetHandleUUID)
			return nil, code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("target handle uuid: %s",
				edge.TargetHandleUUID))
		}

		edgeDatas = append(edgeDatas, &model.MaterialEdge{
			SourceNodeUUID:   sourceHandle.NodeUUID,
			SourceHandleUUID: sourceHandle.HandleUUID,
			TargetNodeUUID:   targetHandle.NodeUUID,
			TargetHandleUUID: targetHandle.HandleUUID,
		})
	}

	if err := m.materialStore.UpsertMaterialEdge(ctx, edgeDatas); err != nil {
		return nil, err
	}

	res := utils.FilterSlice(edgeDatas, func(data *model.MaterialEdge) (material.WSEdge, bool) {
		return material.WSEdge{
			UUID:             data.UUID,
			SourceNodeUUID:   data.SourceNodeUUID,
			TargetNodeUUID:   data.TargetNodeUUID,
			SourceHandleUUID: data.SourceHandleUUID,
			TargetHandleUUID: data.TargetHandleUUID,
			Type:             "step",
		}, true
	})

	return res, nil
}

func (m *materialImpl) DownloadMaterial(ctx context.Context, req *material.DownloadMaterial) (*material.GraphNodeReq, error) {
	respMap := m.materialStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)
	if _, ok := respMap[req.LabUUID]; !ok {
		return nil, code.CanNotGetLabIDErr
	}

	// TODO: 获取组态图
	nodes, err := m.materialStore.GetNodesByLabID(ctx, respMap[req.LabUUID])
	if err != nil {
		return nil, err
	}

	nodesUUIDs := utils.FilterSlice(nodes, func(node *model.MaterialNode) (uuid.UUID, bool) {
		return node.UUID, true
	})

	edges, err := m.materialStore.GetEdgesByNodeUUID(ctx, nodesUUIDs)
	if err != nil {
		return nil, err
	}

	nodeMap := utils.Slice2Map(nodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
		return node.ID, node
	})

	nodeUUIDMap := utils.Slice2Map(nodes, func(node *model.MaterialNode) (uuid.UUID, *model.MaterialNode) {
		return node.UUID, node
	})

	formatNodes := utils.FilterSlice(nodes, func(node *model.MaterialNode) (*material.Node, bool) {
		parentName := ""
		parentUUID := uuid.UUID{}
		parentNode, ok := nodeMap[node.ParentID]
		if ok {
			parentName = parentNode.Name
			parentUUID = parentNode.UUID
		}

		return &material.Node{
			UUID:        node.UUID,
			ParentUUID:  parentUUID,
			DeviceID:    node.Name,
			Name:        node.DisplayName,
			Type:        node.Type,
			Class:       node.Class,
			Parent:      parentName,
			Pose:        node.Pose,
			Config:      node.InitParamData,
			Data:        node.Data,
			Schema:      node.Schema,
			Description: node.Description,
			Model:       node.Model,
			Position:    node.Pose.Data().Position,
		}, true
	})

	handlesUUIDs := make([]uuid.UUID, 0, 2*len(nodes))
	for _, edge := range edges {
		handlesUUIDs = utils.AppendUniqSlice(handlesUUIDs, edge.SourceHandleUUID)
		handlesUUIDs = utils.AppendUniqSlice(handlesUUIDs, edge.TargetHandleUUID)
	}

	edgesData := make([]*model.ResourceHandleTemplate, 0, 2*len(nodes))
	if err := m.envStore.FindDatas(ctx, &edgesData, map[string]any{
		"uuid": handlesUUIDs,
	}, "id", "uuid", "name"); err != nil {
		return nil, err
	}

	edgeDataMap := utils.Slice2Map(edgesData, func(edge *model.ResourceHandleTemplate) (uuid.UUID, *model.ResourceHandleTemplate) {
		return edge.UUID, edge
	})

	formatEdges := utils.FilterSlice(edges, func(edge *model.MaterialEdge) (*material.Edge, bool) {
		return &material.Edge{
			Source:       nodeUUIDMap[edge.SourceNodeUUID].Name,
			Target:       nodeUUIDMap[edge.TargetNodeUUID].Name,
			SourceHandle: edgeDataMap[edge.SourceHandleUUID].Name,
			TargetHandle: edgeDataMap[edge.TargetHandleUUID].Name,
		}, true
	})

	return &material.GraphNodeReq{
		Nodes: formatNodes,
		Edges: formatEdges,
	}, nil
}

func (m *materialImpl) GetMaterialTemplate(ctx context.Context, req *material.TemplateReq) (*material.TemplateResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}
	tpl := &model.ResourceNodeTemplate{}
	if err := m.envStore.GetData(ctx, tpl, map[string]any{
		"uuid": req.TemplateUUID,
	}); err != nil {
		return nil, err
	}

	var handles []*model.ResourceHandleTemplate
	m.envStore.FindDatas(ctx, &handles, map[string]any{
		"resource_node_id": tpl.ID,
	})

	return &material.TemplateResp{
		Handles: utils.FilterSlice(handles, func(h *model.ResourceHandleTemplate) (*material.ResourceHandleTemplate, bool) {
			return &material.ResourceHandleTemplate{
				UUID:        h.UUID,
				Name:        h.Name,
				DisplayName: h.DisplayName,
				Type:        h.Type,
				IOType:      h.IOType,
				Source:      h.Source,
				Key:         h.Key,
				Side:        h.Side,
			}, true
		}),
		UUID:         tpl.UUID,
		ParentUUID:   uuid.UUID{},
		Name:         tpl.Name,
		UserID:       tpl.UserID,
		Header:       tpl.Header,
		Footer:       tpl.Footer,
		Version:      tpl.Version,
		Icon:         tpl.Icon,
		Description:  tpl.Description,
		Model:        tpl.Model,
		Module:       tpl.Module,
		Language:     tpl.Language,
		StatusTypes:  tpl.StatusTypes,
		Tags:         tpl.Tags,
		DataSchema:   tpl.DataSchema,
		ConfigSchema: tpl.ConfigSchema,
		ResourceType: tpl.ResourceType,
		ConfigInfos:  tpl.ConfigInfo,
		Pose:         tpl.Pose,
	}, nil
}

// 该接口只能 bohrium 账号系统使用
// func (m *materialImpl) StartMachine(ctx context.Context, req *material.StartMachineReq) (*material.StartMachineRes, error) {
// 	userInfo := auth.GetCurrentUser(ctx)
// 	if userInfo == nil {
// 		return nil, code.UnLogin
// 	}
// 	labData, err := m.envStore.GetLabByUUID(ctx, req.LabUUID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	data := &model.MaterialMachine{}
// 	err = m.materialStore.ExecTx(ctx, func(txCtx context.Context) error {
// 		machine := config.Global().Dynamic().Machine
// 		err := m.materialStore.GetData(ctx, data, map[string]any{
// 			"lab_id":   labData.ID,
// 			"user_id":  userInfo.ID,
// 			"image_id": machine.ImageID,
// 		}, "machine_id", "id", "uuid")
// 		if err != nil && !errors.Is(err, code.RecordNotFound) {
// 			return err
// 		}

// 		// 开发机不存在
// 		if data.ID == 0 || data.MachineID == 0 {
// 			if err := m.machine.JoinProject(txCtx, &model.JoninProjectReq{
// 				UserID:    userInfo.ID,
// 				OrgID:     userInfo.OrgID,
// 				ProjectID: machine.ProjectID,
// 			}); err != nil {
// 				return err
// 			}

// 			data, err = m.createMachine(txCtx, userInfo, labData)
// 			return err
// 		}

// 		resp, err := m.machine.MachineStatus(ctx, &model.MachineStatusReq{
// 			UserID:    userInfo.ID,
// 			OrgID:     userInfo.OrgID,
// 			MachineID: data.MachineID,
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		if slices.Contains([]model.NodeStatus{
// 			model.NODE_STATUS_FAIL,
// 		}, resp.Status) {
// 			return code.QueryMachineStatusFailErr
// 		}

// 		switch resp.Status {
// 		case model.NODE_STATUS_PENDING, model.NODE_STATUS_RUNNING:
// 			// 正在开启中
// 			return nil
// 		case model.NODE_STATUS_STOPPING, model.NODE_STATUS_IMAGE_BUILDING:
// 			// 正在停止中
// 			return code.MachineNodeStoppingErr
// 		case model.NODE_STATUS_STOPPED:
// 			// 重新开启
// 			machine := config.Global().Dynamic().Machine
// 			cmdAkSk := fmt.Sprintf("--ak %s --sk %s", labData.AccessKey, labData.AccessSecret)
// 			return m.machine.RestartMachine(ctx, &model.RestartMachineReq{
// 				MachineID:    data.MachineID,
// 				UserID:       userInfo.ID,
// 				OrgID:        userInfo.OrgID,
// 				SkuID:        machine.SkuID,
// 				DiskSize:     machine.DiskSize,
// 				ProjectID:    machine.ProjectID,
// 				Device:       "container",
// 				TurnoffAfter: machine.TurnoffAfter,
// 				Cmd:          strings.Join([]string{machine.Cmd, cmdAkSk, ">> /root/unilab.log 2>&1"}, " "),
// 			})
// 		case model.NODE_STATUS_DELETED, model.NODE_STATUS_UNKNOW, model.NODE_STATUS_FAIL:
// 			// 重新创建
// 			data, err = m.createMachine(txCtx, userInfo, labData)
// 			return err
// 		default:
// 			// 返回错误
// 			return code.MachineStartUnknownErr
// 		}
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &material.StartMachineRes{
// 		MachineUUID: data.UUID,
// 	}, nil
// }

// func (m *materialImpl) createMachine(ctx context.Context, userInfo *model.UserData, labData *model.Laboratory) (*model.MaterialMachine, error) {
// 	machine := config.Global().Dynamic().Machine
// 	cmdAkSk := fmt.Sprintf("--upload_registry --ak %s --sk %s", labData.AccessKey, labData.AccessSecret)
// 	machineID, err := m.machine.CreateMachine(ctx, &model.CreateMachineReq{
// 		UserID:       userInfo.ID,
// 		OrgID:        userInfo.OrgID,
// 		Name:         "uni-lab-node",
// 		ImageID:      machine.ImageID,
// 		SkuID:        machine.SkuID,
// 		ProjectID:    machine.ProjectID,
// 		TurnoffAfter: machine.TurnoffAfter,
// 		DiskSize:     machine.DiskSize,
// 		Cmd:          strings.Join([]string{machine.Cmd, cmdAkSk, ">> /root/unilab.log 2>&1"}, " "),
// 		Device:       "container",
// 		Platform:     "ali",
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	data := &model.MaterialMachine{
// 		LabID:     labData.ID,
// 		UserID:    userInfo.ID,
// 		ImageID:   int64(machine.ImageID),
// 		MachineID: machineID,
// 	}

// 	return data, m.materialStore.UpsertMachine(ctx, data)
// }

// func (m *materialImpl) DelMachine(ctx context.Context, req *material.DelMachineReq) error {
// 	userInfo := auth.GetCurrentUser(ctx)
// 	if userInfo == nil {
// 		return code.UnLogin
// 	}
// 	labID, ok := m.materialStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
// 	if !ok {
// 		return code.LabNotFound
// 	}

// 	machine := config.Global().Dynamic().Machine
// 	data := &model.MaterialMachine{}
// 	err := m.materialStore.GetData(ctx, data, map[string]any{
// 		"lab_id":   labID,
// 		"user_id":  userInfo.ID,
// 		"image_id": machine.ImageID,
// 	}, "machine_id", "id")
// 	if err != nil {
// 		return err
// 	}

// 	if data.MachineID == 0 {
// 		return code.MachineNotExistErr
// 	}

// 	if err := m.machine.DelMachine(ctx, &model.DelMachineReq{
// 		UserID:    userInfo.ID,
// 		OrgID:     userInfo.OrgID,
// 		MachineID: data.MachineID,
// 		ProjectID: machine.ProjectID,
// 	}); err != nil {
// 		return err
// 	}

// 	if err := m.materialStore.UpsertMachine(ctx, &model.MaterialMachine{
// 		LabID:     labID,
// 		UserID:    userInfo.ID,
// 		ImageID:   int64(machine.ImageID),
// 		MachineID: 0,
// 	}); err != nil {
// 		logger.Errorf(ctx, "DelMachine delete machine err: %+v", err)
// 	}

// 	return nil
// }

// func (m *materialImpl) StopMachine(ctx context.Context, req *material.StopMachineReq) error {
// 	userInfo := auth.GetCurrentUser(ctx)
// 	if userInfo == nil {
// 		return code.UnLogin
// 	}

// 	labID, ok := m.materialStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
// 	if !ok {
// 		return code.LabNotFound
// 	}

// 	machine := config.Global().Dynamic().Machine
// 	data := &model.MaterialMachine{}
// 	err := m.materialStore.GetData(ctx, data, map[string]any{
// 		"lab_id":   labID,
// 		"user_id":  userInfo.ID,
// 		"image_id": machine.ImageID,
// 	}, "machine_id", "id")
// 	if err != nil {
// 		return err
// 	}

// 	if data.MachineID == 0 {
// 		return code.MachineNotExistErr
// 	}

// 	if err := m.machine.StopMachine(ctx, &model.StopMachineReq{
// 		UserID:    userInfo.ID,
// 		OrgID:     userInfo.OrgID,
// 		MachineID: data.MachineID,
// 		ProjectID: machine.ProjectID,
// 	}); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (m *materialImpl) MachineStatus(ctx context.Context, req *material.MachineStatusReq) (*material.MachineStatusRes, error) {
// 	userInfo := auth.GetCurrentUser(ctx)
// 	if userInfo == nil {
// 		return nil, code.UnLogin
// 	}

// 	labID, ok := m.materialStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
// 	if !ok {
// 		return nil, code.LabNotFound
// 	}

// 	machine := config.Global().Dynamic().Machine
// 	data := &model.MaterialMachine{}
// 	err := m.materialStore.GetData(ctx, data, map[string]any{
// 		"lab_id":   labID,
// 		"user_id":  userInfo.ID,
// 		"image_id": machine.ImageID,
// 	}, "machine_id", "id")
// 	if err != nil && errors.Is(err, code.RecordNotFound) {
// 		return &material.MachineStatusRes{Status: material.NotExist}, nil
// 	}

// 	if data.MachineID == 0 {
// 		return &material.MachineStatusRes{Status: material.Deleted}, err
// 	}

// 	resp, err := m.machine.MachineStatus(ctx, &model.MachineStatusReq{
// 		UserID:    userInfo.ID,
// 		OrgID:     userInfo.OrgID,
// 		MachineID: data.MachineID,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	if resp.Status == model.NODE_STATUS_DELETED {
// 		if err := m.materialStore.UpsertMachine(ctx, &model.MaterialMachine{
// 			LabID:     labID,
// 			UserID:    userInfo.ID,
// 			ImageID:   int64(machine.ImageID),
// 			MachineID: 0,
// 		}); err != nil {
// 			logger.Warnf(ctx, "MachineStatus update data fail err: %+v", err)
// 		}
// 		return &material.MachineStatusRes{Status: material.Deleted}, err
// 	}

// 	if slices.Contains([]model.NodeStatus{
// 		model.NODE_STATUS_UNKNOW,
// 		model.NODE_STATUS_FAIL,
// 	}, resp.Status) {

// 		return &material.MachineStatusRes{Status: material.UnknowStatus}, err
// 	}

// 	switch resp.Status {
// 	case model.NODE_STATUS_PENDING:
// 		return &material.MachineStatusRes{Status: material.Pending}, err
// 	case model.NODE_STATUS_RUNNING:
// 		return &material.MachineStatusRes{Status: material.Running}, err
// 	case model.NODE_STATUS_STOPPING:
// 		return &material.MachineStatusRes{Status: material.Stoping}, err
// 	case model.NODE_STATUS_IMAGE_BUILDING:
// 		return &material.MachineStatusRes{Status: material.Building}, err
// 	case model.NODE_STATUS_STOPPED:
// 		return &material.MachineStatusRes{Status: material.Stoped}, err
// 	default:
// 		return &material.MachineStatusRes{Status: material.UnknowStatus}, err
// 	}
// }

func (m *materialImpl) ResourceList(ctx context.Context, req *material.ResourceReq) (*material.ResourceResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	var labID int64
	if userInfo.LabID == 0 {
		labID = m.materialStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
	} else {
		labID = userInfo.LabID
	}

	count, err := m.envStore.Count(ctx, &model.LaboratoryMember{}, map[string]any{
		"lab_id":  labID,
		"user_id": userInfo.ID,
	})
	if err != nil || count == 0 {
		return nil, code.NoPermission
	}

	if labID == 0 {
		return nil, code.LabNotFound
	}

	var materialType any
	if req.Type == model.MATERIALDEVICE {
		materialType = req.Type
	} else {
		materialType = []model.DEVICETYPE{
			model.MATERIALREPO,
			model.MATERIALPLATE,
			model.MATERIALCONTAINER,
			model.MATERIALRESOURCE,
			model.MATERIALDEVICE,
			model.MATERIALWELL,
			model.MATERIALTIP,
			model.MATERIALTIPRACK,
			model.MATERIALDECK,
			model.MATERIALWORKSTATION,
		}
	}

	nodes := make([]*model.MaterialNode, 0, 1)
	// FIXME: 增加一个索引 lab id->type
	if err := m.materialStore.FindDatas(ctx, &nodes, map[string]any{
		"lab_id": labID,
		"type":   materialType,
	}, "id", "uuid", "name", "parent_id"); err != nil {
		return nil, err
	}

	parentIDs := utils.FilterUniqSlice(nodes, func(n *model.MaterialNode) (int64, bool) {
		if n.ParentID > 0 {
			return n.ParentID, true
		}
		return 0, false
	})

	parentIDMap := m.materialStore.ID2UUID(ctx, &model.MaterialNode{}, parentIDs...)

	return &material.ResourceResp{
		ResourceNameList: utils.FilterSlice(nodes, func(n *model.MaterialNode) (*material.ResourceInfo, bool) {
			return &material.ResourceInfo{
				UUID:       n.UUID,
				Name:       n.Name,
				ParentUUID: parentIDMap[n.ParentID],
			}, true
		}),
	}, nil
}

func (m *materialImpl) DeviceAction(ctx context.Context, req *material.DeviceActionReq) (*material.DeviceActionResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}
	labData := &model.Laboratory{}
	if err := m.envStore.GetData(ctx, labData, map[string]any{
		"uuid": req.LabUUID,
	}); err != nil {
		return nil, err
	}

	count, err := m.envStore.Count(ctx, &model.LaboratoryMember{}, map[string]any{
		"lab_id":  labData.ID,
		"user_id": userInfo.ID,
	})
	if err != nil || count == 0 {
		return nil, code.NoPermission
	}

	deviceData := &model.MaterialNode{}
	if err := m.envStore.GetData(ctx, deviceData, map[string]any{
		"name":   req.Name,
		"lab_id": labData.ID,
	}); err != nil {
		return nil, err
	}

	actions := make([]*model.WorkflowNodeTemplate, 0, 1)
	// FIXME: 增加索引 lab_id -> resource_node_id
	if err := m.materialStore.FindDatas(ctx, &actions, map[string]any{
		"lab_id":           labData.ID,
		"resource_node_id": deviceData.ResourceNodeID,
	}); err != nil {
		return nil, err
	}

	return &material.DeviceActionResp{
		Name: req.Name,
		Actions: utils.FilterSlice(actions, func(n *model.WorkflowNodeTemplate) (*material.DeviceAction, bool) {
			if n.Name == "_execute_driver_command" || n.Name == "_execute_driver_command_async" {
				return nil, false
			}

			gRes := gjson.Get(string(n.Schema), "properties.goal")
			schema := datatypes.JSON{}
			if gRes.Exists() {
				schema = datatypes.JSON(gRes.String())
			}

			return &material.DeviceAction{
				Action:     n.Name,
				Schema:     schema,
				ActionType: n.Type,
			}, true
		}),
	}, nil
}

func (m *materialImpl) EdgeDownloadMaterial(ctx context.Context) (*material.DownloadMaterialResp, error) {
	labUser := auth.GetCurrentUser(ctx)
	if labUser == nil {
		return nil, code.UnLogin
	}

	nodes, err := m.materialStore.GetNodesByLabID(ctx, labUser.LabID)
	if err != nil {
		return nil, err
	}

	nodeMap := utils.Slice2Map(nodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
		return node.ID, node
	})

	formatNodes := utils.FilterSlice(nodes, func(node *model.MaterialNode) (*material.Node, bool) {
		parentName := ""
		parentUUID := uuid.UUID{}
		parentNode, ok := nodeMap[node.ParentID]
		if ok {
			parentName = parentNode.Name
			parentUUID = parentNode.UUID
		}

		return &material.Node{
			UUID:        node.UUID,
			ParentUUID:  parentUUID,
			DeviceID:    node.Name,
			Name:        node.DisplayName,
			Type:        node.Type,
			Class:       node.Class,
			Parent:      parentName,
			Pose:        node.Pose,
			Config:      node.InitParamData,
			Data:        node.Data,
			Schema:      node.Schema,
			Description: node.Description,
			Model:       node.Model,
			Position:    node.Pose.Data().Position,
			Extra:       node.Extra,
		}, true
	})

	return &material.DownloadMaterialResp{
		Nodes: formatNodes,
	}, nil
}
