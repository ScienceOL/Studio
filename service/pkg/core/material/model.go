package material

import (
	"github.com/gofrs/uuid/v5"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
)

type MaterialActionType string

const (
	FetchGrpah      MaterialActionType = "fetch_graph"
	FetchTemplate   MaterialActionType = "fetch_template"
	BatchCreateNode MaterialActionType = "batch_create_nodes"
	BatchUpdateNode MaterialActionType = "batch_update_nodes"
	BatchDelNode    MaterialActionType = "batch_del_nodes"
	BatchCreateEdge MaterialActionType = "batch_create_edges"
	BatchDelEdge    MaterialActionType = "batch_del_edges"
)

type GraphNodeReq struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

type Node struct {
	DeviceID string           `json:"id" binding:"required"`   // 实际是数据库的 name
	Name     string           `json:"name" binding:"required"` // 实际是数据库的 display name
	Type     model.DEVICETYPE `json:"type" binding:"required"`
	Class    string           `json:"class" binding:"required"`
	Children []string         `json:"children,omitempty"`
	Parent   string           `json:"parent" default:""`
	Position datatypes.JSON   `json:"position"`
	Config   datatypes.JSON   `json:"config"`
	Data     datatypes.JSON   `json:"data"`
	// FIXME: 这块后续要优化掉，从 reg 获取
	Schema      datatypes.JSON `json:"schema"`
	Description *string        `json:"description,omitempty"`
	Model       string         `json:"model"`
}

type GraphEdge struct {
	Edges []*Edge `json:"edges"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	// FIXME: 下面两个字段命令 unilab 需要修改命名称
	SourceHandle string `json:"sourceHandle"`
	TargetHandle string `json:"targetHandle"`
	Type         string `json:"type"`
}

type LabWS struct {
	LabUUID uuid.UUID `uri:"lab_uuid" binding:"required"`
}

// ================= websocket 更新物料

type DeviceHandleTemplate struct {
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
	Schema      datatypes.JSON `json:"schema"`
}

type DeviceTemplate struct {
	Handles     []*DeviceHandleTemplate `json:"handles"`
	Params      []*DeviceParamTemplate  `json:"params"`
	UUID        uuid.UUID               `json:"uuid"`
	Name        string                  `json:"name"`
	UserID      string                  `json:"user_id"`
	Header      string                  `json:"header"`
	Footer      string                  `json:"footer"`
	Version     string                  `json:"version"`
	Icon        string                  `json:"icon"`
	Description *string                 `json:"description"`
}

// 前端获取 materials 相关数据
type WSHandle struct {
	UUID        uuid.UUID `json:"uuid"`
	NodeUUID    uuid.UUID `json:"node_uuid"`
	Name        string    `json:"name"`
	Side        string    `json:"side"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	IOType      string    `json:"io_type"`
	Source      string    `json:"source"`
	Key         string    `json:"key"`
	Connected   bool      `json:"connected"`
	Required    bool      `json:"required"`
}

type WSNode struct {
	UUID                 uuid.UUID        `json:"uuid"`
	ParentUUID           uuid.UUID        `json:"parent_uuid"`
	Name                 string           `json:"name"`
	DisplayName          string           `json:"display_name"`
	Description          *string          `json:"description"`
	Type                 model.DEVICETYPE `json:"type"`
	DeviceNodeTemplateID int64            `json:"device_node_template_id"`
	RegID                int64            `json:"reg_id"`
	InitParamData        datatypes.JSON   `json:"init_param_data"`
	Schema               datatypes.JSON   `json:"schema"`
	Data                 datatypes.JSON   `json:"data"`
	Dirs                 datatypes.JSON   `json:"dirs"`
	Position             datatypes.JSON   `json:"position"`
	Pose                 datatypes.JSON   `json:"pose"`
	Model                string           `json:"model"`
	Icon                 string           `json:"icon"`
	Handles              []*WSHandle      `json:"handles"`
}

type WSEdge struct {
	UUID             uuid.UUID `json:"uuid,omitempty"`
	SourceNodeUUID   uuid.UUID `json:"source_node_uuid"`
	TargetNodeUUID   uuid.UUID `json:"target_node_uuid"`
	SourceHandleUUID uuid.UUID `json:"source_handle_uuid"`
	TargetHandleUUID uuid.UUID `json:"target_handle_uuid"`
}

type GraphResp struct {
	Nodes []*WSNode `json:"nodes"`
	Edges []*WSEdge `json:"edges"`
}

type DeviceTemplates struct {
	Templates []*DeviceTemplate `json:"templates"`
}

type WsMsgType struct {
	Action  MaterialActionType `json:"action"`
	MsgUUID uuid.UUID          `json:"msg_uuid"`
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
type WDUpdateNodeEdge struct {
	OldEdge uuid.UUID
	Edge    *Edge
}

type WSData[T any] struct {
	WsMsgType
	Data T `json:"data"`
}
