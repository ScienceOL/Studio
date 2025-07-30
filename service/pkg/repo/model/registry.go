package model

import "gorm.io/datatypes"

type REGSTATUS string

const (
	REGINIT REGSTATUS = "init" // 初始化
	REGDEL  REGSTATUS = "del"  // 删除
)

// 注册表
type Registry struct {
	BaseModelNoUUID
	Name         string                      `gorm:"type:varchar(1024);not null" json:"name"`
	LabID        int64                       `gorm:"type:bigint;not null;index:idx_lab_id" json:"lab_id"`
	Status       REGSTATUS                   `gorm:"type:varchar(20);not null;" json:"status"`
	Module       string                      `gorm:"type:varchar(1024)" json:"module"`
	Model        datatypes.JSON              `gorm:"type:jsonb;" json:"model"`
	Type         string                      `gorm:"type:varchar(255);not null;" json:"type"`
	RegsitryType string                      `gorm:"type:varchar(255);not null;" json:"registry_type"`
	Version      string                      `gorm:"type:varchar(255);not null" json:"version"`
	Labels       datatypes.JSONSlice[string] `gorm:"type:jsonb" json:"labels"` // label 标签
	// ConfigInfo   datatypes.JSON `gorm:"type:jsonb" json:"config_info"` // FIXME: 拓展一张表，一个很大个 json object
	StatusTypes datatypes.JSON `gorm:"type:jsonb" json:"status_types"`
	Icon        string         `gorm:"type:text" json:"icon"`
	Description *string        `gorm:"type:text" json:"description"`
}

func (*Registry) TableName() string {
	return "registry"
}

type RegAction struct {
	BaseModelNoUUID
	RegID       int64          `gorm:"type:bigint;not null;uniqueIndex:idx_reg_name,priority:1" json:"reg_id"`
	Name        string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_reg_name,priority:2" json:"name"`
	Goal        datatypes.JSON `gorm:"type:jsonb" json:"goal"`
	GoalDefault datatypes.JSON `gorm:"type:jsonb" json:"goal_default"`
	Feedback    datatypes.JSON `gorm:"type:jsonb" json:"feedback"`
	Result      datatypes.JSON `gorm:"type:jsonb" json:"result"`
	Schema      datatypes.JSON `gorm:"type:jsonb" json:"schema"`
	Type        string         `gorm:"type:varchar(120);not null" json:"type"`
	Handles     datatypes.JSON `gorm:"type:jsonb" json:"handles"`
}

func (*RegAction) TableName() string {
	return "reg_action"
}
