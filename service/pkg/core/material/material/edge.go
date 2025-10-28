package material

import (
	"context"
	"fmt"
	"time"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/material"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/utils"
)

func (m *materialImpl) EdgeCreateMaterial(ctx context.Context, req *material.CreateMaterialReq) ([]*material.CreateMaterialResp, error) {
	labUser := auth.GetLabUser(ctx)
	if labUser == nil {
		return nil, code.UnLogin
	}

	if len(req.Nodes) == 0 {
		return nil, nil
	}

	return m.createEdgeNodes(ctx, labUser, req)
}

func (m *materialImpl) createEdgeNodes(ctx context.Context, labData *model.UserData, req *material.CreateMaterialReq) ([]*material.CreateMaterialResp, error) {
	resTplNames := make([]string, 0, len(req.Nodes))
	for _, data := range req.Nodes {
		if data.Class == "" {
			continue
		}
		resTplNames = utils.AppendUniqSlice(resTplNames, data.Class)
	}

	resourceNodes := make([]*model.ResourceNodeTemplate, 0, len(resTplNames))
	err := m.envStore.FindDatas(ctx, &resourceNodes, map[string]any{
		"lab_id": labData.LabID,
		"name":   resTplNames,
	}, "id", "name", "icon", "model")
	if err != nil {
		return nil, err
	}

	// 强制校验资源模板是否存在
	if len(resourceNodes) != len(resTplNames) {
		return nil, code.ResNotExistErr
	}

	resMap := utils.Slice2Map(resourceNodes,
		func(item *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
			return item.Name, item
		})

	nodeNames := utils.FilterSlice(
		req.Nodes,
		func(item *material.Material) (
			*utils.Node[uuid.UUID, *material.Material], bool,
		) {
			return &utils.Node[uuid.UUID, *material.Material]{
				Name:   item.UUID,
				Parent: item.ParentUUID,
				Data:   item,
			}, !item.UUID.IsNil()
		})

	levelNodes, err := utils.BuildHierarchy(nodeNames)
	if err != nil {
		return nil, code.InvalidDagErr.WithMsg(err.Error())
	}

	nodeMap := make(map[uuid.UUID]*model.MaterialNode)
	if err = db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range levelNodes {
			datas := make([]*model.MaterialNode, 0, len(nodes))
			for _, n := range nodes {
				data := &model.MaterialNode{
					ParentID:       0,
					LabID:          labData.LabID,
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
				if node := nodeMap[n.ParentUUID]; node != nil {
					data.ParentID = node.ID
				}

				if resInfo := resMap[n.Class]; resInfo != nil {
					data.ResourceNodeID = resInfo.ID
					data.Icon = resInfo.Icon
					data.Model = resInfo.Model
				}

				datas = append(datas, data)
				nodeMap[n.UUID] = data
			}

			if _, err := m.materialStore.UpsertMaterialNode(txCtx, datas, nil, []string{"id", "uuid"}); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return utils.MapToSlice(
		nodeMap, func(
			key uuid.UUID, data *model.MaterialNode,
		) (*material.CreateMaterialResp, bool) {
			return &material.CreateMaterialResp{
				UUID:      key,
				CloudUUID: data.UUID,
				DeviceID:  data.Name,
				Name:      data.DisplayName,
			}, true
		}), nil
}

func (m *materialImpl) EdgeUpsertMaterial(ctx context.Context, req *material.UpsertMaterialReq) ([]*material.UpsertMaterialResp, error) {
	if len(req.Nodes) == 0 {
		// 不更新也新建
		return nil, code.ParamErr
	}

	// 运行中，可能是创建物料，也可能是更新物料
	var mountID int64
	if req.MountUUID.IsNil() {
		mountID = 0
	} else {
		mountID = m.materialStore.UUID2ID(ctx, &model.MaterialNode{}, req.MountUUID)[req.MountUUID]
		if mountID == 0 {
			return nil, code.ParamErr.WithMsg("parent node uuid not exist")
		}
	}

	return m.upsertNode(ctx, mountID, req)
	// parent id 如果不存在，只更新或者创建
	// if parentID == 0 {
	// 	return m.upsertMaterialNode(ctx, req)
	// } else {
	// 	return m.delUpsertMaterialNode(ctx, req.MountUUID, parentID, req)
	// }
}

func (m *materialImpl) upsertNode(ctx context.Context, mountID int64, req *material.UpsertMaterialReq) ([]*material.UpsertMaterialResp, error) {
	labUser := auth.GetLabUser(ctx)
	if labUser == nil {
		return nil, code.UnLogin
	}

	labID := labUser.LabID
	labUUID := labUser.LabUUID

	rootNodes := make([]*model.MaterialNode, 0, 1)
	if err := m.materialStore.FindDatas(ctx, &rootNodes, map[string]any{
		"lab_id":    labID,
		"parent_id": mountID,
	}, "id", "uuid"); err != nil {
		return nil, err
	}

	// 获取更新或创建节点的 nodeUUIDs
	nodeUUIDs := utils.FilterUniqSlice(req.Nodes, func(n *material.Material) (uuid.UUID, bool) {
		return n.UUID, !n.UUID.IsNil()
	})

	if len(nodeUUIDs) != len(req.Nodes) {
		return nil, code.ParamErr.WithMsg("exist node uuid is empty")
	}

	dbNodes := make([]*model.MaterialNode, 0, len(nodeUUIDs))
	if err := m.materialStore.FindDatas(ctx, &dbNodes, map[string]any{
		"uuid":   nodeUUIDs,
		"lab_id": labID,
	}, "id", "uuid", "parent_id"); err != nil {
		return nil, err
	}
	dbNodeMap := utils.Slice2Map(dbNodes, func(n *model.MaterialNode) (uuid.UUID, *model.MaterialNode) {
		return n.UUID, n
	})
	reqNodeMap := utils.Slice2Map(req.Nodes, func(n *material.Material) (uuid.UUID, bool) {
		return n.UUID, true
	})
	updateUUIDMap := make(map[uuid.UUID]int64)
	createUUIDMap := make(map[uuid.UUID]int64)
	delUUIDs := make([]uuid.UUID, 0, 10)
	// 识别新建节点和更新节点
	for _, nodeUUID := range nodeUUIDs {
		if n, ok := dbNodeMap[nodeUUID]; ok {
			// 更新的 node
			updateUUIDMap[nodeUUID] = n.ID
		} else {
			// 新建的 node
			createUUIDMap[nodeUUID] = 0
		}
	}

	// 识别删除节点
	for _, rootNode := range rootNodes {
		// 孩子节点不在更新列表内的不做任何操作
		_, ok := reqNodeMap[rootNode.UUID]
		if !ok {
			continue
		}
		childrenNode, err := m.materialStore.GetDescendants(ctx, labID, rootNode.ID)
		if err != nil {
			return nil, err
		}

		for _, childNode := range childrenNode {
			_, ok = reqNodeMap[childNode.UUID]
			if !ok {
				delUUIDs = append(delUUIDs, childNode.UUID)
			}
		}
	}

	nodeParentMap := make(map[uuid.UUID]*model.MaterialNode)
	if !req.MountUUID.IsNil() {
		mountNode := &model.MaterialNode{}
		mountNode.UUID = req.MountUUID
		mountNode.ID = mountID
		nodeParentMap[req.MountUUID] = mountNode
	}

	// 多dag 图构建，dag 图 edge 传递一定是合法的
	nodeNames := utils.FilterSlice(
		req.Nodes,
		func(item *material.Material) (
			*utils.Node[uuid.UUID, *material.Material], bool,
		) {
			return &utils.Node[uuid.UUID, *material.Material]{
				Name:   item.UUID,
				Parent: item.ParentUUID,
				Data:   item,
			}, !item.UUID.IsNil()
		})
	levelNodes, err := utils.BuildHierarchy(nodeNames)
	if err != nil {
		return nil, err
	}

	// 查找注册表引用
	resTplNames := utils.FilterUniqSlice(req.Nodes, func(n *material.Material) (string, bool) {
		return n.Class, n.Class != ""
	})

	resourceNodes := make([]*model.ResourceNodeTemplate, 0, len(resTplNames))
	if err := m.envStore.FindDatas(ctx, &resourceNodes, map[string]any{
		"lab_id": labID,
		"name":   resTplNames,
	}, "id", "name", "icon", "model"); err != nil {
		return nil, err
	}
	resMap := utils.Slice2Map(resourceNodes, func(r *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
		return r.Name, r
	})

	// 强制校验资源模板是否存在
	if len(resourceNodes) != len(resTplNames) {
		return nil, code.ResNotExistErr
	}

	if err = db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		// 删除 root node 下的子节点
		if len(delUUIDs) > 0 {
			delResp, err := m.materialStore.DelNodes(txCtx, delUUIDs)
			if err != nil {
				return err
			}

			resData := &common.WSData[*repo.DelNodeInfo]{
				WsMsgType: common.WsMsgType{Action: string(material.BatchDelNode), MsgUUID: uuid.NewV4()},
				Data:      delResp,
			}

			if err := m.msgCenter.Broadcast(ctx, &notify.SendMsg{
				Channel: notify.MaterialModify,
				LabUUID: labUUID,
				Data:    resData,
			}); err != nil {
				logger.Errorf(ctx, "batchDelEdge fail err: %+v", err)
			}
		}

		for _, nodes := range levelNodes {
			newDatas := make([]*model.MaterialNode, 0, len(nodes))
			updateDatas := make([]*model.MaterialNode, 0, len(nodes))
			for _, n := range nodes {
				data := &model.MaterialNode{
					ParentID:       0,
					LabID:          labID,
					Name:           n.DeviceID,
					DisplayName:    n.Name,
					Description:    n.Description,
					Class:          n.Class,
					Type:           n.Type,
					ResourceNodeID: 0,
					InitParamData:  n.Config,
					Data:           n.Data,
					Pose:           n.Pose,
					Icon:           n.Icon,
					Schema:         n.Schema,
					EdgeUUID:       n.UUID,
				}
				data.UpdatedAt = time.Now()
				if node := nodeParentMap[n.ParentUUID]; node != nil {
					data.ParentID = node.ID
				}

				if resInfo := resMap[n.Class]; resInfo != nil {
					data.ResourceNodeID = resInfo.ID
					data.Icon = utils.Or(data.Icon, resInfo.Icon)
					data.Model = resInfo.Model
				}

				if id, ok := updateUUIDMap[n.UUID]; ok {
					data.ID = id
					updateDatas = append(updateDatas, data)
				} else {
					newDatas = append(newDatas, data)
				}

				nodeParentMap[n.UUID] = data
			}

			// 更新
			if len(updateDatas) > 0 {
				updateKeys := []string{
					"display_name",
					"description",
					"status",
					"type",
					"resource_node_id",
					"class",
					"init_param_data",
					"schema",
					"data",
					"pose",
					"model",
					"icon",
					"updated_at",
					"parent_id",
					"name",
				}

				if _, err := m.materialStore.UpsertMaterialNode(txCtx,
					updateDatas,
					[]string{"id"},
					[]string{"uuid", "id"},
					updateKeys...); err != nil {
					return err
				}
			}

			// 新建
			if len(newDatas) > 0 {
				if _, err := m.materialStore.UpsertMaterialNode(txCtx, newDatas, nil, []string{"uuid", "id"}); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if !req.MountUUID.IsNil() {
		delete(nodeParentMap, req.MountUUID)
	}

	return utils.MapToSlice(nodeParentMap, func(key uuid.UUID, n *model.MaterialNode) (*material.UpsertMaterialResp, bool) {
		return &material.UpsertMaterialResp{
			UUID:        key,
			CloudUUID:   n.UUID,
			Name:        n.Name,
			DisplayName: n.DisplayName,
		}, true
	}), nil
}

func (m *materialImpl) upsertMaterialNode(ctx context.Context, req *material.UpsertMaterialReq) ([]*material.UpsertMaterialResp, error) {
	labUser := auth.GetLabUser(ctx)
	// 构建 dag 结构

	resTplNames := make([]string, 0, len(req.Nodes))
	for _, data := range req.Nodes {
		if data.Class == "" {
			continue
		}
		resTplNames = utils.AppendUniqSlice(resTplNames, data.Class)
	}

	resourceNodes := make([]*model.ResourceNodeTemplate, 0, len(resTplNames))
	err := m.envStore.FindDatas(ctx, &resourceNodes, map[string]any{
		"lab_id": labUser.LabID,
		"name":   resTplNames,
	}, "id", "name", "icon", "model")
	if err != nil {
		return nil, err
	}

	// 强制校验资源模板是否存在
	if len(resourceNodes) != len(resTplNames) {
		return nil, code.ResNotExistErr
	}

	// 查询节点是否存在
	nodeUUIDs := utils.FilterSlice(req.Nodes, func(n *material.Material) (uuid.UUID, bool) {
		return n.UUID, true
	})

	dbNodes := make([]*model.MaterialNode, 0, len(nodeUUIDs))
	if err := m.materialStore.FindDatas(ctx, &dbNodes, map[string]any{
		"uuid": nodeUUIDs,
	}, "id", "uuid"); err != nil {
		return nil, err
	}

	dbNodeMap := utils.Slice2Map(dbNodes, func(n *model.MaterialNode) (uuid.UUID, int64) {
		return n.UUID, n.ID
	})

	resMap := utils.Slice2Map(resourceNodes,
		func(item *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
			return item.Name, item
		})
	nodeNames := utils.FilterSlice(
		req.Nodes,
		func(item *material.Material) (
			*utils.Node[uuid.UUID, *material.Material], bool,
		) {
			return &utils.Node[uuid.UUID, *material.Material]{
				Name:   item.UUID,
				Parent: item.ParentUUID,
				Data:   item,
			}, !item.UUID.IsNil()
		})

	// FIXME: 可能是多棵树，挂载在根节点
	levelNodes, err := utils.BuildHierarchy(nodeNames)
	if err != nil {
		return nil, code.InvalidDagErr.WithMsg(err.Error())
	}

	nodeMap := make(map[uuid.UUID]*model.MaterialNode)
	if err = db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range levelNodes {
			newDatas := make([]*model.MaterialNode, 0, len(nodes))
			updateDatas := make([]*model.MaterialNode, 0, len(nodes))
			for _, n := range nodes {
				data := &model.MaterialNode{
					ParentID:       0,
					LabID:          labUser.LabID,
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
					EdgeUUID:       n.UUID,
				}
				if node := nodeMap[n.ParentUUID]; node != nil {
					data.ParentID = node.ID
				}

				if resInfo := resMap[n.Class]; resInfo != nil {
					data.ResourceNodeID = resInfo.ID
					data.Icon = resInfo.Icon
					data.Model = resInfo.Model
				}

				if nodeID, ok := dbNodeMap[n.UUID]; ok {
					data.ID = nodeID
					updateDatas = append(updateDatas, data)
				} else {
					newDatas = append(newDatas, data)
				}

				nodeMap[n.UUID] = data
			}

			// 更新
			if len(updateDatas) > 0 {
				if _, err := m.materialStore.UpsertMaterialNode(txCtx, updateDatas, []string{"id"}, []string{"uuid", "id"}); err != nil {
					return err
				}
			}

			// 新建
			if len(newDatas) > 0 {
				if _, err := m.materialStore.UpsertMaterialNode(txCtx, newDatas, nil, []string{"uuid", "id"}); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return utils.MapToSlice(nodeMap, func(key uuid.UUID, n *model.MaterialNode) (*material.UpsertMaterialResp, bool) {
		return &material.UpsertMaterialResp{
			UUID:        key,
			CloudUUID:   n.UUID,
			Name:        n.Name,
			DisplayName: n.DisplayName,
		}, true
	}), nil
}

func (m *materialImpl) delUpsertMaterialNode(ctx context.Context, parentUUID uuid.UUID, parentID int64, req *material.UpsertMaterialReq) ([]*material.UpsertMaterialResp, error) {
	labUser := auth.GetLabUser(ctx)
	dbNodes, err := m.materialStore.GetDescendants(ctx, labUser.LabID, parentID)
	if err != nil {
		return nil, err
	}

	reqNodeMap := utils.Slice2Map(req.Nodes, func(n *material.Material) (uuid.UUID, *material.Material) {
		return n.UUID, n
	})

	delNodeUUIDs := make([]uuid.UUID, 0, len(req.Nodes))
	for _, childNode := range dbNodes {
		if _, ok := reqNodeMap[childNode.UUID]; ok {
			continue
		}

		delNodeUUIDs = append(delNodeUUIDs, childNode.UUID)
	}

	// 删除孩子节点
	if len(delNodeUUIDs) > 0 {
		if _, err := m.materialStore.DelNodes(ctx, delNodeUUIDs); err != nil {
			return nil, err
		}
	}

	resTplNames := make([]string, 0, len(req.Nodes))
	for _, data := range req.Nodes {
		if data.Class == "" {
			continue
		}
		resTplNames = utils.AppendUniqSlice(resTplNames, data.Class)
	}

	resourceNodes := make([]*model.ResourceNodeTemplate, 0, len(resTplNames))
	err = m.envStore.FindDatas(ctx, &resourceNodes, map[string]any{
		"lab_id": labUser.LabID,
		"name":   resTplNames,
	}, "id", "name", "icon", "model")
	if err != nil {
		return nil, err
	}

	// 强制校验资源模板是否存在
	if len(resourceNodes) != len(resTplNames) {
		return nil, code.ResNotExistErr
	}

	dbNodeMap := utils.Slice2Map(dbNodes, func(n *model.MaterialNode) (uuid.UUID, int64) {
		return n.UUID, n.ID
	})

	resMap := utils.Slice2Map(resourceNodes,
		func(item *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
			return item.Name, item
		})
	nodeNames := utils.FilterSlice(
		req.Nodes,
		func(item *material.Material) (
			*utils.Node[uuid.UUID, *material.Material], bool,
		) {
			return &utils.Node[uuid.UUID, *material.Material]{
				Name:   item.UUID,
				Parent: item.ParentUUID,
				Data:   item,
			}, !item.UUID.IsNil()
		})

	levelNodes, err := utils.BuildHierarchy(nodeNames)
	if err != nil {
		return nil, code.InvalidDagErr.WithMsg(err.Error())
	}

	nodeMap := make(map[uuid.UUID]*model.MaterialNode)
	nodeMap[parentUUID] = &model.MaterialNode{
		BaseModel: model.BaseModel{ID: parentID},
	}
	if err = db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range levelNodes {
			newDatas := make([]*model.MaterialNode, 0, len(nodes))
			updateDatas := make([]*model.MaterialNode, 0, len(nodes))
			for _, n := range nodes {
				data := &model.MaterialNode{
					ParentID:       0,
					LabID:          labUser.LabID,
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
					EdgeUUID:       n.UUID,
				}
				if node := nodeMap[n.ParentUUID]; node != nil {
					data.ParentID = node.ID
				}

				if resInfo := resMap[n.Class]; resInfo != nil {
					data.ResourceNodeID = resInfo.ID
					data.Icon = resInfo.Icon
					data.Model = resInfo.Model
				}

				if nodeID, ok := dbNodeMap[n.UUID]; ok {
					data.ID = nodeID
					updateDatas = append(updateDatas, data)
				} else {
					newDatas = append(newDatas, data)
				}

				nodeMap[n.UUID] = data
			}

			// 根据 id 更新数据, 返回 uuid
			if len(updateDatas) > 0 {
				// 根据 id 更新
				if _, err := m.materialStore.UpsertMaterialNode(txCtx, updateDatas, []string{"id"}, []string{"uuid", "id"}); err != nil {
					return err
				}
			}

			// 根据 uuid 更新，并且返回 uuid
			if len(newDatas) > 0 {
				if _, err := m.materialStore.UpsertMaterialNode(txCtx,
					newDatas, nil,
					[]string{"uuid", "id"}); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return utils.MapToSlice(nodeMap, func(key uuid.UUID, n *model.MaterialNode) (*material.UpsertMaterialResp, bool) {
		return &material.UpsertMaterialResp{
			UUID:        key,
			CloudUUID:   n.UUID,
			Name:        n.Name,
			DisplayName: n.DisplayName,
		}, key != parentUUID
	}), nil
}

func (m *materialImpl) EdgeCreateEdge(ctx context.Context, req *material.CreateMaterialEdgeReq) error {
	labUser := auth.GetLabUser(ctx)
	if labUser == nil {
		return code.UnLogin
	}

	edges := req.Edges

	nodeUUIDs := make([]uuid.UUID, 0, 2*len(edges))
	for _, e := range edges {
		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, e.SourceUUID)
		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, e.TargetUUID)
	}

	edgeInfo, err := m.materialStore.GetNodeHandlesByUUIDV1(ctx, nodeUUIDs)
	if err != nil {
		return err
	}
	edgeDatas := make([]*model.MaterialEdge, 0, len(edges))
	for _, edge := range edges {
		sourceNode, ok := edgeInfo[edge.SourceUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges source not exist source node uuid: %s", edge.SourceUUID)
			return code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("source node uuid: %s", edge.SourceUUID))
		}

		sourceHandle, ok := sourceNode[edge.SourceHandle]
		if !ok {
			logger.Errorf(ctx, "addWSEdges source handle not exist source uuid: %s",
				edge.SourceHandle)
			return code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("source handle uuid: %s",
				edge.SourceHandle))
		}

		targetNode, ok := edgeInfo[edge.TargetUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges target not exist target node uuid: %s", edge.TargetUUID)
			return code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("target node uuid: %s", edge.TargetUUID))
		}

		targetHandle, ok := targetNode[edge.TargetHandle]
		if !ok {
			logger.Errorf(ctx, "addWSEdges target handle not exist uuid: %s", edge.TargetHandle)
			return code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("target handle uuid: %s",
				edge.TargetHandle))
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

	// res := utils.FilterSlice(edgeDatas, func(data *model.MaterialEdge) (material.WSEdge, bool) {
	// 	return material.WSEdge{
	// 		UUID:             data.UUID,
	// 		SourceNodeUUID:   data.SourceNodeUUID,
	// 		TargetNodeUUID:   data.TargetNodeUUID,
	// 		SourceHandleUUID: data.SourceHandleUUID,
	// 		TargetHandleUUID: data.TargetHandleUUID,
	// 		Type:             "step",
	// 	}, true
	// })
	return nil
}

func (m *materialImpl) EdgeQueryMaterial(ctx context.Context, req *material.MaterialQueryReq) (*material.MaterialQueryResp, error) {
	labUser := auth.GetLabUser(ctx)
	if labUser == nil {
		return nil, code.UnLogin
	}

	if len(req.UUIDS) == 0 {
		return nil, nil
	}

	nodes := make([]*model.MaterialNode, 0, len(req.UUIDS))
	if err := m.materialStore.FindDatas(ctx, &nodes, map[string]any{
		"uuid": req.UUIDS,
	}); err != nil {
		return nil, err
	}

	parentIDs := utils.FilterUniqSlice(nodes, func(n *model.MaterialNode) (int64, bool) {
		return n.ParentID, n.ParentID > 0
	})

	parentNodes := make([]*model.MaterialNode, 0, len(parentIDs))
	if err := m.materialStore.FindDatas(ctx, &parentNodes, map[string]any{
		"id": parentIDs,
	}, "id", "uuid"); err != nil {
		return nil, err
	}

	parentNodeMap := utils.Slice2Map(parentNodes, func(n *model.MaterialNode) (int64, uuid.UUID) {
		return n.ID, n.UUID
	})

	return &material.MaterialQueryResp{
		Nodes: utils.FilterSlice(nodes, func(n *model.MaterialNode) (*material.EdgeNode, bool) {
			return &material.EdgeNode{
				UUID:        n.UUID,
				ParentUUID:  parentNodeMap[n.ParentID],
				Name:        n.Name,
				Class:       n.Class,
				DisplayName: n.DisplayName,
				Description: n.Description,
				Status:      n.Status,
				Type:        n.Type,
				Config:      n.InitParamData,
				Schema:      n.Schema,
				Data:        n.Data,
				Pose:        n.Pose.Data(),
				Model:       n.Model,
				Icon:        n.Icon,
				Extra:       n.Extra,
			}, true
		}),
	}, nil
}
