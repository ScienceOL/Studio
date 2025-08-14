package model

import "gorm.io/datatypes"

type ResourceNodeTemplate struct {
	BaseModel
	Name         string                      `gorm:"type:varchar(255);not null;uniqueIndex:idx_rnpt_lnv,priority:2" json:"name"`
	LabID        int64                       `gorm:"type:bigint;not null;uniqueIndex:idx_rnpt_lnv,priority:1" json:"lab_id"`
	UserID       string                      `gorm:"type:varchar(120);not null" json:"user_id"`
	ParentID     int64                       `gorm:"type:bigint;uniqueIndex:idx_rnt_lnpv,priority:3" json:"parent_id"`
	Header       string                      `gorm:"type:text" json:"header"`
	Footer       string                      `gorm:"type:text" json:"footer"`
	Version      string                      `gorm:"type:varchar(50);not null;default:'1.0.0';uniqueIndex:idx_rnt_lnpv,priority:4" json:"version"`
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
type WorkflowNodeTemplate struct {
	BaseModel
	Name                   string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_ant_ldmn,priority:4" json:"name"`       // action name
	LabID                  int64          `gorm:"type:bigint;not null;uniqueIndex:idx_ant_ldmn,priority:1" json:"lab_id"`           // 实验室 id
	ResourceNodeTemplateID int64          `gorm:"type:bigint;not null" json:"resource_node_template_id"`                            // 引用模板 id
	DeviceActionID         int64          `gorm:"type:bigint;not null;uniqueIndex:idx_ant_ldmn,priority:2" json:"device_action_id"` // 引用的对应 action id
	MaterialNodeID         int64          `gorm:"type:bigint;not null;uniqueIndex:idx_ant_ldmn,priority:3" json:"material_node_id"` // 物料 id
	DisplayName            string         `gorm:"type:varchar(255);not null" json:"display_name"`
	Header                 string         `gorm:"type:text" json:"header"`
	Footer                 *string        `gorm:"type:text" json:"footer"`
	ParamType              string         `gorm:"type:varchar(50);default:'DEFAULT'" json:"param_type"`
	Schema                 datatypes.JSON `gorm:"type:jsonb" json:"schema"`
	ExecuteScript          string         `gorm:"type:text" json:"execute_script"`
	NodeType               string         `gorm:"type:varchar(50);not null;default:'ILab'" json:"node_type"`
	Icon                   string         `gorm:"type:text" json:"icon"`
	// TODO: ParamDataKey string 现状都是 default ，是不是没有用？
}

func (*WorkflowNodeTemplate) TableName() string {
	return "workflow_node_template"
}

// 节点模板 handle
type WorkflowHandleTemplate struct {
	BaseModel
	NodeTemplateID int64  `gorm:"type:bigint;not null;uniqueIndex:idx_aht_ahi,priority:1" json:"node_template_id"` // 节点模板的 id
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
