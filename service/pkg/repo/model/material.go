package model

import (
	"github.com/scienceol/studio/service/pkg/common"
	"gorm.io/datatypes"
)

type DEVICETYPE string

const (
	MATERIALREPO        DEVICETYPE = "repository"
	MATERIALPLATE       DEVICETYPE = "plate"
	MATERIALCONTAINER   DEVICETYPE = "container"
	MATERIALDEVICE      DEVICETYPE = "device"
	MATERIALWELL        DEVICETYPE = "well"
	MATERIALTIP         DEVICETYPE = "tip"
	MATERIALDECK        DEVICETYPE = "deck"
	MATERIALWORKSTATION DEVICETYPE = "workstation"
)

// TODO: 字段是否有部分可以删除，或者是合并到一个 json 字段内
type MaterialNode struct {
	BaseModel
	ParentID    int64   `gorm:"type:bigint" json:"parent_id"`
	LabID       int64   `gorm:"type:bigint;not null;uniqueIndex:idx_mn_ln,priority:1" json:"lab_id"`
	Name        string  `gorm:"type:varchar(255);not null;uniqueIndex:idx_mn_ln,priority:2" json:"name"`
	DisplayName string  `gorm:"type:varchar(255);not null" json:"display_name"`
	Description *string `gorm:"type:text" json:"description"`
	// Status               string         `gorm:"type:varchar(20);not null;default:'idle'" json:"status"`
	Type                 DEVICETYPE     `gorm:"type:varchar(20);not null" json:"type"`
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
	NodeID      int64  `gorm:"type:bigint;not null;uniqueIndex:idx_mh_nns,priority:1" json:"node_id"`
	Name        string `gorm:"type:varchar(20);not null;uniqueIndex:idx_mh_nns,priority:2" json:"name"`
	Side        string `gorm:"type:varchar(20);not null" json:"side"`
	DisplayName string `gorm:"type:text" json:"display_name"`
	Type        string `gorm:"type:varchar(100);not null;default:'any'" json:"type"`
	IOType      string `gorm:"type:varchar(20);not null" json:"io_type"`
	Source      string `gorm:"type:varchar(100)" json:"source"`
	Key         string `gorm:"type:varchar(100);not null" json:"key"`
	Connected   bool   `gorm:"default:false" json:"connected"`
	Required    bool   `gorm:"default:false" json:"required"`
}

func (*MaterialHandle) TableName() string {
	return "material_handle"
}

type MaterialEdge struct {
	BaseModelNoUUID
	SourceNodeUUID   common.BinUUID `gorm:"type:varchar(36);not null;uniqueIndex:idx_me_stst,priority:1" json:"source_node_uuid"`
	TargetNodeUUID   common.BinUUID `gorm:"type:varchar(36);not null;uniqueIndex:idx_me_stst,priority:2" json:"target_node_uuid"`
	SourceHandleUUID common.BinUUID `gorm:"type:varchar(36);not null;uniqueIndex:idx_me_stst,priority:3" json:"source_handle_uuid"`
	TargetHandleUUID common.BinUUID `gorm:"type:varchar(36);not null:uniqueIndex:idx_me_stst,priority:4" json:"target_handle_uuid"`
}

func (*MaterialEdge) TableName() string {
	return "material_edge"
}
