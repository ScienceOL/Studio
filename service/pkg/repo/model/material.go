package model

import (
	"github.com/scienceol/studio/service/pkg/common/uuid"
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
	ParentID       int64                    `gorm:"type:bigint;uniqueIndex:idx_mn_ln,priority:3" json:"parent_id"`
	LabID          int64                    `gorm:"type:bigint;not null;uniqueIndex:idx_mn_ln,priority:1" json:"lab_id"`
	Name           string                   `gorm:"type:varchar(255);not null;uniqueIndex:idx_mn_ln,priority:2" json:"name"`
	DisplayName    string                   `gorm:"type:varchar(255);not null" json:"display_name"`
	Description    *string                  `gorm:"type:text" json:"description"`
	Status         string                   `gorm:"type:varchar(20);not null;default:'idle'" json:"status"`
	Type           DEVICETYPE               `gorm:"type:varchar(20);not null" json:"type"`
	ResourceNodeID int64                    `gorm:"type:bigint;index:idx_template" json:"resource_node_id"` // 资源模板 id
	Class          string                   `gorm:"type:text" json:"class"`
	InitParamData  datatypes.JSON           `gorm:"type:jsonb" json:"init_param_data"` // TODO: 这是原来的 config 对应的数据
	Schema         datatypes.JSON           `gorm:"type:jsonb" json:"schema"`          // TODO: 从 registry 里面获取，需要 edge 配合修改
	Data           datatypes.JSON           `gorm:"type:jsonb" json:"data"`
	Pose           datatypes.JSONType[Pose] `gorm:"type:jsonb" json:"pose"`
	Model          datatypes.JSON           `gorm:"type:varchar(1000)" json:"model"`
	Icon           string                   `gorm:"type:text" json:"icon"`
	// Tags                   datatypes.JSONSlice[string] `gorm:"type:jsonb" json:"tags"` // label 标签
}

func (*MaterialNode) TableName() string {
	return "material_node"
}

type MaterialEdge struct {
	BaseModel
	SourceNodeUUID   uuid.UUID `gorm:"type:uuid;not null;index:idx_me_source_node;uniqueIndex:idx_me_stst,priority:1" json:"source_node_uuid"`
	TargetNodeUUID   uuid.UUID `gorm:"type:uuid;not null;index:idx_me_target_node;uniqueIndex:idx_me_stst,priority:2" json:"target_node_uuid"`
	SourceHandleUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_me_stst,priority:3" json:"source_handle_uuid"`
	TargetHandleUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_me_stst,priority:4" json:"target_handle_uuid"`
}

func (*MaterialEdge) TableName() string {
	return "material_edge"
}
