package model

import (
	"github.com/scienceol/studio/service/pkg/common"
	"gorm.io/datatypes"
)

// TODO: 字段是否有部分可以删除，或者是合并到一个 json 字段内
type MaterialNode struct {
	BaseModel
	ParentID             int64          `gorm:"type:bigint;index:idx_parent_lab" json:"parent_id"`
	LabID                int64          `gorm:"type:bigint;not null;index:idx_parent_lab" json:"lab_id"`
	Name                 string         `gorm:"type:varchar(255);not null" json:"name"`
	DisplayName          string         `gorm:"type:varchar(255);not null" json:"display_name"`
	Description          *string        `gorm:"type:text" json:"description"`
	Status               string         `gorm:"type:varchar(20);not null;default:'idle'" json:"status"`
	Type                 string         `gorm:"type:varchar(20);not null" json:"type"`
	DeviceNodeTemplateID int64          `gorm:"type:bigint;index:idx_template" json:"device_node_template_id"`
	RegID                int64          `gorm:"type:bigint;index:idx_reg" json:"reg_id"`
	InitParamData        datatypes.JSON `gorm:"type:jsonb" json:"init_param_data"` // TODO: 这是原来的 config 对应的数据
	Schema               datatypes.JSON `gorm:"type:jsonb" json:"schema"`          // TODO: 从 registry 里面获取，需要 edge 配合修改
	Data                 datatypes.JSON `gorm:"type:jsonb" json:"data"`
	Dirs                 datatypes.JSON `gorm:"type:jsonb" json:"dirs"`
	Position             datatypes.JSON `gorm:"type:jsonb" json:"position"`
	Pose                 datatypes.JSON `gorm:"type:jsonb" json:"pose"`
	Model                string         `gorm:"type:varchar(1000)" json:"model"`
}

func (*MaterialNode) TableName() string {
	return "material_node"
}

type MaterialHandle struct {
	BaseModel
	NodeID      int64  `gorm:"type:bigint;not null;index:idx_node_id" json:"node_id"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string `gorm:"type:text" json:"display_name"`
	Type        string `gorm:"type:varchar(100);not null;default:'any'" json:"type"`
	IOType      string `gorm:"type:varchar(20);not null" json:"io_type"`
	Source      string `gorm:"type:varchar(100)" json:"source"`
	Key         string `gorm:"type:varchar(100);not null" json:"key"`
	Side        string `gorm:"type:varchar(20);not null" json:"side"`
	Connected   bool   `gorm:"default:false" json:"connected"`
	Required    bool   `gorm:"default:false" json:"required"`
}

func (*MaterialHandle) TableName() string {
	return "material_handle"
}

type MaterialEdge struct {
	BaseModelNoUUID
	SourceNodeUUID   common.BinUUID `gorm:"type:varchar(36);not null;index:idx_source_target" json:"source_node_uuid"`
	TargetNodeUUID   common.BinUUID `gorm:"type:varchar(36);not null;index:idx_source_target" json:"target_node_uuid"`
	SourceHandleUUID common.BinUUID `gorm:"type:varchar(36);not null" json:"source_handle_uuid"`
	TargetHandleUUID common.BinUUID `gorm:"type:varchar(36);not null" json:"target_handle_uuid"`
}

func (*MaterialEdge) TableName() string {
	return "material_edge"
}
