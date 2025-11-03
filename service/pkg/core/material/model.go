package material

import (
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/model"
	"gorm.io/datatypes"
)

type ActionType string

const (
	FetchGraph        ActionType = "fetch_graph"
	FetchTemplate     ActionType = "fetch_template"
	TemplateDetail    ActionType = "template_detail"
	SaveGraph         ActionType = "save_graph"
	CreateNode        ActionType = "create_node"
	UpdateNode        ActionType = "update_node"
	BatchDelNode      ActionType = "batch_del_nodes"
	BatchCreateEdge   ActionType = "batch_create_edges"
	BatchDelEdge      ActionType = "batch_del_edges"
	UpdateNodeData    ActionType = "update_node_data" // 只更更新 data
	UpdateNodeResData ActionType = "update_material_resource_data"
	UpdateNodeCreate  ActionType = "update_node_create"
)

type GraphNodeReq struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

type Node struct {
	UUID       uuid.UUID                      `json:"uuid"`
	ParentUUID uuid.UUID                      `json:"parent_uuid"`
	DeviceID   string                         `json:"id" binding:"required"`   // 实际是数据库的 name
	Name       string                         `json:"name" binding:"required"` // 实际是数据库的 display name
	Type       model.DEVICETYPE               `json:"type" binding:"required"`
	Class      string                         `json:"class" binding:"required"`
	Children   []string                       `json:"children,omitempty"`
	Parent     string                         `json:"parent" default:""`
	Pose       datatypes.JSONType[model.Pose] `json:"pose" swaggertype:"object"`
	Config     datatypes.JSON                 `json:"config" swaggertype:"object"`
	Data       datatypes.JSON                 `json:"data" swaggertype:"object"`
	// FIXME: 这块后续要优化掉，从 reg 获取
	Schema      datatypes.JSON  `json:"schema" swaggertype:"object"`
	Description *string         `json:"description,omitempty"`
	Model       datatypes.JSON  `json:"model" swaggertype:"object"`
	Position    model.Position  `json:"position"`
	Position3D  *model.Position `json:"position_3d,omitempty"`
	Extra       datatypes.JSON  `json:"extra,omitempty" swaggertype:"object"`
}

type GraphEdge struct {
	Edges []*Edge `json:"edges"`
}

type Edge struct {
	SourceUUID uuid.UUID `json:"source_uuid"`
	TargetUUID uuid.UUID `json:"target_uuid"`
	Source     string    `json:"source"`
	Target     string    `json:"target"`
	// FIXME: 下面两个字段命令 unilab 需要修改命名称
	SourceHandle string `json:"sourceHandle"`
	TargetHandle string `json:"targetHandle"`
	Type         string `json:"type"`
}

type LabWS struct {
	LabUUID uuid.UUID `uri:"lab_uuid" binding:"required"`
}

type DownloadMaterial struct {
	LabUUID uuid.UUID `uri:"lab_uuid" binding:"required"`
}

type TemplateReq struct {
	TemplateUUID uuid.UUID `uri:"template_uuid" binding:"required"`
}

// ================= websocket 更新物料

type ResourceHandleTemplate struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	IOType      string    `json:"io_type"`
	Source      string    `json:"source"`
	Key         string    `json:"key"`
	Side        string    `json:"side"`
}

type DeviceParamTemplate struct {
	UUID        uuid.UUID      `json:"uuid"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Placeholder string         `json:"placeholder"`
	Schema      datatypes.JSON `json:"schema" swaggertype:"object"`
}

type TemplateResp struct {
	Handles      []*ResourceHandleTemplate                 `json:"handles"`
	UUID         uuid.UUID                                 `json:"uuid"`
	ParentUUID   uuid.UUID                                 `json:"parent_uuid"`
	Name         string                                    `json:"name"`
	UserID       string                                    `json:"user_id"`
	Header       string                                    `json:"header"`
	Footer       string                                    `json:"footer"`
	Version      string                                    `json:"version"`
	Icon         string                                    `json:"icon"`
	Description  *string                                   `json:"description"`
	Model        datatypes.JSON                            `json:"model" swaggertype:"object"`
	Module       string                                    `json:"module"`
	Language     string                                    `json:"language"`
	StatusTypes  datatypes.JSON                            `json:"status_types" swaggertype:"object"`
	Tags         datatypes.JSONSlice[string]               `json:"tags" swaggertype:"array,string"`
	DataSchema   datatypes.JSON                            `json:"data_schema" swaggertype:"object"`
	ConfigSchema datatypes.JSON                            `json:"config_schema" swaggertype:"object"`
	ResourceType string                                    `json:"resource_type"`
	ConfigInfos  datatypes.JSONSlice[model.ResourceConfig] `json:"config_infos,omitempty" swaggertype:"array,object"`
	Pose         datatypes.JSONType[model.Pose]            `json:"pose" swaggertype:"object"`
}

type ResourceTemplate struct {
	UUID         uuid.UUID                   `json:"uuid"`
	Name         string                      `json:"name"`
	Tags         datatypes.JSONSlice[string] `json:"tags" swaggertype:"array,string"`
	ResourceType string                      `json:"resource_type"`
}

type ResourceTemplates struct {
	Templates []*ResourceTemplate `json:"templates"`
}

// 前端获取 materials 相关数据
type WSHandle struct {
	// NodeUUID    uuid.UUID `json:"node_uuid"`
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Side        string    `json:"side"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	IOType      string    `json:"io_type"`
	Source      string    `json:"source"`
	Key         string    `json:"key"`
}

type WSNode struct {
	UUID            uuid.UUID                      `json:"uuid"`
	ParentUUID      uuid.UUID                      `json:"parent_uuid"`
	Name            string                         `json:"name"`
	DisplayName     string                         `json:"display_name"`
	Description     *string                        `json:"description"`
	Type            model.DEVICETYPE               `json:"type"`
	ResTemplateUUID uuid.UUID                      `json:"res_template_uuid"`
	ResTemplateName string                         `json:"res_template_name"`
	InitParamData   datatypes.JSON                 `json:"init_param_data" swaggertype:"object"`
	Schema          datatypes.JSON                 `json:"schema,omitempty" swaggertype:"object"`
	Data            datatypes.JSON                 `json:"data" swaggertype:"object"`
	PlateWellDatas  map[string]datatypes.JSON      `json:"plate_well_datas" swaggertype:"object"`
	Status          string                         `json:"status"`
	Header          string                         `json:"header"`
	Pose            datatypes.JSONType[model.Pose] `json:"pose" swaggertype:"object"`
	Model           datatypes.JSON                 `json:"model" swaggertype:"object"`
	Icon            string                         `json:"icon"`
	Handles         []*WSHandle                    `json:"handles"`
}

type WSEdge struct {
	UUID             uuid.UUID `json:"uuid"`
	SourceNodeUUID   uuid.UUID `json:"source_node_uuid"`
	TargetNodeUUID   uuid.UUID `json:"target_node_uuid"`
	SourceHandleUUID uuid.UUID `json:"source_handle_uuid"`
	TargetHandleUUID uuid.UUID `json:"target_handle_uuid"`
	Type             string    `json:"type"`
}

type WSGraph struct {
	Nodes []*WSNode `json:"nodes"`
	Edges []*WSEdge `json:"edges"`
}

// 创建节点
type WSNodes struct {
	Nodes []*Node `json:"nodes"`
}

type UpdateNodeInfo struct {
	OldNodeUUID uuid.UUID
	NewNode     *Node
}

// 更新节点
type WSUpdateNodes struct {
	Nodes []*UpdateNodeInfo
}

// 添加边
type WSNodeEdges struct {
	Edges []*Edge
}

// 删除边
type WSDelNodeEdges struct {
	EdgeUUID []string `json:"edge_uuid"`
}

// 更新边
type WSUpdateNodeEdge struct {
	OldEdge uuid.UUID
	Edge    *Edge
}

// 更新节点
type WSUpdateNode struct {
	UUID          uuid.UUID                       `json:"uuid"`
	ParentUUID    *uuid.UUID                      `json:"parent_uuid,omitempty"`
	DisplayName   *string                         `json:"display_name,omitempty"`
	Description   *string                         `json:"description,omitempty"`
	InitParamData *datatypes.JSON                 `json:"init_param_data,omitempty"`
	Data          *datatypes.JSON                 `json:"data,omitempty"`
	Pose          *datatypes.JSONType[model.Pose] `json:"pose,omitempty"`
	Schema        *datatypes.JSON                 `json:"schema,omitempty"`
	Extra         *datatypes.JSON                 `json:"extra,omitempty"`
}

type InnerBaseConfig struct {
	Rotation model.Rotation `json:"rotation"`
	Category string         `json:"category"`
	SizeX    float32        `json:"size_x"`
	SizeY    float32        `json:"size_y"`
	SizeZ    float32        `json:"size_z"`
	Type     string         `json:"type"`
}

type StartMachineReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" form:"lab_uuid" uri:"lab_uuid" binding:"required"`
}

type StartMachineRes struct {
	MachineUUID uuid.UUID `json:"machine_uuid" binding:"required"`
}

type DelMachineReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" form:"lab_uuid" uri:"lab_uuid" binding:"required"`
}

type DelMachineRes struct{}

type StopMachineReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" form:"lab_uuid" uri:"lab_uuid" binding:"required"`
}

type StopMachineRes struct{}

type MachineStatusReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" form:"lab_uuid" uri:"lab_uuid" binding:"required"`
}

type MachineStatus string

const (
	UnknowStatus MachineStatus = "unknown"
	Deleted      MachineStatus = "deleted"
	Stoped       MachineStatus = "stopped"
	Stoping      MachineStatus = "stopping"
	Building     MachineStatus = "building"
	Running      MachineStatus = "running"
	Pending      MachineStatus = "pending"
	NotExist     MachineStatus = "not_exist"
)

type MachineStatusRes struct {
	Status MachineStatus
}

type MaterialReq struct {
	ID           string `form:"id"`
	WithChildren bool   `form:"with_children"`
}

type MaterialQueryReq struct {
	UUIDS []uuid.UUID `json:"uuids"`
}

type MaterialQueryResp struct {
	Nodes []*EdgeNode `json:"nodes"`
}

type EdgeNode struct {
	UUID        uuid.UUID        `json:"uuid"`
	ParentUUID  uuid.UUID        `json:"parent_uuid"`
	Name        string           `json:"name"`
	DisplayName string           `json:"display_name"`
	Description *string          `json:"description"`
	Class       string           `json:"class"`
	Status      string           `json:"status"`
	Type        model.DEVICETYPE `json:"type"`
	Config      datatypes.JSON   `json:"config" swaggertype:"object"`
	Schema      datatypes.JSON   `json:"schema" swaggertype:"object"`
	Data        datatypes.JSON   `json:"data" swaggertype:"object"`
	Pose        model.Pose       `json:"pose"`
	Model       datatypes.JSON   `json:"model" swaggertype:"object"`
	Icon        string           `json:"icon"`
	Extra       datatypes.JSON   `json:"extra,omitempty" swaggertype:"object"`
}

type HandleData struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	HasConnected bool   `json:"hasConnected"`
	Required     bool   `json:"required"`
}

type DataParam struct {
	ParamDataKey   string         `json:"param_data_key"`
	ParamType      string         `json:"param_type"`
	Title          string         `json:"title"`
	ParamInputData datatypes.JSON `json:"param_input_data" swaggertype:"object"`
	SelectChoices  string         `json:"select_choices"`
	Attachment     string         `json:"attachment"`
}

type MaterialData struct {
	Header       string        `json:"header"`
	Handles      []*HandleData `json:"handles"`
	Params       []*DataParam  `json:"params"`
	Executors    []any         `json:"executors"`
	Footer       string        `json:"footer"`
	NodeCardIcon string        `json:"node_card_icon"`
}

type MaterialResp struct {
	ID             string         `json:"id"`
	CloudUUID      uuid.UUID      `json:"cloud_uuid"`
	Type           string         `json:"type"`
	Data           any            `json:"data" swaggertype:"object"`
	Position       model.Position `json:"position"`
	Status         string         `json:"status"`
	Minimized      bool           `json:"minimized"`
	Disabled       bool           `json:"disabled"`
	Version        string         `json:"version"`
	DragHandle     string         `json:"dragHandle"`
	DeviceID       string         `json:"device_id"`
	Name           string         `json:"name"`
	ExperimentEnv  int64          `json:"experiment_env"`
	Description    *string        `json:"description"`
	Collapsed      bool           `json:"collapsed"`
	Width          float32        `json:"width"`
	Height         float32        `json:"height"`
	ChildNodesUUID []string       `json:"child_nodes_uuid"`
	ParentNodeUUID string         `json:"parent_node_uuid"`
	EqType         string         `json:"eq_type"`
	Dirs           []string       `json:"dirs"`
	Config         datatypes.JSON `json:"config" swaggertype:"object"`
	Parent         string         `json:"parent"`
	Children       []string       `json:"children"`
	Class          string         `json:"class,omitempty"`

	// positionAbsolute  "positionAbsolute": node.positionAbsolute,
}

type UpdateMaterialReq struct {
	Nodes []*Node `json:"nodes"`
}

type UpdateMaterialData struct {
	UUID uuid.UUID `json:"uuid"`
	Data any       `json:"data"`
}

type UpdateMaterialDeviceNotify struct {
	Action string                `json:"action"`
	Data   []*UpdateMaterialData `json:"data"`
}

type UpdateMaterialResNotify struct {
	Action string    `json:"action"`
	Data   []*WSNode `json:"data"`
}

type ResourceReq struct {
	LabUUID uuid.UUID        `json:"lab_uuid" form:"lab_uuid" uri:"lab_uuid"`
	Type    model.DEVICETYPE `json:"type" form:"type" uri:"type"`
}

type ResourceInfo struct {
	UUID       uuid.UUID `json:"uuid"`
	Name       string    `json:"name"`
	ParentUUID uuid.UUID `json:"parent_uuid"`
}

type ResourceResp struct {
	ResourceNameList []*ResourceInfo `json:"resource_name_list"`
}

// ResourceTemplateReq 获取资源模板详细信息的请求
type ResourceTemplateReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" form:"lab_uuid" uri:"lab_uuid" binding:"required"`
}

// ResourceTemplateInfo 资源模板详细信息
type ResourceTemplateInfo struct {
	UUID          uuid.UUID      `json:"uuid"`
	Name          string         `json:"name"`
	Icon          string         `json:"icon"`
	Description   *string        `json:"description"`
	ResourceType  string         `json:"resource_type"`
	Language      string         `json:"language"`
	Version       string         `json:"version"`
	Module        string         `json:"module"`
	Model         datatypes.JSON `json:"model" swaggertype:"object"`
	DataSchema    datatypes.JSON `json:"data_schema" swaggertype:"object"`
	ConfigSchema  datatypes.JSON `json:"config_schema" swaggertype:"object"`
	Tags          []string       `json:"tags"`
	Actions       []*ActionInfo  `json:"actions"`        // 该资源支持的动作列表
	MaterialCount int            `json:"material_count"` // 该资源实例化的物料数量
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

// ActionInfo 动作信息
type ActionInfo struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Schema      datatypes.JSON `json:"schema" swaggertype:"object"`
	Goal        datatypes.JSON `json:"goal" swaggertype:"object"`
	GoalDefault datatypes.JSON `json:"goal_default" swaggertype:"object"`
	Feedback    datatypes.JSON `json:"feedback" swaggertype:"object"`
	Result      datatypes.JSON `json:"result" swaggertype:"object"`
}

// ResourceTemplateResp 资源模板列表响应
type ResourceTemplateResp struct {
	Templates []*ResourceTemplateInfo `json:"templates"`
}

type DeviceActionReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" uri:"lab_uuid" form:"lab_uuid" binding:"required"`
	Name    string    `json:"name" uri:"name" form:"name" binding:"required"`
}

type DeviceAction struct {
	Action     string         `json:"action"`
	Schema     datatypes.JSON `json:"schema" swaggertype:"object"`
	ActionType string         `json:"action_type"`
}

type DeviceActionResp struct {
	Name    string          `json:"name"`
	Actions []*DeviceAction `json:"actions"`
}

type UpdatePair struct {
	ReqNode *Node
	DBNode  *model.MaterialNode
}

type SaveGrapReq struct {
	LabUUID uuid.UUID `json:"lab_uuid"`
	Graph   WSGraph   `json:"graph"`
}

// ==================== 新版带 uuid 版本
type CreateMaterialReq struct {
	Nodes []*Material `json:"nodes"`
}

type Material struct {
	UUID       uuid.UUID                      `json:"uuid" binding:"required"` // 有可能是 edge 编造 || 云端的
	ParentUUID uuid.UUID                      `json:"parent_uuid"`
	DeviceID   string                         `json:"id" binding:"required"`   // 实际是数据库的 name
	Name       string                         `json:"name" binding:"required"` // 实际是数据库的 display name
	Type       model.DEVICETYPE               `json:"type" binding:"required"`
	Class      string                         `json:"class" binding:"required"`
	Children   []string                       `json:"children,omitempty"`
	Parent     string                         `json:"parent" default:""`
	Pose       datatypes.JSONType[model.Pose] `json:"pose" swaggertype:"object"`
	Config     datatypes.JSON                 `json:"config" swaggertype:"object"`
	Data       datatypes.JSON                 `json:"data" swaggertype:"object"`
	// FIXME: 这块后续要优化掉，从 reg 获取
	Schema      datatypes.JSON  `json:"schema" swaggertype:"object"`
	Description *string         `json:"description,omitempty"`
	Model       datatypes.JSON  `json:"model" swaggertype:"object"`
	Position    model.Position  `json:"position"`
	Position3D  *model.Position `json:"position_3d,omitempty"`
	Icon        string          `json:"icon,omitempty"`
}

type CreateMaterialResp struct {
	UUID      uuid.UUID `json:"uuid"`
	CloudUUID uuid.UUID `json:"cloud_uuid"`
	DeviceID  string    `json:"id" binding:"required"`   // 实际是数据库的 name
	Name      string    `json:"name" binding:"required"` // 实际是数据库的 display name
}

type UpsertMaterialReq struct {
	MountUUID uuid.UUID   `json:"mount_uuid"`
	Nodes     []*Material `json:"nodes"`
}

type UpsertMaterialResp struct {
	UUID        uuid.UUID `json:"uuid"`
	CloudUUID   uuid.UUID `json:"cloud_uuid"`
	Name        string    `json:"name" binding:"required"`
	DisplayName string    `json:"display_name"`
}

type CreateMaterialEdgeReq struct {
	Edges []*MaterialEdge `json:"edges"`
}

// TODO: 后续是否添加边的映射关系
type CreateMaterialEdgeResp struct {
	UUID           uuid.UUID `json:"uuid"`
	SourceNodeUUID uuid.UUID `json:"source_node_uuid"`
	TargetNodeUUID uuid.UUID `json:"target_node_uuid"`
	SourceHandle   string    `json:"source_handle"`
	TargetHandle   string    `json:"target_handle"`
}

type MaterialEdge struct {
	SourceUUID   uuid.UUID `json:"source_uuid"`
	TargetUUID   uuid.UUID `json:"target_uuid"`
	SourceHandle string    `json:"sourceHandle"`
	TargetHandle string    `json:"targetHandle"`
	Type         string    `json:"type"`
}

type DownloadMaterialResp struct {
	Nodes []*Node `json:"nodes"`
}
