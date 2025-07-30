package material

import (
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
)

type Node struct {
	DeviceID string           `json:"id" binding:"required"`
	Name     string           `json:"name" binding:"required"`
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

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	// FIXME: 下面两个字段命令 unilab 需求
	// FIXME: handle 是按照 node ID、handle key、side 作为唯一主键的，
	// 这个连接信息缺少 side 信息，有 bug，
	// 不知道真实的连线是连接到哪个 side 的 handle key
	SourceHandle string `json:"sourceHandle"`
	TargetHandle string `json:"targetHandle"`
	Type         string `json:"type"`
}
