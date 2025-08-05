package material

import (
	"github.com/google/uuid"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
)

type MaterialActionType string

const (
	FetchNodes      MaterialActionType = "fetch_nodes"
	BatchCreateNode MaterialActionType = "batch_create_nodes"
	BatchUpdateNode MaterialActionType = "batch_update_nodes"
	BatchDelNode    MaterialActionType = "batch_del_nodes"
	BatchCreateEdge MaterialActionType = "batch_create_edges"
	BatchDelEdge    MaterialActionType = "batch_del_edges"
)

type GraphNode struct {
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
	LabUUID common.BinUUID `uri:"lab_uuid" binding:"required"`
}

// ================= websocket 更新物料

type WsMsgType struct {
	Action MaterialActionType `json:"action"`
}

// 创建节点
type WSNodes struct {
	Nodes []*Node `json:"nodes"`
}

// 删除节点
type WSDelNodes struct {
	NodeUUIDs []common.BinUUID `json:"node_uuids"`
}

type UpdateNodeInfo struct {
	OldNodeUUID common.BinUUID
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
	OldEdge common.BinUUID
	Edge    *Edge
}

type WSData[T any] struct {
	WsMsgType
	UUID uuid.UUID `json:"uuid"`
	Data T         `json:"data"`
}
