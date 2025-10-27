package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/datatypes"
)

type Label struct {
	Label string    `json:"label"`
	Value string    `json:"value"`
	UUID  uuid.UUID `json:"uuid"`
}

type schemaHelper struct {
	materialStore repo.MaterialRepo
}

func (s *schemaHelper) handleSchema(ctx context.Context, materialNodes []*model.MaterialNode, schema datatypes.JSON) (datatypes.JSON, map[string]any) {
	dataSchema := map[string]any{}
	if err := json.Unmarshal(schema, &dataSchema); err != nil {
		logger.Errorf(ctx, "handleSchema Unmarshal schema fail schmea: %s", schema)
		return schema, nil
	}

	ok, placeholder, innerProperties := getPlaceholder(dataSchema)
	if !ok {
		return schema, nil
	}

	allDefault := map[string]any{}
	allUISchema := map[string]any{}
	for fieldName, placeholderType := range placeholder {
		var keyDefault map[string]any
		var uiSchema map[string]any
		switch placeholderType {
		case "unilabos_devices":
			keyDefault, uiSchema = s.handleDevices(ctx, materialNodes, innerProperties, fieldName)
		case "unilabos_resources":
			keyDefault, uiSchema = s.handleResources(ctx, materialNodes, innerProperties, fieldName)
		case "unilabos_nodes":
			keyDefault, uiSchema = s.handleNodes(ctx, materialNodes, innerProperties, fieldName)
		case "unilabos_transfer_liquid":
			keyDefault, uiSchema = s.handleTransferLiquid(ctx, materialNodes, innerProperties, fieldName)
		default:
			logger.Errorf(ctx, "unknown placeholder type: %s", placeholderType)
		}

		maps.Copy(allDefault, keyDefault)
		if len(uiSchema) > 0 {
			maps.Copy(allUISchema, map[string]any{fieldName: uiSchema})
		}
	}

	if len(allUISchema) > 0 {
		dataSchema["uiSchema"] = allUISchema
	}

	b, err := json.Marshal(dataSchema)
	if err != nil {
		logger.Errorf(ctx, "handleSchema Marshal err: %+v", err)
		return schema, nil
	}

	return datatypes.JSON(b), allDefault
}

func getPlaceholder(dataSchema map[string]any) (bool, map[string]string, map[string]any) {
	propertiesI, ok := dataSchema["properties"]
	if !ok {
		return false, nil, nil
	}

	properties, ok := propertiesI.(map[string]any)
	if !ok {
		return false, nil, nil
	}

	goalI, ok := properties["goal"]
	if !ok {
		return false, nil, nil
	}

	goal, ok := goalI.(map[string]any)
	if !ok {
		return false, nil, nil
	}

	placeholderI, ok := goal["_unilabos_placeholder_info"]
	if !ok {
		return false, nil, nil
	}

	placehodler, ok := placeholderI.(map[string]any)
	if !ok {
		return false, nil, nil
	}

	placehodlerMap := make(map[string]string)
	for key, value := range placehodler {
		valueStr, ok := value.(string)
		if !ok {
			return false, nil, nil
		}

		placehodlerMap[key] = valueStr
	}

	innerPropertiesI, ok := goal["properties"]
	if !ok {
		return false, nil, nil
	}

	innerProperties, ok := innerPropertiesI.(map[string]any)
	if !ok {
		return false, nil, nil
	}

	return true, placehodlerMap, innerProperties
}

func (s *schemaHelper) handleTransferLiquid(ctx context.Context, materialNodes []*model.MaterialNode, dataSchema map[string]any, key string) (map[string]any, map[string]any) {
	// plate
	// 不处理  tip_rack
	nodeMap := utils.Slice2Map(materialNodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
		return node.ID, node
	})

	targetType := []model.DEVICETYPE{
		model.MATERIALPLATE,
		model.MATERIALTIPRACK,
	}

	labels := utils.MapToSlice(nodeMap, func(id int64, node *model.MaterialNode) (*Label, bool) {
		if !slices.Contains(targetType, node.Type) {
			return nil, false
		}

		dirName := utils.Or(s.getDir(ctx, node, nodeMap, 100), node.Name)

		return &Label{
			Label: fmt.Sprintf("/%s (%s)", dirName, node.DisplayName),
			Value: "/" + dirName,
			UUID:  node.UUID,
		}, true
	})

	uiSchema := s.updateFieldWithOptions(dataSchema, key, labels, "请选择资源", "选择资源实例")
	if len(labels) == 1 {
		return map[string]any{key: labels[0].Value}, uiSchema
	}

	return nil, uiSchema
}

func (s *schemaHelper) handleNodes(ctx context.Context, materialNodes []*model.MaterialNode, dataSchema map[string]any, key string) (map[string]any, map[string]any) {
	nodeMap := utils.Slice2Map(materialNodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
		return node.ID, node
	})

	labels := utils.MapToSlice(nodeMap, func(id int64, node *model.MaterialNode) (*Label, bool) {
		dirName := utils.Or(s.getDir(ctx, node, nodeMap, 100), node.Name)
		if node.Type == model.MATERIALDEVICE {
			return &Label{
				Label: fmt.Sprintf("%s (%s) -%s", node.Name, node.DisplayName, node.Type),
				Value: "/" + dirName,
				UUID:  node.UUID,
			}, true
		} else {
			return &Label{
				Label: fmt.Sprintf("/%s (%s) -%s", dirName, node.DisplayName, node.Type),
				Value: "/" + dirName,
				UUID:  node.UUID,
			}, true
		}
	})

	uiSchema := s.updateFieldWithOptions(dataSchema, key, labels, "请选择节点", "选择节点实例")
	if len(labels) == 1 {
		return map[string]any{key: labels[0].Value}, uiSchema
	}
	return nil, uiSchema
}

func (s *schemaHelper) getDir(ctx context.Context, node *model.MaterialNode, nodes map[int64]*model.MaterialNode, maxDeep int) string {
	current := node
	res := make([]string, 0, 10)
	visited := make(map[int64]bool) // 防止循环引用
	for current != nil && maxDeep > 0 {
		res = append(res, current.Name)
		if _, ok := visited[current.ID]; ok {
			logger.Errorf(ctx, "getDir has circular id: %d", current.ID)
			return ""
		}
		visited[current.ID] = true
		current, _ = nodes[current.ParentID]
		maxDeep--
	}

	slices.Reverse(res)
	return strings.Join(res, "/")
}

func (s *schemaHelper) handleDevices(ctx context.Context, materialNodes []*model.MaterialNode, dataSchema map[string]any, key string) (map[string]any, map[string]any) {
	nodeMap := utils.Slice2Map(materialNodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
		return node.ID, node
	})

	labels := utils.FilterSlice(materialNodes, func(node *model.MaterialNode) (*Label, bool) {
		if node.Type != model.MATERIALDEVICE {
			return nil, false
		}

		// return &Label{
		// 	Label: fmt.Sprintf("%s (%s)", node.Name, node.DisplayName),
		// 	Value: node.Name,
		// 	UUID:  node.UUID,
		// }, true

		dirName := utils.Or(s.getDir(ctx, node, nodeMap, 100), node.Name)

		return &Label{
			Label: fmt.Sprintf("/%s (%s)", dirName, node.DisplayName),
			Value: "/" + dirName,
			UUID:  node.UUID,
		}, true
	})

	uiSchema := s.updateFieldWithOptions(dataSchema, key, labels, "请选择设备", "选择设备实例")
	if len(labels) == 1 {
		return map[string]any{key: labels[0].Value}, uiSchema
	}
	return nil, uiSchema
}

func (s *schemaHelper) handleResources(ctx context.Context, materialNodes []*model.MaterialNode, dataSchema map[string]any, key string) (map[string]any, map[string]any) {
	nodeMap := utils.Slice2Map(materialNodes, func(node *model.MaterialNode) (int64, *model.MaterialNode) {
		return node.ID, node
	})

	targetType := []model.DEVICETYPE{
		model.MATERIALREPO,
		model.MATERIALPLATE,
		model.MATERIALCONTAINER,
		model.MATERIALRESOURCE,
		model.MATERIALWELL,
		model.MATERIALTIP,
		model.MATERIALDECK,
		model.MATERIALTIPRACK,
		model.MATERIALTIPSPOT,
	}

	labels := utils.MapToSlice(nodeMap, func(id int64, node *model.MaterialNode) (*Label, bool) {
		if !slices.Contains(targetType, node.Type) {
			return nil, false
		}

		dirName := utils.Or(s.getDir(ctx, node, nodeMap, 100), node.Name)

		return &Label{
			Label: fmt.Sprintf("/%s (%s)", dirName, node.DisplayName),
			Value: "/" + dirName,
			UUID:  node.UUID,
		}, true
	})

	uiSchema := s.updateFieldWithOptions(dataSchema, key, labels, "请选择资源", "选择资源实例")
	if len(labels) == 1 {
		return map[string]any{key: labels[0].Value}, uiSchema
	}
	return nil, uiSchema
}

func (s *schemaHelper) updateFieldWithOptions(properties map[string]any, fieldName string, options []*Label, placeholder, description string) map[string]any {
	_ = placeholder
	// 查找目标字段
	fieldInterface, exists := properties[fieldName]
	if !exists {
		return nil
	}

	field, ok := fieldInterface.(map[string]any)
	if !ok {
		return nil
	}

	// 获取字段类型，默认为 string
	fieldType, _ := field["type"].(string)
	if fieldType == "" {
		fieldType = "string"
	}

	// 准备枚举值和选项
	enumValues := make([]any, len(options))
	enumOptions := make([]map[string]any, len(options))

	for i, opt := range options {
		enumValues[i] = opt.Value
		enumOptions[i] = map[string]any{
			"value": opt.Value,
			"label": opt.Label,
			"uuid":  opt.UUID,
		}
	}

	switch fieldType {
	case "object":
		// 对象类型：改为string类型，但添加特殊标识_object_selection表示需要转换为对象
		field["type"] = "string"
		field["enum"] = enumValues
		field["enumOptions"] = enumOptions
		field["description"] = description
		field["_object_selection"] = true // 特殊标识：表示这是一个对象选择字段

		// UI配置
		return map[string]any{
			"ui:widget":      "select",
			"ui:placeholder": placeholder,
		}

	case "array":
		// 数组类型：处理数组项
		items, ok := field["items"].(map[string]any)
		if !ok {
			items = map[string]any{"type": "string"}
		}

		itemsType, _ := items["type"].(string)
		if itemsType == "" {
			itemsType = "string"
		}

		switch itemsType {
		case "string":
			// 字符串数组：每个项都是字符串选项
			field["type"] = "array"
			field["items"] = map[string]any{
				"type":        "string",
				"enum":        enumValues,
				"enumOptions": enumOptions,
			}
			field["description"] = description

			// UI配置
			return map[string]any{
				"ui:widget":      "ArraySelectWidget",
				"ui:placeholder": placeholder,
			}

		case "object":
			// 对象数组：每个项都是对象格式 {id: name}
			// 为对象数组生成枚举选项，name字段只保留路径最后一部分
			var objEnumValues []any
			var objEnumOptions []map[string]any

			for _, opt := range options {
				// 从完整的value路径中提取最后一部分作为name
				// 例如："/PRCXI_Deck/PlateT11" -> "PlateT11"
				nameParts := strings.Split(opt.Value, "/")
				var cleanName string

				if len(nameParts) > 0 && nameParts[len(nameParts)-1] != "" {
					cleanName = nameParts[len(nameParts)-1]
				} else if len(nameParts) > 1 {
					cleanName = nameParts[len(nameParts)-2]
				} else {
					cleanName = opt.Value
				}

				objValue := map[string]any{
					"id":   opt.Value,
					"name": cleanName,
				}

				objEnumValues = append(objEnumValues, objValue)
				objEnumOptions = append(objEnumOptions, map[string]any{
					"value": objValue,
					"label": opt.Label,
					"uuid":  opt.UUID,
				})
			}

			field["type"] = "array"
			field["items"] = map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":   map[string]any{"type": "string"},
					"name": map[string]any{"type": "string"},
				},
				"enum":        objEnumValues,
				"enumOptions": objEnumOptions,
			}
			field["description"] = description

			// UI配置
			return map[string]any{
				"ui:widget":      "ArraySelectWidget",
				"ui:placeholder": placeholder,
			}

		default:
			return nil
		}

	default:
		// 默认字符串类型：保持原有逻辑
		field["type"] = "string"
		field["enum"] = enumValues
		field["enumOptions"] = enumOptions
		field["description"] = description

		// UI配置
		return map[string]any{
			"ui:widget":      "select",
			"ui:placeholder": placeholder,
		}
	}
}
