package material

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
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
)

type materialImpl struct {
	envStore      repo.EnvRepo
	materialStore repo.MaterialRepo
	wsClient      *melody.Melody
	msgCenter     notify.MsgCenter
}

func NewMaterial(wsClient *melody.Melody) material.Service {
	m := &materialImpl{
		envStore:      eStore.NewEnv(),
		materialStore: mStore.NewMaterialImpl(),
		wsClient:      wsClient,
		msgCenter:     events.NewEvents(),
	}
	events.NewEvents().Registry(context.Background(), notify.MaterialModify, m.HandleNotify)

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

func (m *materialImpl) createNodes(ctx context.Context, labData *model.Laboratory, req *material.GraphNodeReq) error {
	resTplNames := make([]string, 0, len(req.Nodes))
	for _, data := range req.Nodes {
		if data.Type == model.MATERIALDEVICE ||
			data.Type == model.MATERIALCONTAINER {
			resTplNames = utils.AppendUniqSlice(resTplNames, data.Class)
		}
	}

	resMap, err := m.envStore.GetResourceTemplate(ctx, labData.ID, resTplNames)
	if err != nil {
		return err
	}

	// 强制校验资源模板是否存在
	if len(resMap) != len(resTplNames) {
		return code.ResNotExistErr
	}

	levelNodes := sortNodeLevel(ctx, req.Nodes)
	nodeMap := make(map[string]*model.MaterialNode)
	if err = db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range levelNodes {
			datas := make([]*model.MaterialNode, 0, len(nodes))
			for _, n := range nodes {
				data := &model.MaterialNode{
					ParentID:               0,
					LabID:                  labData.ID,
					Name:                   n.DeviceID,
					DisplayName:            n.Name,
					Description:            n.Description,
					Type:                   n.Type,
					ResourceNodeTemplateID: 0,
					InitParamData:          n.Config,
					Data:                   n.Data,
					Pose:                   n.Pose,
					Model:                  n.Model,
					Icon:                   "",
					Schema:                 n.Schema,
				}
				if node := nodeMap[n.Parent]; node != nil {
					data.ParentID = node.ID
				}

				if resInfo := resMap[n.Class]; resInfo != nil {
					data.ResourceNodeTemplateID = resInfo.ID
					data.Icon = resInfo.Icon // TODO: 是否有缺省值
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

func sortNodeLevel(ctx context.Context, nodes []*material.Node) [][]*material.Node {
	nodeMap := make(map[string]*material.Node)
	for _, node := range nodes {
		nodeMap[node.DeviceID] = node
	}

	mapLevel := make(map[string]int)
	for _, node := range nodes {
		getNodeLevel(ctx, mapLevel, nodeMap, node)
	}

	type IndexLevel struct {
		Level int
		Nodes []*material.Node
	}

	levelNodeMap := make(map[int][]*material.Node)
	for name, level := range mapLevel {
		levelNodeMap[level] = append(levelNodeMap[level], nodeMap[name])
	}

	indexLevel := make([]*IndexLevel, 0, len(mapLevel))
	for level, nodes := range levelNodeMap {
		indexLevel = append(indexLevel, &IndexLevel{
			Level: level,
			Nodes: nodes,
		})
	}

	sort.Slice(indexLevel, func(i, j int) bool {
		return indexLevel[i].Level < indexLevel[j].Level
	})

	res := make([][]*material.Node, 0, len(indexLevel))
	for _, groupNodes := range indexLevel {
		res = append(res, groupNodes.Nodes)
	}

	return res
}

func getNodeLevel(ctx context.Context, cache map[string]int, nodeMap map[string]*material.Node, node *material.Node) int {
	if node.Parent == "" {
		cache[node.DeviceID] = 0
		return 0
	}

	cacheNodeLevel, ok := cache[node.DeviceID]
	if ok {
		return cacheNodeLevel
	}

	parentNodeLevel, ok := cache[node.Parent]
	if ok {
		cache[node.DeviceID] = parentNodeLevel + 1
		return 0
	}

	parentNode, ok := nodeMap[node.Parent]
	if !ok {
		logger.Warnf(ctx, "node parent invalidate node name: %s, node parent name: %s", node.Name, node.Parent)
		cache[node.DeviceID] = 0
		return 0
	}

	parentLevel := getNodeLevel(ctx, cache, nodeMap, parentNode)
	cache[node.DeviceID] = parentLevel + 1
	return cache[node.DeviceID]
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

func (m *materialImpl) HandleWSMsg(ctx context.Context, s *melody.Session, b []byte) error {
	msgType := &material.WsMsgType{}
	err := json.Unmarshal(b, msgType)
	if err != nil {
		return err
	}

	switch msgType.Action {
	case material.FetchGrpah: // 首次获取组态图
		return m.fetchGraph(ctx, s, msgType.MsgUUID)
	case material.FetchTemplate: // 首次获取模板
		return m.fetchDeviceTemplate(ctx, s, msgType.MsgUUID)
	case material.BatchCreateNode: // TODO: 这个不实现，一次修改数量太多，没必要，通知也复杂
		return m.batchCreateNodes(ctx, s, b)
	case material.BatchUpdateNode: // 批量更新节点
		return m.batchUpateNode(ctx, s, b)
	case material.BatchDelNode: // 批量删除节点
		return m.batchDelNode(ctx, s, b)
	case material.BatchCreateEdge: // 批量创建边
		return m.batchCreateEdge(ctx, s, b)
	case material.BatchDelEdge: // 批量删除边
		return m.batchDelEdge(ctx, s, b)
	default:
		return common.ReplyWSErr(s, code.UnknownWSActionErr.WithMsg(string(msgType.Action)))
	}
}

// 获取组态图
func (m *materialImpl) fetchGraph(ctx context.Context, s *melody.Session, msgUUID uuid.UUID) error {
	// 获取所有组态图信息
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return err
	}
	nodes, err := m.materialStore.GetNodesByLabID(ctx, labData.ID)
	if err != nil {
		common.ReplyWSErr(s, err)
		return err
	}

	resTplIDS := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (int64, bool) {
		if nodeItem.ResourceNodeTemplateID == 0 {
			return 0, false
		}
		return nodeItem.ResourceNodeTemplateID, true
	})
	resTplIDS = utils.RemoveDuplicates(resTplIDS)

	nodesMap := utils.SliceToMap(nodes, func(item *model.MaterialNode) (int64, *model.MaterialNode) {
		return item.ID, item
	})

	resHandlesMap, err := m.envStore.GetResourceHandleTemplates(ctx, resTplIDS)
	if err != nil {
		common.ReplyWSErr(s, err)
		return err
	}

	nodeUUIDs := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (uuid.UUID, bool) {
		return nodeItem.UUID, true
	})
	edges, err := m.materialStore.GetEdgesByNodeUUID(ctx, nodeUUIDs)
	if err != nil {
		common.ReplyWSErr(s, err)
		return err
	}
	resNodeTplUUIDMap, err := m.envStore.GetResourceNodeTemplateUUID(ctx, resTplIDS)
	if err != nil {
		common.ReplyWSErr(s, err)
		return err
	}

	respNodes := utils.FilterSlice(nodes, func(nodeItem *model.MaterialNode) (*material.WSNode, bool) {
		var parentUUID uuid.UUID
		parentNode, ok := nodesMap[nodeItem.ParentID]
		if ok {
			parentUUID = parentNode.UUID
		}
		resNodeTplUUID, _ := resNodeTplUUIDMap[nodeItem.ResourceNodeTemplateID]

		handles, _ := resHandlesMap[nodeItem.ID]
		return &material.WSNode{
			UUID:                nodeItem.UUID,
			ParentUUID:          parentUUID,
			Name:                nodeItem.Name,
			DisplayName:         nodeItem.DisplayName,
			Description:         nodeItem.Description,
			Type:                nodeItem.Type,
			ResNodeTemplateUUID: resNodeTplUUID,
			InitParamData:       nodeItem.InitParamData,
			Schema:              nodeItem.Schema,
			Data:                nodeItem.Data,
			Pose:                nodeItem.Pose,
			Model:               nodeItem.Model,
			Icon:                nodeItem.Icon,
			Handles: utils.FilterSlice(handles, func(handleItem *model.ResourceHandleTemplate) (*material.WSHandle, bool) {
				var nodeUUID uuid.UUID
				if nodeData, ok := nodesMap[handleItem.NodeID]; ok {
					nodeUUID = nodeData.UUID
				}

				return &material.WSHandle{
					UUID:        handleItem.UUID,
					NodeUUID:    nodeUUID,
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
			SourceNodeUUID:   item.SourceNodeUUID,
			TargetNodeUUID:   item.TargetNodeUUID,
			SourceHandleUUID: item.SourceHandleUUID,
			TargetHandleUUID: item.SourceHandleUUID,
		}, true
	})

	resp := &material.GraphResp{
		Nodes: respNodes,
		Edges: respEdges,
	}

	return common.ReplyWSOk(s, &material.WSData[*material.GraphResp]{
		WsMsgType: material.WsMsgType{
			Action:  material.FetchGrpah,
			MsgUUID: msgUUID,
		},
		Data: resp,
	})
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

// 获取设备模板
func (m *materialImpl) fetchDeviceTemplate(ctx context.Context, s *melody.Session, msgUUID uuid.UUID) error {
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return err
	}

	tplNodes, err := m.envStore.GetAllDeviceTemplateByLabID(ctx, labData.ID)
	if err != nil {
		return err
	}
	nodeIDs := utils.FilterSlice(tplNodes, func(item *model.ResourceNodeTemplate) (int64, bool) {
		return item.ID, true
	})

	tplHandles, err := m.envStore.GetAllDeviceTemplateHandlesByID(ctx, nodeIDs)
	if err != nil {
		return err
	}

	tplDatas := utils.FilterSlice(tplNodes, func(nodeItem *model.ResourceNodeTemplate) (*material.DeviceTemplate, bool) {
		return &material.DeviceTemplate{
			Handles: utils.FilterSlice(tplHandles, func(handleItem *model.ResourceHandleTemplate) (*material.DeviceHandleTemplate, bool) {
				// FIXME: 此处效率可以优化
				if handleItem.NodeID != nodeItem.ID {
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
			UUID:         nodeItem.UUID,
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
			Labels:       nodeItem.Labels,
			DataSchema:   nodeItem.DataSchema,
			ConfigSchema: nodeItem.ConfigSchema,
		}, true
	})
	resData := &material.DeviceTemplates{
		Templates: tplDatas,
	}

	return common.ReplyWSOk(s, &material.WSData[*material.DeviceTemplates]{
		WsMsgType: material.WsMsgType{
			Action:  material.FetchTemplate,
			MsgUUID: msgUUID,
		},
		Data: resData,
	})
}

// 批量创建节点
func (m *materialImpl) batchCreateNodes(_ context.Context, _ *melody.Session, _ []byte) error {
	return nil
}

// 批量更新节点
func (m *materialImpl) batchUpateNode(_ context.Context, _ *melody.Session, _ []byte) error {
	// node := model.MaterialNode{
	// ParentID         :
	// DisplayName :
	// Description :
	// InitParamData    :
	// Data             :
	// Pose             :
	// Icon             :
	// }

	return nil
}

// 批量删除节点
func (m *materialImpl) batchDelNode(ctx context.Context, s *melody.Session, b []byte) error {
	data := &material.WSData[[]uuid.UUID]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelNode unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	res, err := m.materialStore.DelNodes(ctx, data.Data)
	if err != nil {
		common.ReplyWSErr(s, err)
		return err
	}

	resData := &material.WSData[*repo.DelNodeInfo]{
		WsMsgType: material.WsMsgType{Action: data.Action, MsgUUID: data.MsgUUID},
		Data:      res,
	}

	if err := common.ReplyWSOk(s, resData); err != nil {
		logger.Errorf(ctx, "batchDelNode reply ws ok fail err: %+v", err)
	}
	// 广播
	labUUID, _ := s.Get("lab_uuid")
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		logger.Warnf(ctx, "batchDelNode broadcast can not get user info")
		return nil
	}
	m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Action:  notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labUUID.(uuid.UUID),
		Data:    resData,
	})

	return nil
}

// 批量创建 edge
func (m *materialImpl) batchCreateEdge(ctx context.Context, s *melody.Session, b []byte) error {
	userInfo := auth.GetCurrentUser(ctx)
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return err
	}
	data := &material.WSData[[]material.WSEdge]{}
	err = json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelEdge unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}
	if err := m.addWSEdges(ctx, data.Data); err != nil {
		common.ReplyWSErr(s, err)
		return err
	}

	wsData := &material.WSData[[]material.WSEdge]{
		WsMsgType: material.WsMsgType{
			Action:  data.Action,
			MsgUUID: data.MsgUUID,
		},
		Data: data.Data,
	}

	if err = common.ReplyWSOk(s, wsData); err != nil {
		logger.Errorf(ctx, "batchCreateEdge fail err: %+v", err)
	}

	return m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Action:  notify.MaterialModify,
		UserID:  userInfo.ID,
		LabUUID: labData.UUID,
		Data:    wsData,
	})
}

func (m *materialImpl) batchDelEdge(ctx context.Context, s *melody.Session, b []byte) error {
	data := &material.WSData[[]uuid.UUID]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelEdge unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	if err := m.materialStore.DelEdges(ctx, data.Data); err != nil {
		common.ReplyWSErr(s, err)
		return err
	}
	resData := &material.WSData[[]uuid.UUID]{
		WsMsgType: material.WsMsgType{
			Action:  data.Action,
			MsgUUID: data.MsgUUID,
		},
		Data: data.Data,
	}

	if err = common.ReplyWSOk(s, resData); err != nil {
		logger.Errorf(ctx, "batchDelEdge reply ws ok fail err: %+v", err)
	}

	userInfo := auth.GetCurrentUser(ctx)
	labData, err := m.getLab(ctx, s)
	if err != nil {
		return err
	}

	return m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Action:  notify.MaterialModify,
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

func (m *materialImpl) HandleWSConnect(ctx context.Context, s *melody.Session) error {
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

func (m *materialImpl) addWSEdges(ctx context.Context, edges []material.WSEdge) error {
	nodeUUIDs := make([]uuid.UUID, 0, 2*len(edges))
	for _, e := range edges {
		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, e.SourceNodeUUID)
		nodeUUIDs = utils.AppendUniqSlice(nodeUUIDs, e.TargetNodeUUID)
	}

	edgeInfo, err := m.materialStore.GetNodeHandlesByUUID(ctx, nodeUUIDs)
	if err != nil {
		return err
	}
	edgeDatas := make([]*model.MaterialEdge, 0, len(edges))
	for _, edge := range edges {
		sourceNode, ok := edgeInfo[edge.SourceNodeUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges source not exist source node uuid: %s", edge.SourceNodeUUID)
			return code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("source node uuid: %s", edge.SourceNodeUUID))
		}

		sourceHandle, ok := sourceNode[edge.SourceHandleUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges source handle not exist source uuid: %s",
				edge.SourceHandleUUID)
			return code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("source handle uuid: %s",
				edge.SourceHandleUUID))
		}

		targetNode, ok := edgeInfo[edge.TargetNodeUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges target not exist target node uuid: %s", edge.TargetNodeUUID)
			return code.EdgeNodeNotExistErr.WithMsg(fmt.Sprintf("target node uuid: %s", edge.TargetNodeUUID))
		}

		targetHandle, ok := targetNode[edge.TargetHandleUUID]
		if !ok {
			logger.Errorf(ctx, "addWSEdges target handle not exist uuid: %s", edge.TargetHandleUUID)
			return code.EdgeHandleNotExistErr.WithMsg(fmt.Sprintf("target handle uuid: %s",
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
		return err
	}

	return nil
}
