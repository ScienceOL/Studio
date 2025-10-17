package model

import (
	"gorm.io/datatypes"
)

// 资源模板
type ResourceNodeTemplate struct {
	BaseModel
	Name         string                      `gorm:"type:varchar(255);not null;uniqueIndex:idx_rnt_lnpv,priority:2" json:"name"`
	LabID        int64                       `gorm:"type:bigint;not null;uniqueIndex:idx_rnt_lnpv,priority:1" json:"lab_id"`
	UserID       string                      `gorm:"type:varchar(120);not null" json:"user_id"`
	ParentID     int64                       `gorm:"type:bigint;uniqueIndex:idx_rnt_lnpv,priority:3" json:"parent_id"`
	Header       string                      `gorm:"type:text" json:"header"`
	Footer       string                      `gorm:"type:text" json:"footer"`
	Icon         string                      `gorm:"type:text" json:"icon"`
	Description  *string                     `gorm:"type:text" json:"description"`
	Model        datatypes.JSON              `gorm:"type:jsonb;" json:"model"`
	Module       string                      `gorm:"type:varchar(1024)" json:"module"`
	ResourceType string                      `gorm:"type:varchar(255);not null;default:'device'" json:"resource_type"`
	Language     string                      `gorm:"type:varchar(255);not null;" json:"language"`
	StatusTypes  datatypes.JSON              `gorm:"type:jsonb" json:"status_types"`
	Tags         datatypes.JSONSlice[string] `gorm:"type:jsonb" json:"tags"` // label 标签
	DataSchema   datatypes.JSON              `gorm:"type:jsonb" json:"data_schema"`
	ConfigSchema datatypes.JSON              `gorm:"type:jsonb" json:"config_schema"`
	Pose         datatypes.JSONType[Pose]    `gorm:"type:jsonb;not null;default:'{}'" json:"pose"`
	Version      string                      `gorm:"type:varchar(50);not null;default:'1.0.0'" json:"version"`

	ParentNode *ResourceNodeTemplate   `gorm:"-"`
	ParentName string                  `gorm:"-"`
	ConfigInfo []*ResourceNodeTemplate `gorm:"-"`
}

func (*ResourceNodeTemplate) TableName() string {
	return "resource_node_template"
}

type ResourceHandleTemplate struct {
	BaseModel
	ResourceNodeID int64  `gorm:"type:bigint;not null;uniqueIndex:idx_rht_rdis,priority:1" json:"resource_node_id"` // 资源模板节点 id
	Name           string `gorm:"type:varchar(255);not null;uniqueIndex:idx_rht_rdis,priority:2" json:"name"`
	DisplayName    string `gorm:"type:varchar(255);not null" json:"display_name"`
	Type           string `gorm:"type:varchar(50);not null" json:"type"`
	IOType         string `gorm:"type:varchar(20);not null;uniqueIndex:idx_rht_rdis,priority:3" json:"io_type"`
	Source         string `gorm:"type:varchar(255)" json:"source"`
	Key            string `gorm:"type:varchar(255);not null" json:"key"`
	Side           string `gorm:"type:varchar(20);not null;uniqueIndex:idx_rht_rdis,priority:4" json:"side"`
}

func (*ResourceHandleTemplate) TableName() string {
	return "resource_handle_template"
}

type WorkflowNodeTemplate struct {
	BaseModel
	LabID          int64          `gorm:"type:bigint;not null;index:idx_da_lab_id" json:"lab_id"`
	ResourceNodeID int64          `gorm:"type:bigint;not null;uniqueIndex:idx_da_id_name,priority:1" json:"resource_node_id"` // 资源模板节点 id
	Name           string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_da_id_name,priority:2" json:"name"`
	Class          string         `gorm:"type:varchar(200)" json:"class"`
	Goal           datatypes.JSON `gorm:"type:jsonb" json:"goal"`
	GoalDefault    datatypes.JSON `gorm:"type:jsonb" json:"goal_default"`
	Feedback       datatypes.JSON `gorm:"type:jsonb" json:"feedback"`
	Result         datatypes.JSON `gorm:"type:jsonb" json:"result"`
	Schema         datatypes.JSON `gorm:"type:jsonb" json:"schema"`
	Type           string         `gorm:"type:text;not null" json:"type"`
	Icon           string         `gorm:"type:text" json:"icon"`
	Header         string         `gorm:"type:text" json:"header"`
	Footer         string         `gorm:"type:text" json:"footer"`

	Handles datatypes.JSONType[ActionHandle] `gorm:"-"`
}

func (*WorkflowNodeTemplate) TableName() string {
	return "workflow_node_template"
}

// 节点模板 handle
type WorkflowHandleTemplate struct {
	BaseModel
	WorkflowNodeID int64  `gorm:"type:bigint;not null;uniqueIndex:idx_aht_ahi,priority:1" json:"workflow_node_id"` // 工作流节点模板 id
	HandleKey      string `gorm:"type:varchar(100);not null;uniqueIndex:idx_aht_ahi,priority:2" json:"handle_key"`
	IoType         string `gorm:"type:varchar(10);not null;uniqueIndex:idx_aht_ahi,priority:3" json:"io_type"`
	DisplayName    string `gorm:"type:varchar(255);not null" json:"display_name"`
	Type           string `gorm:"type:varchar(100);not null" json:"type"`
	DataSource     string `gorm:"type:varchar(10)" json:"data_source"`
	DataKey        string `gorm:"type:varchar(100)" json:"data_key"`
}

func (*WorkflowHandleTemplate) TableName() string {
	return "workflow_handle_template"
}
