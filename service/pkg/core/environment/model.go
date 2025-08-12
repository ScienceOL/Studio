package environment

import (
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"gorm.io/datatypes"
)

type LaboratoryEnvReq struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type LaboratoryEnvResp struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

func (l *LaboratoryEnvResp) GetUUIDString() string {
	return l.UUID.String()
}

type UpdateEnvReq struct {
	UUID        uuid.UUID `json:"uuid" binding:"required"`
	Name        string    `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
}

type LaboratoryResp struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
}

type RegAction struct {
	Feedback    datatypes.JSON `json:"feedback"`
	Goal        datatypes.JSON `json:"goal"`
	GoalDefault datatypes.JSON `json:"goal_default"`
	Result      datatypes.JSON `json:"result"`
	Schema      datatypes.JSON `json:"schema"`
	Type        string         `json:"type"`
	Handles     datatypes.JSON `json:"handles"`
}

type RegClass struct {
	ActionValueMappings map[string]RegAction `json:"action_value_mappings"`
	Module              string               `json:"module"`
	StatusTypes         datatypes.JSON       `json:"status_types"`
	Type                string               `json:"type"`
}

type RegHandle struct {
	DataKey     string `json:"data_key"`
	DataSource  string `json:"data_source"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
	HandlerKey  string `json:"handler_key"`
	IoType      string `json:"io_type"`
	Label       string `json:"label"`
	Side        string `json:"side"`
}

type RegSchema struct {
	Properties datatypes.JSON `json:"properties"`
	Required   []string       `json:"required"`
	Type       string         `json:"type"`
}

type RegInitParamSchema struct {
	Data   *RegSchema `json:"data,omitempty"`
	Config *RegSchema `json:"config,omitempty"`
}

type ResourceReq struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	RegName         string              `json:"id" binding:"required"`
	Description     *string             `json:"description,omitempty"`
	Icon            string              `json:"icon,omitempty"`
	Language        string              `json:"registry_type" binding:"required"`
	Version         string              `json:"version" default:"0.0.1"`
	FilePath        string              `json:"file_path"`
	Class           RegClass            `json:"class"`
	Handles         []*RegHandle        `json:"handles"`
	InitParamSchema *RegInitParamSchema `json:"init_param_schema,omitempty"`
	Model           datatypes.JSON      `json:"model"`
}
