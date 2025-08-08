package model

import "gorm.io/datatypes"

type ResourceNodeTemplate struct {
	BaseModel
	Name         string                      `gorm:"type:varchar(255);not null;uniqueIndex:idx_rnt_lnv,priority:2" json:"name"`
	LabID        int64                       `gorm:"type:bigint;not null;uniqueIndex:idx_rnt_lnv,priority:1" json:"lab_id"`
	UserID       string                      `gorm:"type:varchar(120);not null" json:"user_id"`
	Header       string                      `gorm:"type:text" json:"header"`
	Footer       string                      `gorm:"type:text" json:"footer"`
	Version      string                      `gorm:"type:varchar(50);not null;default:'1.0.0';uniqueIndex:idx_rnt_lnv,priority:3" json:"version"`
	Icon         string                      `gorm:"type:text" json:"icon"`
	Description  *string                     `gorm:"type:text" json:"description"`
	Model        datatypes.JSON              `gorm:"type:jsonb;" json:"model"`
	Module       string                      `gorm:"type:varchar(1024)" json:"module"`
	Language     string                      `gorm:"type:varchar(255);not null;" json:"language"`
	StatusTypes  datatypes.JSON              `gorm:"type:jsonb" json:"status_types"`
	Tags         datatypes.JSONSlice[string] `gorm:"type:jsonb" json:"tags"` // label 标签
	DataSchema   datatypes.JSON              `gorm:"type:jsonb" json:"data_schema"`
	ConfigSchema datatypes.JSON              `gorm:"type:jsonb" json:"config_schema"`
	// ConfigInfo   datatypes.JSON `gorm:"type:jsonb" json:"config_info"` // FIXME: 拓展一张表，一个很大个 json object
}

func (*ResourceNodeTemplate) TableName() string {
	return "resource_node_template"
}

type DeviceAction struct {
	BaseModel
	ResNodeID   int64          `gorm:"type:bigint;not null;uniqueIndex:idx_res_node_id_name,priority:1" json:"res_node_id"`
	Name        string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_res_node_id_name,priority:2" json:"name"`
	Goal        datatypes.JSON `gorm:"type:jsonb" json:"goal"`
	GoalDefault datatypes.JSON `gorm:"type:jsonb" json:"goal_default"`
	Feedback    datatypes.JSON `gorm:"type:jsonb" json:"feedback"`
	Result      datatypes.JSON `gorm:"type:jsonb" json:"result"`
	Schema      datatypes.JSON `gorm:"type:jsonb" json:"schema"`
	Type        string         `gorm:"type:varchar(120);not null" json:"type"`
	Handles     datatypes.JSON `gorm:"type:jsonb" json:"handles"`
}

func (*DeviceAction) TableName() string {
	return "device_action"
}

type ResourceHandleTemplate struct {
	BaseModel
	NodeID      int64  `gorm:"type:bigint;not null;uniqueIndex:idx_dnht_dnhtnn,priority:1" json:"node_id"`
	Name        string `gorm:"type:varchar(255);not null;uniqueIndex:idx_dnht_dnhtnn,priority:2" json:"name"`
	DisplayName string `gorm:"type:varchar(255);not null" json:"display_name"`
	Type        string `gorm:"type:varchar(50);not null" json:"type"`
	IOType      string `gorm:"type:varchar(20);not null" json:"io_type"`
	Source      string `gorm:"type:varchar(255)" json:"source"`
	Key         string `gorm:"type:varchar(255);not null" json:"key"`
	Side        string `gorm:"type:varchar(20);not null" json:"side"`
}

func (*ResourceHandleTemplate) TableName() string {
	return "resource_handle_template"
}

// 节点模板
type ActionNodeTemplate struct {
	BaseModel
	Name        string
	DisplayName string
	

}

func (*ActionNodeTemplate) TableName() string {
	return "action_node_template"
}

// 节点模板 handle
type ActionHandleTemplate struct {
	BaseModel
	ActionID    int64
	Name        string
	DisplayName string
	Icon        string
}

func (*ActionHandleTemplate) TableName() string {
	return "action_handle_template"
}
