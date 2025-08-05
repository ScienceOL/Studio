package material

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

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

func (m *materialImpl) CreateMaterial(ctx context.Context, req *material.GraphNode) error {
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

	regNames := make([]string, 0, len(req.Nodes))
	for _, data := range req.Nodes {
		if data.Type == model.MATERIALDEVICE ||
			data.Type == model.MATERIALCONTAINER {
			regNames = utils.AppendUniqSlice(regNames, data.Class)
		}
	}

	regMap, err := m.envStore.GetRegs(ctx, labData.ID, regNames)
	if err != nil {
		return err
	}

	// 强制校验注册表是否存在
	if len(regMap) != len(regNames) {
		return code.RegNotExistErr
	}

	levelNodes := sortNodeLevel(ctx, req.Nodes)
	nodeMap := make(map[string]*model.MaterialNode)
	err = db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range levelNodes {
			datas := make([]*model.MaterialNode, 0, len(nodes))
			deviceTemplateIDs := make([]int64, 0, len(nodes))
			handleNodes := make(map[int64][]*model.MaterialNode)
			for _, n := range nodes {
				data := &model.MaterialNode{
					ParentID:             0,
					LabID:                labData.ID,
					Name:                 n.DeviceID,
					DisplayName:          n.Name,
					Description:          n.Description,
					Type:                 n.Type,
					DeviceNodeTemplateID: 0,
					RegID:                0,
					InitParamData:        n.Config,
					// Schema              :
					Data: n.Data,
					// Dirs:
					Position: n.Position,
					// Pose                :
					Model: n.Model,
				}
				if node := nodeMap[n.Parent]; node != nil {
					data.ParentID = node.ID
				}
				if regInfo := regMap[n.Class]; regInfo != nil {
					if n.Class == "virtual_transfer_pump" {
						fmt.Println(n.Class)
					}
					deviceTemplateIDs = utils.AppendUniqSlice(deviceTemplateIDs, regInfo.DeviceNodeTemplateID)
					data.RegID = regInfo.RegID
					data.DeviceNodeTemplateID = regInfo.DeviceNodeTemplateID
					handleNodes[regInfo.DeviceNodeTemplateID] = append(handleNodes[regInfo.DeviceNodeTemplateID], data)
				}

				datas = append(datas, data)
				nodeMap[n.DeviceID] = data
			}

			if err := m.materialStore.UpsertMaterialNode(txCtx, datas); err != nil {
				return err
			}

			deviceTemplateHandles, err := m.envStore.GetDeviceTemplateHandels(txCtx, deviceTemplateIDs)
			if err != nil {
				return err
			}
			materialHandles := make([]*model.MaterialHandle, 0, 10)
			for templateNodeID, templateHandles := range deviceTemplateHandles {
				materialNodes, ok := handleNodes[templateNodeID]
				if !ok {
					continue
				}
				for _, node := range materialNodes {
					for _, h := range templateHandles {
						handleData := &model.MaterialHandle{
							NodeID:      node.ID,
							Name:        h.Name,
							DisplayName: utils.Or(h.DisplayName, h.Key),
							Type:        h.Type,
							IOType:      h.IOType,
							Source:      h.Source,
							Key:         h.Key,
							Side:        utils.Or(h.Side, "WEST"),
							Connected:   false,
							Required:    false,
						}
						materialHandles = append(materialHandles, handleData)
					}
				}
			}
			if err := m.materialStore.UpsertMaterialHandle(txCtx, materialHandles); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	_ = m.addEdges(ctx, labData.ID, req.Edges, false)
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
	case material.FetchNodes:
		return m.fetchGraph(ctx, s)
	case material.BatchCreateNode:
		return m.batchCreateNodes(ctx, s, b)
	case material.BatchUpdateNode:
		return m.batchUpateNode(ctx, s, b)
	case material.BatchDelNode:
		return m.batchDelNode(ctx, s, b)
	case material.BatchCreateEdge:
		return m.batchCreateEdge(ctx, s, b)
	case material.BatchDelEdge:
		return m.batchDelEdge(ctx, s, b)
	default:
		return common.ReplyWSErr(s, code.UnknowWSActionErr.WithMsg(string(msgType.Action)))
	}
}

func (m *materialImpl) fetchGraph(ctx context.Context, s *melody.Session) error {
	// TODO: 获取组态图
	return nil
}

func (m *materialImpl) batchCreateNodes(ctx context.Context, s *melody.Session, b []byte) error {
	data := &material.WSData[material.WSNodes]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchCreateNodes unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}
	// TODO: 补充代码逻辑

	return nil
}

func (m *materialImpl) batchUpateNode(ctx context.Context, s *melody.Session, b []byte) error {

	return nil
}

func (m *materialImpl) batchDelNode(ctx context.Context, s *melody.Session, b []byte) error {
	data := &material.WSData[material.WSDelNodes]{}
	err := json.Unmarshal(b, data)
	if err != nil {
		logger.Errorf(ctx, "batchDelNode unmarshal data err: %+v", err)
		return code.UnmarshalWSDataErr.WithMsg(err.Error())
	}

	res, err := m.materialStore.DelNodes(ctx, data.Data.NodeUUIDs)
	if err != nil {
		common.ReplyWSErr(s, err)
		return err
	}

	resData := &material.WSData[*repo.DelNodeInfo]{
		WsMsgType: material.WsMsgType{Action: data.Action},
		UUID:      data.UUID,
		Data:      res,
	}

	common.ReplyWSOk(s, data)
	// 广播
	m.msgCenter.Broadcast(ctx, &notify.SendMsg{
		Action: notify.MaterialModify,
		Data:   resData,
	})

	return nil
}

func (m *materialImpl) batchCreateEdge(ctx context.Context, s *melody.Session, b []byte) error {

	return nil
}

func (m *materialImpl) batchUpateEdge(ctx context.Context, s *melody.Session, b []byte) error {

	return nil
}

func (m *materialImpl) batchDelEdge(ctx context.Context, s *melody.Session, b []byte) error {

	return nil
}

// 接受到 redis 广播通知消息
func (m *materialImpl) HandleNotify(ctx context.Context, msg string) error {
	return m.wsClient.BroadcastFilter([]byte(msg), func(s *melody.Session) bool {
		return true
		sessionValue, ok := s.Get("lab id")
		if !ok {
			return false
		}
		return utils.Compare(sessionValue, "lab id")
	})
}
