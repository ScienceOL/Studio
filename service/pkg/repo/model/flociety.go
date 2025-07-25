package model

import "gorm.io/datatypes"

type DeviceNodeTemplate struct {
	BaseModel
	Name        string  `gorm:"type:varchar(255);not null;uniqueIndex:idx_lrnv,priority:3" json:"name"`
	LabID       int64   `gorm:"type:bigint;not null;uniqueIndex:idx_lrnv,priority:1" json:"lab_id"`
	RegID       int64   `gorm:"type:bigint;not null;uniqueIndex:idx_lrnv,priority:2" json:"reg_id"`
	UserID      string  `gorm:"type:varchar(120);not null" json:"user_id"`
	Header      string  `gorm:"type:text" json:"header"`
	Footer      string  `gorm:"type:text" json:"footer"`
	Version     string  `gorm:"type:varchar(50);not null;default:'1.0.0';uniqueIndex:idx_lrnv,priority:4" json:"version"`
	Icon        string  `gorm:"type:text" json:"icon"`
	Description *string `gorm:"type:text" json:"description"`
}

func (*DeviceNodeTemplate) TableName() string {
	return "device_node_template"
}

type DeviceNodeHandleTemplate struct {
	BaseModelNoUUID
	NodeID      int64  `gorm:"type:bigint;not null;uniqueIndex:idx_dnhtnn,priority:1" json:"node_id"`
	Name        string `gorm:"type:varchar(255);not null;uniqueIndex:idx_dnhtnn,priority:2" json:"name"`
	DisplayName string `gorm:"type:varchar(255);not null" json:"display_name"`
	Type        string `gorm:"type:varchar(50);not null" json:"type"`
	IOType      string `gorm:"type:varchar(20);not null" json:"io_type"`
	Source      string `gorm:"type:varchar(255)" json:"source"`
	Key         string `gorm:"type:varchar(255);not null" json:"key"`
	Side        string `gorm:"type:varchar(20);not null" json:"side"`
}

func (*DeviceNodeHandleTemplate) TableName() string {
	return "device_node_handle_template"
}

type DeviceNodeParamTemplate struct {
	BaseModelNoUUID
	NodeID      int64  `gorm:"type:bigint;not null;uniqueIndex:idx_dnptnn,priority:1" json:"node_id"`
	Name        string `gorm:"type:varchar(255);not null;uniqueIndex:idx_dnptnn,priority:2" json:"name"`
	Type        string `gorm:"type:varchar(50);not null" json:"type"`
	Placeholder string `gorm:"type:varchar(500)" json:"placeholder"`
	// InputData   datatypes.JSON `gorm:"type:json" json:"input_data"`
	// Choices     datatypes.JSON `gorm:"type:json" json:"choices"`
	Schema datatypes.JSON `gorm:"type:json" json:"schema"`
	// UISchema    datatypes.JSON `gorm:"type:json" json:"ui_schema"`
}

func (*DeviceNodeParamTemplate) TableName() string {
	return "device_node_param_template"
}

// TODO: NodeTemplateLibrary -----> ActionNodeTemplate
