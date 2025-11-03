package environment

import (
	"time"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/model"
	"gorm.io/datatypes"
)

type LaboratoryEnvReq struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type LaboratoryEnvResp struct {
	UUID         uuid.UUID `json:"uuid"`
	Name         string    `json:"name"`
	AccessKey    string    `json:"access_key"`
	AccessSecret string    `json:"access_secret"`
}

func (l *LaboratoryEnvResp) GetUUIDString() string {
	return l.UUID.String()
}

type UpdateEnvReq struct {
	UUID        uuid.UUID `json:"uuid" binding:"required"`
	Name        string    `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
}

type DelLabReq struct {
	UUID uuid.UUID `json:"uuid" binding:"required"`
}

type LabType string

const (
	LABUUID LabType = "uuid"
	LABAK   LabType = "ak"
)

type LabInfoReq struct {
	UUID uuid.UUID `json:"uuid" form:"uuid" uri:"uuid" binding:"required"`
	Type LabType   `json:"type" form:"type" uri:"type"`
}

type LaboratoryResp struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	UserID      string    `json:"user_id"`
	Description *string   `json:"description"`
	MemberCount int64     `json:"member_count"`
	IsAdmin     bool      `json:"is_admin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LabInfoResp struct {
	UUID         uuid.UUID               `json:"uuid"`
	Name         string                  `json:"name"`
	UserID       string                  `json:"user_id"`
	IsAdmin      bool                    `json:"is_admin"`
	AccessKey    string                  `json:"access_key"`
	AccessSecret string                  `json:"access_secret"`
	Status       model.EnvironmentStatus `json:"status"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

type RegAction struct {
	Feedback    datatypes.JSON                         `json:"feedback" swaggertype:"object"`
	Goal        datatypes.JSON                         `json:"goal" swaggertype:"object"`
	GoalDefault datatypes.JSON                         `json:"goal_default" swaggertype:"object"`
	Result      datatypes.JSON                         `json:"result" swaggertype:"object"`
	Schema      datatypes.JSON                         `json:"schema" swaggertype:"object"`
	Type        string                                 `json:"type"`
	Handles     datatypes.JSONType[model.ActionHandle] `json:"handles" swaggertype:"object"`
}

type RegClass struct {
	ActionValueMappings map[string]RegAction `json:"action_value_mappings"`
	Module              string               `json:"module"`
	StatusTypes         datatypes.JSON       `json:"status_types" swaggertype:"object"`
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
	Properties datatypes.JSON `json:"properties" swaggertype:"object"`
	Required   []string       `json:"required"`
	Type       string         `json:"type"`
}

type RegInitParamSchema struct {
	Data   *RegSchema `json:"data,omitempty"`
	Config *RegSchema `json:"config,omitempty"`
}

type ResourceReq struct {
	Resources []*Resource `json:"resources"`
}

// type Config struct {
// 	Class    string         `json:"class"`
// 	Config   datatypes.JSON `json:"config"`
// 	Data     datatypes.JSON `json:"data"`
// 	ID       string         `json:"id"`
// 	Name     string         `json:"name"`
// 	Parent   string         `json:"parent"`
// 	Position model.Position `json:"position"`
// 	Type     string         `json:"type"`
// 	// SampleID
// }

type Resource struct {
	RegName         string                                    `json:"id" binding:"required"`
	Description     *string                                   `json:"description,omitempty"`
	Icon            string                                    `json:"icon,omitempty"`
	ResourceType    string                                    `json:"registry_type" binding:"required"`
	Version         string                                    `json:"version" default:"0.0.1"`
	FilePath        string                                    `json:"file_path"`
	Class           RegClass                                  `json:"class"`
	Handles         []*RegHandle                              `json:"handles"`
	InitParamSchema *RegInitParamSchema                       `json:"init_param_schema,omitempty"`
	Model           datatypes.JSON                            `json:"model" swaggertype:"object"`
	Tags            datatypes.JSONSlice[string]               `json:"category" swaggertype:"array,string"`
	ConfigInfo      datatypes.JSONSlice[model.ResourceConfig] `json:"config_info" swaggertype:"array,object"`

	SelfDB *model.ResourceNodeTemplate
}

type LabMemberReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" uri:"lab_uuid" form:"lab_uuid"`
	common.PageReq
}

type DelLabMemberReq struct {
	LabUUID    uuid.UUID `json:"lab_uuid" uri:"lab_uuid" form:"lab_uuid"`
	MemberUUID uuid.UUID `json:"member_uuid" uri:"member_uuid" form:"member_uuid"`
}

type LabMemberResp struct {
	UUID    uuid.UUID                  `json:"uuid"`
	UserID  string                     `json:"user_id"`
	LabID   int64                      `json:"lab_id"`
	Role    model.LaboratoryMemberRole `json:"role"`
	IsAdmin bool                       `json:"is_admin"`
}

type InviteReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" uri:"lab_uuid" form:"lab_uuid"`
}

type InviteResp struct {
	Path string `json:"url"`
}

type AcceptInviteReq struct {
	UUID uuid.UUID `json:"uuid" uri:"uuid" form:"uuid"`
}

// Swagger-only concrete wrappers to avoid generics in annotations
// LaboratoryListResp wraps a paginated list of laboratories for Swagger docs
type LaboratoryListResp struct {
	HasMore  bool              `json:"has_more"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Data     []*LaboratoryResp `json:"data"`
}

// LabMemberListResp wraps a paginated list of lab members for Swagger docs
type LabMemberListResp struct {
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Data     []*LabMemberResp `json:"data"`
}
