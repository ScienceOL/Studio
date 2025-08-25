package material

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/material"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	mStore "github.com/scienceol/studio/service/pkg/repo/material"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/datatypes"
)

type materialImpl struct {
	envStore      repo.LaboratoryRepo
	materialStore repo.MaterialRepo
	wsClient      *melody.Melody
	msgCenter     notify.MsgCenter
}

func NewMaterial(ctx context.Context, wsClient *melody.Melody) material.Service {
	m := &materialImpl{
		envStore:      eStore.New(),
		materialStore: mStore.NewMaterialImpl(),
		wsClient:      wsClient,
		msgCenter:     events.NewEvents(),
	}
	events.NewEvents().Registry(ctx, notify.MaterialModify, m.HandleNotify)

	return m
}

func (m *materialImpl) CreateMaterial(ctx context.Context, req *material.GraphNodeReq) error {
	labUser := auth.GetLabUser(ctx)
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

func (m *materialImpl) RecalculatePosition(ctx context.Context, req *material.GraphNodeReq) {
	index := 0
	_ = utils.FilterSlice(req.Nodes, func(node *material.Node) (*material.Node, bool) {
		pose := node.Pose.Data()
		pose.Layout = utils.Or(pose.Layout, "2d")
		pose.Position = model.Position{
			X: utils.Or(pose.Position.X, 0),
			Y: utils.Or(pose.Position.Y, 0),
			Z: utils.Or(pose.Position.Z, 0),
		}
		pose.Size = model.Size{
			Width:  utils.Or(pose.Size.Width, 200),
			Height: utils.Or(pose.Size.Height, 200),
		}
		pose.Scale = model.Scale{
			X: utils.Or(pose.Scale.X, 1),
			Y: utils.Or(pose.Scale.Y, 1),
			Z: utils.Or(pose.Scale.Z, 1),
		}
		pose.Rotation = model.Rotation{
			X: utils.Or(pose.Rotation.X, 0),
			Y: utils.Or(pose.Rotation.Y, 0),
			Z: utils.Or(pose.Rotation.Z, 0),
		}

		pose.Position.X = float32((pose.Size.Width + 20) * (index % 10))
		pose.Position.Y = float32(2 * pose.Size.Height * (index / 10))

		node.Pose = datatypes.NewJSONType(pose)
		index += 1
		return nil, false
	})
}

func (m *materialImpl) createNodes(ctx context.Context, labData *model.Laboratory, req *material.GraphNodeReq) error {
	resTplNames := make([]string, 0, len(req.Nodes))
	for _, data := range req.Nodes {
		// TODO: 以后所有的材料都有模板
		if data.Type == model.MATERIALDEVICE ||
			data.Type == model.MATERIALCONTAINER {
			resTplNames = utils.AppendUniqSlice(resTplNames, data.Class)
		}
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

	resMap := utils.SliceToMap(resourceNodes, func(item *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
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
				if data.Pose.Data().Layout == "" {
					poseData := data.Pose.Data()
					poseData.Layout = "2d"
					data.Pose = datatypes.NewJSONType(poseData)
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

			if err := m.materialStore.UpsertMaterialNode(txCtx, datas); err != nil {
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
	labUser := auth.GetLabUser(ctx)
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

	switch material.ActionType(msgType.Action) {
	case material.FetchGraph: // 首次获取组态图
		return m.fetchGraph(ctx, s, msgType.MsgUUID, material.FetchGraph)
	case material.FetchTemplate: // 首次获取模板
		return m.fetchDeviceTemplate(ctx, s, msgType.MsgUUID)
	case material.SaveGraph:
		return m.saveGraph(ctx, s, b)
	case material.CreateNode:
		return m.createNode(ctx, s, b)
	case material.UpdateNode: // 批量更新节点
		return m.upateNode(ctx, s, b)
	case material.BatchDelNode: // 批量删除节点
		return m.batchDelNode(ctx, s, b)
	case material.BatchCreateEdge: // 批量创建边
		return m.batchCreateEdge(ctx, s, b)
	case material.BatchDelEdge: // 批量删除边
		return m.batchDelEdge(ctx, s, b)
	default:
		return common.ReplyWSErr(s, msgType.Action, msgType.MsgUUID, code.UnknownWSActionErr)
	}
}

// 获取组态图
func (m *materialImpl) fetchGraph(ctx context.Context, s *melody.Session, msgUUID uuid.UUID, action material.ActionType) error {
	// 获取所有组态图信息
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return err
	}
	nodes, err := m.materialStore.GetNodesByLabID(ctx, labData.ID)
	if err != nil {
		common.ReplyWSErr(s, string(action), msgUUID, err)
		return err
	}

	resTplIDS := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (int64, bool) {
		if nodeItem.ResourceNodeID == 0 {
			return 0, false
		}
		return nodeItem.ResourceNodeID, true
	})
	resTplIDS = utils.RemoveDuplicates(resTplIDS)

	nodesMap := utils.SliceToMap(nodes, func(item *model.MaterialNode) (int64, *model.MaterialNode) {
		return item.ID, item
	})

	resHandlesMap, err := m.envStore.GetResourceHandleTemplates(ctx, resTplIDS)
	if err != nil {
		common.ReplyWSErr(s, string(action), msgUUID, err)
		return err
	}

	nodeUUIDs := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (uuid.UUID, bool) {
		return nodeItem.UUID, true
	})
	edges, err := m.materialStore.GetEdgesByNodeUUID(ctx, nodeUUIDs)
	if err != nil {
		common.ReplyWSErr(s, string(action), msgUUID, err)
		return err
	}
	resNodes, err := m.envStore.GetResourceNodeTemplates(ctx, resTplIDS)
	if err != nil {
		common.ReplyWSErr(s, string(action), msgUUID, err)
		return err
	}

	resNodeTplMap := utils.SliceToMap(resNodes, func(item *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
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

	return common.ReplyWSOk(s, string(action), msgUUID, resp)
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

func (m *materialImpl) buildTplNode(ctx context.Context, nodes []*model.ResourceNodeTemplate) []*model.ResourceNodeTemplate {
	nodeMap := utils.SliceToMap(nodes, func(node *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return node.ID, node
	})

	rootNodes := make([]*model.ResourceNodeTemplate, 0, len(nodes))

	for _, n := range nodes {
		if n.ParentID != 0 {
			node, ok := nodeMap[n.ParentID]
			if ok {
				node.ConfigInfo = append(node.ConfigInfo, n)
			} else {
				logger.Errorf(ctx, "buildTplNode can not found parent node id: %d, parent id: %d", n.ID, n.ParentID)
			}
		} else {
			rootNodes = append(rootNodes, n)
		}
	}
	return rootNodes
}

func (m *materialImpl) getChildren(ctx context.Context, node *model.ResourceNodeTemplate, maxDeep int) ([]*model.ResourceNodeTemplate, error) {
	if maxDeep <= 0 {
		return nil, code.MaxTplNodeDeepErr
	}

	children := make([]*model.ResourceNodeTemplate, 0, len(node.ConfigInfo))
	for _, child := range node.ConfigInfo {
		if child == nil {
			continue
		}

		if len(child.ConfigInfo) > 0 {
			deepChildren, err := m.getChildren(ctx, child, maxDeep-1)
			if err != nil {
				return nil, err
			}
			children = append(children, deepChildren...)
		}

		children = append(children, child)
	}
	return children, nil
}

// 获取设备模板
func (m *materialImpl) fetchDeviceTemplate(ctx context.Context, s *melody.Session, msgUUID uuid.UUID) error {
	labData, err := m.getLab(ctx, s)
	if err != nil {
		common.ReplyWSErr(s, string(material.FetchTemplate), msgUUID, err)
		return err
	}

	tplNodes, err := m.envStore.GetAllResourceTemplateByLabID(ctx, labData.ID)
	if err != nil {
		common.ReplyWSErr(s, string(material.FetchTemplate), msgUUID, err)
		return err
	}
	tplIDs := utils.FilterSlice(tplNodes, func(item *model.ResourceNodeTemplate) (int64, bool) {
		return item.ID, true
	})

	tplHandles, err := m.envStore.GetResourceHandleTemplates(ctx, tplIDs)
	if err != nil {
		common.ReplyWSErr(s, string(material.FetchTemplate), msgUUID, err)
		return err
	}

	tplNodeMap := utils.SliceToMap(tplNodes, func(item *model.ResourceNodeTemplate) (int64, *model.ResourceNodeTemplate) {
		return item.ID, item
	})

	rootNode := m.buildTplNode(ctx, tplNodes)

	tplDatas, err := utils.FilterSliceWithErr(rootNode, func(nodeItem *model.ResourceNodeTemplate) ([]*material.DeviceTemplate, bool, error) {
		childrenNodes, err := m.getChildren(ctx, nodeItem, 5)
		if err != nil {
			return nil, false, err
		}

		return []*material.DeviceTemplate{&material.DeviceTemplate{
			Handles: utils.FilterSlice(tplHandles[nodeItem.ID], func(handleItem *model.ResourceHandleTemplate) (*material.DeviceHandleTemplate, bool) {
				// FIXME: 此处效率可以优化
				if handleItem.ResourceNodeID != nodeItem.ID {
					return nil, false
				}
				return &material.DeviceHandleTemplate{
					UUID:        handleItem.UUID,
					Name:        handleItem.Name,
					DisplayName: handleItem.DisplayName,
					Type:        handleItem.Type,
					IOType:      handleItem.IOType,
					Source:      handleItem.Source,
					Key:         handleItem.Key,
					Side:        handleItem.Side,
				}, true
			}),
			UUID: nodeItem.UUID,
			ParentUUID: utils.SafeValue(func() uuid.UUID {
				if node, ok := tplNodeMap[nodeItem.ParentID]; ok {
					return node.UUID
				} else {
					return uuid.UUID{}
				}
			}, uuid.UUID{}),
			Name:         nodeItem.Name,
			UserID:       nodeItem.UserID,
			Header:       nodeItem.Header,
			Footer:       nodeItem.Footer,
			Version:      nodeItem.Version,
			Icon:         nodeItem.Icon,
			Description:  nodeItem.Description,
			Model:        nodeItem.Model,
			Module:       nodeItem.Module,
			Language:     nodeItem.Language,
			StatusTypes:  nodeItem.StatusTypes,
			Tags:         nodeItem.Tags,
			DataSchema:   nodeItem.DataSchema,
			ConfigSchema: nodeItem.ConfigSchema,
			ResourceType: nodeItem.ResourceType,
			ConfigInfos: utils.FilterSlice(childrenNodes, func(child *model.ResourceNodeTemplate) (*material.DeviceTemplate, bool) {
				return &material.DeviceTemplate{
					UUID: child.UUID,
					ParentUUID: utils.SafeValue(func() uuid.UUID {
						if node, ok := tplNodeMap[child.ParentID]; ok && node.ParentID != 0 {
							return node.UUID
						} else {
							return uuid.UUID{}
						}
					}, uuid.UUID{}),
					Name:         child.Name,
					UserID:       child.UserID,
					Header:       child.Header,
					Footer:       child.Footer,
					Version:      child.Version,
					Icon:         child.Icon,
					Description:  child.Description,
					Model:        child.Model,
					Module:       child.Module,
					Language:     child.Language,
					StatusTypes:  child.StatusTypes,
					Tags:         child.Tags,
					DataSchema:   child.DataSchema,
					ConfigSchema: child.ConfigSchema,
					ResourceType: child.ResourceType,
					Pose:         child.Pose,
				}, true

			}),
		}}, true, nil
	})

	if err != nil {
		common.ReplyWSErr(s, string(material.FetchTemplate), msgUUID, err)
		return err
	}

	resData := &material.DeviceTemplates{
		Templates: tplDatas,
	}

	return common.ReplyWSOk(s, string(material.FetchTemplate), msgUUID, resData)
}

// 全量保存
func (m *materialImpl) saveGraph(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[*material.WSGraph]{}
	if err := json.Unmarshal(b, req); err != nil {
		logger.Errorf(ctx, "saveGraph unmarshal data fail err: %+v", err)
		return common.ReplyWSErr(s, string(material.SaveGraph), req.MsgUUID, code.ParamErr.WithErr(err))
	}
	labData, err := m.getLab(ctx, s)
	if err != nil {
		logger.Errorf(ctx, "saveGraph get lab fail err: %+v", err)
		return common.ReplyWSErr(s, string(material.SaveGraph), req.MsgUUID, err)
	}

	nodeUUIDs := make([]uuid.UUID, 0, len(req.Data.Nodes))
	tplUUIDs := make([]uuid.UUID, 0, len(req.Data.Nodes))
	for _, n := range req.Data.Nodes {
		if n.UUID.IsNil() {
			logger.Errorf(ctx, "saveGraph check node uuid id empty")
			return common.ReplyWSErr(s, string(material.SaveGraph), req.MsgUUID, code.ParamErr)
		}

		if !n.UUID.IsNil() {
			nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, n.UUID)
		}

		if !n.ResTemplateUUID.IsNil() {
			tplUUIDs = utils.AppendUniqSlice(tplUUIDs, n.ResTemplateUUID)
		}
	}

	mUUID2IDMap := m.materialStore.UUID2ID(ctx, &model.MaterialNode{}, nodeUUIDs...)
	resUUID2IDMap := m.materialStore.UUID2ID(ctx, &model.ResourceNodeTemplate{}, tplUUIDs...)

	nodes, err := utils.FilterSliceWithErr(req.Data.Nodes, func(item *material.WSNode) ([]*model.MaterialNode, bool, error) {
		if item.UUID.IsNil() || item.Name == "" {
			return nil, false, code.ParamErr.WithMsg("saveGraph node uuid is empty")
		}
		data := &model.MaterialNode{
			ParentID:       mUUID2IDMap[item.ParentUUID],
			LabID:          labData.ID,
			Name:           item.Name,
			DisplayName:    item.DisplayName,
			Description:    item.Description,
			Type:           item.Type,
			ResourceNodeID: resUUID2IDMap[item.ResTemplateUUID],
			InitParamData:  item.InitParamData,
			Data:           item.Data,
			Pose:           item.Pose,
			Model:          item.Model,
			Icon:           item.Icon,
			Schema:         item.Schema,
			// Class:                  item.Class,
		}
		return []*model.MaterialNode{data}, true, nil
	})

	if err != nil {
		logger.Errorf(ctx, "saveGraph check node param err: %+v", err)
		return common.ReplyWSErr(s, string(material.SaveGraph), req.MsgUUID, err)
	}

	if err := m.materialStore.UpsertMaterialNode(ctx, nodes); err != nil {
		logger.Errorf(ctx, "saveGraph upsert material node err: %+v", err)
		return common.ReplyWSErr(s, string(material.SaveGraph), req.MsgUUID, err)
	}

	return m.fetchGraph(ctx, s, req.MsgUUID, material.SaveGraph)
}

// 批量创建节点
func (m *materialImpl) createNode(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[*material.WSNode]{}
	err := json.Unmarshal(b, req)
	if err != nil {
		logger.Errorf(ctx, "createNode unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}
	labData, err := m.getLab(ctx, s)
	if err != nil {
		logger.Errorf(ctx, "createNode can not get lab err: %+v", err)
		return err
	}
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		common.ReplyWSErr(s, string(material.CreateNode), req.MsgUUID, code.UnLogin)
		return code.UnLogin
	}

	reqData := req.Data
	mData := &model.MaterialNode{}

	if reqData.ResTemplateUUID.IsNil() {
		common.ReplyWSErr(s, string(material.CreateNode), req.MsgUUID, code.TemplateNodeNotFoundErr)
		return code.TemplateNodeNotFoundErr
	}

	resNodeTpl, resNodeChildrenTpl := m.getResourceTemplates(ctx, reqData.ResTemplateUUID)
	if resNodeTpl == nil {
		common.ReplyWSErr(s, string(material.CreateNode), req.MsgUUID, code.TemplateNodeNotFoundErr)
		return code.TemplateNodeNotFoundErr
	} else {
		mData.ResourceNodeID = resNodeTpl.ID
		mData.Icon = resNodeTpl.Icon
	}

	if !reqData.ParentUUID.IsNil() {
		if nodeID, ok := m.materialStore.UUID2ID(ctx,
			&model.MaterialNode{},
			reqData.ParentUUID)[reqData.ParentUUID]; !ok {
			common.ReplyWSErr(s, string(material.CreateNode), req.MsgUUID, err)
			return err
		} else {
			mData.ParentID = nodeID
		}
	}

	mData.LabID = labData.ID
	mData.Name = reqData.Name
	mData.DisplayName = reqData.DisplayName
	mData.Description = reqData.Description
	mData.Type = utils.Or(reqData.Type, model.DEVICETYPE(resNodeTpl.ResourceType)) // 只有设备类型
	mData.InitParamData = reqData.InitParamData
	mData.Schema = reqData.Schema
	mData.Data = reqData.Data
	mData.Pose = reqData.Pose
	mData.Model = reqData.Model

	if err := m.materialStore.UpsertMaterialNode(ctx, []*model.MaterialNode{mData}); err != nil {
		common.ReplyWSErr(s, string(material.CreateNode), req.MsgUUID, err)
		return err
	}
	reqData.UUID = mData.UUID

	childrenMaterialNode := utils.FilterSlice(resNodeChildrenTpl, func(tpl *model.ResourceNodeTemplate) (*model.MaterialNode, bool) {
		data := &model.MaterialNode{
			ResourceNodeID:       tpl.ID,
			Icon:                 tpl.Icon,
			ParentID:             mData.ID,
			LabID:                labData.ID,
			Name:                 tpl.Name,
			DisplayName:          tpl.Name,
			Type:                 model.DEVICETYPE(tpl.ResourceType),
			InitParamData:        tpl.DataSchema,
			Schema:               tpl.ConfigSchema,
			Data:                 tpl.DataSchema,
			Pose:                 tpl.Pose,
			ResourceNodeTemplate: tpl,
		}
		return data, true
	})

	if err := m.materialStore.UpsertMaterialNode(ctx, childrenMaterialNode); err != nil {
		common.ReplyWSErr(s, string(material.CreateNode), req.MsgUUID, err)
		return err
	}

	resDatas := make([]*material.WSNode, 0, 1+len(childrenMaterialNode))
	resDatas = append(resDatas, reqData)
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
	if err = common.ReplyWSOk(s, string(material.CreateNode), req.MsgUUID, resDatas); err != nil {
		logger.Errorf(ctx, "updateNode reply ws ok fail err: %+v", err)
	}

	labUUID, _ := s.Get("lab_uuid")
	if userInfo == nil {
		logger.Warnf(ctx, "batchDelNode broadcast can not get user info")
		return nil
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

	return nil
}

func (m *materialImpl) getResourceTemplates(ctx context.Context, resourceNodeUUID uuid.UUID) (*model.ResourceNodeTemplate, []*model.ResourceNodeTemplate) {
	res := make([]*model.ResourceNodeTemplate, 0, 2)
	if err := m.envStore.FindDatas(ctx, &res, map[string]any{
		"uuid": resourceNodeUUID,
	}, "id", "uuid", "parent_id", "icon"); err != nil {
		logger.Errorf(ctx, "getResourceTemplate fail err: %+v", err)
		return nil, nil
	}

	if len(res) != 1 {
		logger.Errorf(ctx, "getResourceTemplate can not found resource node template")
		return nil, nil
	}

	// FIXME: 12 孔板父节点套 13 个字节点，是否去掉顶层节点
	childRes := make([]*model.ResourceNodeTemplate, 0, 1)
	if err := m.envStore.FindDatas(ctx, &childRes, map[string]any{
		"parent_id": res[0].ID,
	}, "id"); err != nil {
		logger.Warnf(ctx, "getResourceTemplate fail err: %+v", err)
		return res[0], nil
	}

	// FIXME: 顶层节点下只有一层套壳节点么？
	if len(childRes) != 1 {
		return res[0], nil
	}
	childrenRes := make([]*model.ResourceNodeTemplate, 0, 1)
	if err := m.envStore.FindDatas(ctx, &childrenRes, map[string]any{
		"parent_id": childRes[0].ID,
	}, []string{"id", "uuid", "icon"}...); err != nil {
		logger.Warnf(ctx, "getResourceTemplate fail err: %+v", err)
		return res[0], nil
	}

	return res[0], childrenRes
}

// 批量更新节点
func (m *materialImpl) upateNode(ctx context.Context, s *melody.Session, b []byte) error {
	req := &common.WSData[*material.WSUpdateNode]{}
	err := json.Unmarshal(b, req)
	if err != nil {
		logger.Errorf(ctx, "batchDelNode unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	data := req.Data
	if data.UUID.IsNil() {
		common.ReplyWSErr(s, string(material.UpdateNode), req.MsgUUID, code.ParamErr)
		return code.ParamErr.WithMsg("update node uuid is empyt")
	}

	keys := make([]string, 0, 7)
	materialData := &model.MaterialNode{
		BaseModel: model.BaseModel{
			UUID: data.UUID,
		},
	}
	if data.ParentUUID != nil && !(*data.ParentUUID).IsNil() {
		parentID, err := m.materialStore.GetNodeIDByUUID(ctx, *data.ParentUUID)
		if err != nil {
			common.ReplyWSErr(s,
				string(material.UpdateNode),
				req.MsgUUID,
				code.ParentNodeNotFoundErr.WithMsg((*data.ParentUUID).String()))
			return code.ParamErr.WithMsg("update node uuid is empyt")
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

	err = m.materialStore.UpdateNodeByUUID(ctx, materialData, keys...)
	if err != nil {
		common.ReplyWSErr(s, string(material.UpdateNode), req.MsgUUID, code.UpdateNodeErr.WithMsg(err.Error()))
		return err
	}

	if err := common.ReplyWSOk(s, string(material.UpdateNode), req.MsgUUID); err != nil {
		logger.Errorf(ctx, "updateNode reply ws ok fail err: %+v", err)
	}

	// 广播
	labUUID, _ := s.Get("lab_uuid")
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		logger.Warnf(ctx, "batchDelNode broadcast can not get user info")
		return nil
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

	return nil
}

// 批量删除节点
func (m *materialImpl) batchDelNode(ctx context.Context, s *melody.Session, b []byte) error {
	// FIXME: 如果删除父节点，子节点全部删除.
	data := &common.WSData[[]uuid.UUID]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelNode unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	res, err := m.materialStore.DelNodes(ctx, data.Data)
	if err != nil {
		common.ReplyWSErr(s, string(material.BatchDelNode), data.MsgUUID, err)
		return err
	}

	if err := common.ReplyWSOk(s, data.Action, data.MsgUUID, res); err != nil {
		logger.Errorf(ctx, "batchDelNode reply ws ok fail err: %+v", err)
	}
	// 广播
	labUUID, _ := s.Get("lab_uuid")
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		logger.Warnf(ctx, "batchDelNode broadcast can not get user info")
		return nil
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

	return nil
}

// 批量创建 edge
func (m *materialImpl) batchCreateEdge(ctx context.Context, s *melody.Session, b []byte) error {
	userInfo := auth.GetCurrentUser(ctx)
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return err
	}
	data := &common.WSData[[]material.WSEdge]{}
	err = json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelEdge unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}
	resDatas, err := m.addWSEdges(ctx, data.Data)
	if err != nil {
		_ = common.ReplyWSErr(s, string(material.BatchCreateEdge), data.MsgUUID, err)
		return err
	}

	if err = common.ReplyWSOk(s, string(material.BatchCreateEdge), data.MsgUUID, resDatas); err != nil {
		logger.Errorf(ctx, "batchCreateEdge send msg fail err: %+v", err)
	}

	wsData := &common.WSData[[]material.WSEdge]{
		WsMsgType: common.WsMsgType{
			Action:  data.Action,
			MsgUUID: data.MsgUUID,
		},
		Data: resDatas,
	}

	return m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labData.UUID,
		Data:    wsData,
	})
}

func (m *materialImpl) batchDelEdge(ctx context.Context, s *melody.Session, b []byte) error {
	data := &common.WSData[[]uuid.UUID]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelEdge unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	if err := m.materialStore.DelEdges(ctx, data.Data); err != nil {
		common.ReplyWSErr(s, string(material.BatchDelEdge), data.MsgUUID, err)
		return err
	}

	if err = common.ReplyWSOk(s, string(material.BatchDelEdge), data.MsgUUID, data.Data); err != nil {
		logger.Errorf(ctx, "batchDelEdge reply ws ok fail err: %+v", err)
	}

	userInfo := auth.GetCurrentUser(ctx)
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return err
	}

	resData := &common.WSData[[]uuid.UUID]{
		WsMsgType: common.WsMsgType{
			Action:  data.Action,
			MsgUUID: data.MsgUUID,
		},
		Data: data.Data,
	}
	return m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Channel: notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labData.UUID,
		Data:    resData,
	})
}

// 接受到 redis 广播通知消息
func (m *materialImpl) HandleNotify(ctx context.Context, msg string) error {
	notifyData := &notify.SendMsg{}
	if err := json.Unmarshal([]byte(msg), notifyData); err != nil {
		logger.Errorf(ctx, "HandleNotify unmarshal data err: %+v", err)
		return err
	}

	data, _ := json.Marshal(notifyData.Data)
	return m.wsClient.BroadcastFilter(data, func(s *melody.Session) bool {
		sessionValue, ok := s.Get("lab_uuid")
		if !ok {
			return false
		}

		userInfo, ok := s.Get(auth.USERKEY)
		if !ok {
			return false
		}

		if sessionValue.(uuid.UUID) == notifyData.LabUUID {
			return false
		}

		if u, ok := userInfo.(*model.UserData); !ok || u == nil {
			return false
		} else if u.ID == notifyData.UserID {
			return false
		}

		return true
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
