package model

import (
	"fmt"

	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DEVICETYPE string

const (
	MATERIALREPO        DEVICETYPE = "repository"
	MATERIALPLATE       DEVICETYPE = "plate"
	MATERIALCONTAINER   DEVICETYPE = "container"
	MATERIALRESOURCE    DEVICETYPE = "resource"
	MATERIALDEVICE      DEVICETYPE = "device"
	MATERIALWELL        DEVICETYPE = "well"
	MATERIALTIP         DEVICETYPE = "tip"
	MATERIALTIPRACK     DEVICETYPE = "tip_rack"
	MATERIALTIPSPOT     DEVICETYPE = "tip_spot"
	MATERIALDECK        DEVICETYPE = "deck"
	MATERIALWORKSTATION DEVICETYPE = "workstation"
)

// TODO: 字段是否有部分可以删除，或者是合并到一个 json 字段内
type MaterialNode struct {
	BaseModel
	ParentID       int64                    `gorm:"type:bigint;index:idx_mn_p;uniqueIndex:idx_mn_lpn,priority:2" json:"parent_id"`
	LabID          int64                    `gorm:"type:bigint;not null;uniqueIndex:idx_mn_lpn,priority:1" json:"lab_id"`
	Name           string                   `gorm:"type:varchar(255);not null;uniqueIndex:idx_mn_lpn,priority:3" json:"name"`
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
	Model          datatypes.JSON           `gorm:"type:jsonb" json:"model"`
	Icon           string                   `gorm:"type:text" json:"icon"`
	Extra          datatypes.JSON           `gorm:"type:jsonb" json:"extra"`
	// Tags                   datatypes.JSONSlice[string] `gorm:"type:jsonb" json:"tags"` // label 标签

	ResourceNodeTemplate *ResourceNodeTemplate `gorm:"-"`
	EdgeUUID             uuid.UUID             `gorm:"-"`
}

func (m *MaterialNode) AfterFind(tx *gorm.DB) (err error) {
	addr := config.Global().Storage.Addr
	bucket := config.Global().Storage.Bucket
	m.Icon = fmt.Sprintf("%s/%s/media/device_icon/%s", addr, bucket, m.Icon)
	return nil
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

type MaterialLog struct {
	BaseModel
}

// 开发机
type MaterialMachine struct {
	BaseModel
	LabID     int64  `gorm:"type:bigint;not null;uniqueIndex:idx_mm_lum,priority:1" json:"lab_id"`
	UserID    string `gorm:"type:varchar(120);not null;uniqueIndex:idx_mm_lum,priority:2" json:"user_id"`
	ImageID   int64  `gorm:"type:bigint;not null;uniqueIndex:idx_mm_lum,priority:3" json:"image_id"`
	MachineID int64  `gorm:"type:bigint;not null" json:"machine_id"`
}

func (*MaterialMachine) TableName() string {
	return "material_machine"
}
