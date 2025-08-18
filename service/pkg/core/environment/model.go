package environment

import (
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
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
	Feedback    datatypes.JSON                         `json:"feedback"`
	Goal        datatypes.JSON                         `json:"goal"`
	GoalDefault datatypes.JSON                         `json:"goal_default"`
	Result      datatypes.JSON                         `json:"result"`
	Schema      datatypes.JSON                         `json:"schema"`
	Type        string                                 `json:"type"`
	Handles     datatypes.JSONType[model.ActionHandle] `json:"handles"`
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

type InnerBaseConfig struct {
	Rotation model.Rotation `json:"rotation"`
	Category string         `json:"category"`
	SizeX    float32        `json:"size_x"`
	SizeY    float32        `json:"size_y"`
	SizeZ    float32        `json:"size_z"`
	Type     string         `json:"type"`
}

type Config struct {
	Class    string         `json:"class"`
	Config   datatypes.JSON `json:"config"`
	Data     datatypes.JSON `json:"data"`
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Parent   string         `json:"parent"`
	Position model.Position `json:"position"`
	Type     string         `json:"type"`
	// SampleID
}

type Resource struct {
	RegName         string                      `json:"id" binding:"required"`
	Description     *string                     `json:"description,omitempty"`
	Icon            string                      `json:"icon,omitempty"`
	ResourceType    string                      `json:"registry_type" binding:"required"`
	Version         string                      `json:"version" default:"0.0.1"`
	FilePath        string                      `json:"file_path"`
	Class           RegClass                    `json:"class"`
	Handles         []*RegHandle                `json:"handles"`
	InitParamSchema *RegInitParamSchema         `json:"init_param_schema,omitempty"`
	Model           datatypes.JSON              `json:"model"`
	Tags            datatypes.JSONSlice[string] `json:"category"`
	ConfigInfo      []*Config                   `json:"config_info"`
}
