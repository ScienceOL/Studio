package material

import (
	"context"
	"sort"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/material"
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
	envStore      repo.EnvRepo
	materialStore repo.MaterialRepo
}

func NewMaterial() material.Service {
	return &materialImpl{
		envStore:      eStore.NewEnv(),
		materialStore: mStore.NewMaterialImpl(),
	}
}

func (m *materialImpl) CreateMaterial(ctx context.Context, req []*material.Node) error {
	uuid := common.BinUUID(datatypes.BinUUIDFromString(""))
	labData, err := m.envStore.GetLabByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	regNames := make([]string, 0, len(req))
	for _, data := range req {
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
		return code.REGNOTEXISTErr
	}

	levelNodes := sortNodeLevel(ctx, req)
	nodeMap := make(map[string]*model.MaterialNode)
	return db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, nodes := range levelNodes {
			datas := make([]*model.MaterialNode, 0, len(nodes))
			deviceTemplateIDs := make([]int64, 0, len(nodes))
			handleNodes := make(map[int64]*model.MaterialNode)
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
				if regInfo := regMap[n.DeviceID]; regInfo != nil {
					deviceTemplateIDs = utils.AppendUniqSlice(deviceTemplateIDs, regInfo.DeviceNodeTemplateID)
					data.RegID = regInfo.RegID
					data.DeviceNodeTemplateID = regInfo.DeviceNodeTemplateID
					handleNodes[regInfo.DeviceNodeTemplateID] = data
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
				materialNode, ok := handleNodes[templateNodeID]
				if !ok {
					continue
				}
				for _, h := range templateHandles {
					handleData := &model.MaterialHandle{
						NodeID:      materialNode.ID,
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
			if err := m.materialStore.UpsertMaterialHandle(txCtx, materialHandles); err != nil {
				return err
			}
		}
		return nil
	})
}

func sortNodeLevel(ctx context.Context, nodes []*material.Node) [][]*material.Node {
	nodeMap := make(map[string]*material.Node)
	for _, node := range nodes {
		nodeMap[node.Name] = node
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
		cache[node.Name] = 0
		return 0
	}

	cacheNodeLevel, ok := cache[node.Name]
	if ok {
		return cacheNodeLevel
	}

	parentNodeLevel, ok := cache[node.Parent]
	if ok {
		cache[node.Name] = parentNodeLevel + 1
		return 0
	}

	parentNode, ok := nodeMap[node.Parent]
	if !ok {
		logger.Warnf(ctx, "node parent invalidate node name: %s, node parent name: %s", node.Name, node.Parent)
		cache[node.Name] = 0
		return 0
	}

	parentLevel := getNodeLevel(ctx, cache, nodeMap, parentNode)
	cache[node.Name] = parentLevel + 1
	return cache[node.Name]
}
