package workflow

import (
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
)

type LabWorkflow struct {
	UUID uuid.UUID `json:"uuid" uri:"uuid" form:"uuid"`
}

type WorkflowReq struct {
	Name        string    `json:"name"`
	LabUUID     uuid.UUID `json:"lab_uuid" binding:"required"`
	Description *string   `json:"description,omitempty"`
}

type WorkflowResp struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}

type TemplateHandle struct {
	HandleKey string `json:"handle_key"`
	IoType    string `json:"io_type"`
}

type TemplateNodeResp struct {
	UUID            uuid.UUID         `json:"uuid"`
	Name            string            `json:"name"`
	UserID          string            `json:"user_id"`
	Type            string            `json:"type"`
	Icon            string            `json:"icon"`
	TemplateHandles []*TemplateHandle `json:"template_handles"`
}

type TplPageReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" binding:"required"`
	common.PageReq
}

// 模板列表响应
type TemplateListResp struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`         // 模板名称（从device_action name字段取）
	LabName     string    `json:"lab_name"`     // 实验室名字
	HandleCount int       `json:"handle_count"` // handle数量
	CreatedAt   string    `json:"created_at"`   // 创建时间
}

// 节点详情响应
type NodeTemplateDetailResp struct {
	UUID        uuid.UUID      `json:"uuid"`
	Name        string         `json:"name"`         // 模板名称
	Class       string         `json:"class"`        // 类名
	Type        string         `json:"type"`         // 类型
	Icon        string         `json:"icon"`         // 图标
	Schema      datatypes.JSON `json:"schema"`       // 数据模式
	Goal        datatypes.JSON `json:"goal"`         // 目标参数
	GoalDefault datatypes.JSON `json:"goal_default"` // 默认目标参数
	Feedback    datatypes.JSON `json:"feedback"`     // 反馈参数
	Result      datatypes.JSON `json:"result"`       // 结果参数
	LabName     string         `json:"lab_name"`     // 实验室名称
	CreatedAt   string         `json:"created_at"`   // 创建时间
	Handles     []*NodeHandle  `json:"handles"`      // handle列表
}

// 节点Handle信息
type NodeHandle struct {
	UUID        uuid.UUID `json:"uuid"`
	HandleKey   string    `json:"handle_key"`   // handle键
	IoType      string    `json:"io_type"`      // 输入输出类型
	DisplayName string    `json:"display_name"` // 显示名称
	Type        string    `json:"type"`         // 数据类型
	DataSource  string    `json:"data_source"`  // 数据源
	DataKey     string    `json:"data_key"`     // 数据键
}

// ======================================websocket============================
type ActionType string

const (
	FetchGraph        ActionType = "fetch_graph"
	FetchTemplate     ActionType = "fetch_template"
	FetchDevice       ActionType = "fetch_device"
	CreateNode        ActionType = "create_node"
	CreateGroup       ActionType = "create_group"
	UpdateNode        ActionType = "update_node"
	BatchDelNode ActionType = "batch_del_nodes"
	BatchCreateEdge   ActionType = "batch_create_edges"
	BatchDelEdge      ActionType = "batch_del_edges"
	SaveWorkflow      ActionType = "save_workflow"
	RunWorkflow       ActionType = "run_workflow"
	WorkflowUpdate    ActionType = "workflow_update"
)

type WSNodeHandle struct {
	UUID        uuid.UUID `json:"uuid"`
	HandleKey   string    `json:"handle_key"`
	IoType      string    `json:"io_type"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	DataSource  string    `json:"data_source"`
	DataKey     string    `json:"data_key"`
}

type WSGroup struct {
	UUID     uuid.UUID                      `json:"uuid"`
	Children []uuid.UUID                    `json:"children"`
	Pose     datatypes.JSONType[model.Pose] `json:"pose"`
}

type WSGroupRes struct {
	UUID     uuid.UUID   `json:"uuid"`
	Type     string      `json:"type"`
	Children []uuid.UUID `json:"children"`
}

type WSNode struct {
	UUID         uuid.UUID                      `json:"uuid"`
	Name         string                         `json:"name"`
	TemplateUUID uuid.UUID                      `json:"template_uuid"`
	ParentUUID   uuid.UUID                      `json:"parent_uuid"`
	UserID       string                         `json:"user_id"`
	Status       string                         `json:"status"`
	Type         model.WorkflowNodeType         `json:"type"`
	Icon         string                         `json:"icon"`
	Pose         datatypes.JSONType[model.Pose] `json:"pose"`
	Param        datatypes.JSON                 `json:"param"`
	Schema       datatypes.JSON                 `json:"schema"`
	Handles      []*WSNodeHandle                `json:"handles"`
	Footer       string                         `json:"footer"`
	DeviceName   *string                        `json:"device_name,omitempty"`
	Disabled     bool                           `json:"disabled"`
	Minimized    bool                           `json:"minimized"`
	LabNodeType  string                         `json:"lab_node_type"`
}

type WSWorkflowEdge struct {
	UUID             uuid.UUID `json:"uuid"`
	SourceNodeUUID   uuid.UUID `json:"source_node_uuid"`
	TargetNodeUUID   uuid.UUID `json:"target_node_uuid"`
	SourceHandleUUID uuid.UUID `json:"source_handle_uuid"`
	TargetHandleUUID uuid.UUID `json:"target_handle_uuid"`
}

type WSGraph struct {
	Nodes []*WSNode         `json:"nodes"`
	Edges []*WSWorkflowEdge `json:"edges"`
}

type WSTemplate struct {
	UUID          uuid.UUID      `json:"uuid"`
	Name          string         `json:"name"`
	DisplayName   string         `json:"display_name"`
	Header        string         `json:"header"`
	Footer        *string        `json:"footer"`
	Schema        datatypes.JSON `json:"schema"`
	ExecuteScript string         `json:"execute_script"`
	NodeType      string         `json:"node_type"`
	Icon          string         `json:"icon"`
}

type WSTemplateHandles struct {
	Template *WSTemplate     `json:"template"`
	Handles  []*WSNodeHandle `json:"handles"`
}

type WSNodeTpl struct {
	Name            string               `json:"name"`
	UUID            uuid.UUID            `json:"uuid"`
	HandleTemplates []*WSTemplateHandles `json:"handles"`
}

type WSTemplates struct {
	Templates []*WSNodeTpl `json:"templates"`
}

type WSCreateNode struct {
	TemplateUUID uuid.UUID                      `json:"template_uuid"`
	ParentUUID   uuid.UUID                      `json:"parent_uuid"`
	Type         model.WorkflowNodeType         `json:"type"`
	Icon         string                         `json:"icon"`
	Pose         datatypes.JSONType[model.Pose] `json:"pose"`
	Param        *datatypes.JSON                `json:"param,omitempty"`
	Footer       string                         `json:"footer"`
	Name         string                         `json:"name"`
}

type WSUpdateNode struct {
	UUID       uuid.UUID                       `json:"uuid"`
	ParentUUID *uuid.UUID                      `json:"parent_uuid,omitempty"`
	Status     *string                         `json:"status,omitempty"`
	Type       *model.WorkflowNodeType         `json:"type,omitempty"`
	Icon       *string                         `json:"icon,omitempty"`
	Pose       *datatypes.JSONType[model.Pose] `json:"pose,omitempty"`
	Param      *datatypes.JSON                 `json:"param,omitempty"`
	Footer     *string                         `json:"footer,omitempty"`
	Name       *string                         `json:"name,omitempty"`
	Disabled   *bool                           `json:"disabled,omitempty"`
	Minimized  *bool                           `json:"minimized,omitempty"`
	DeviceName *string                         `json:"device_name,omitempty"`
}

type WSDelNodes struct {
	NodeUUIDs []uuid.UUID `json:"node_uuids"`
	EdgeUUIDs []uuid.UUID `json:"edge_uuids"`
}

// 工作流列表请求
type WorkflowListReq struct {
	LabUUID uuid.UUID `json:"lab_uuid" form:"lab_uuid"`
	common.PageReq
}

// 工作流列表响应
type WorkflowListResp struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	UserID      string    `json:"user_id"`
}

// 工作流列表返回（滚动加载）
type WorkflowListResult struct {
	HasMore bool                `json:"has_more"`
	Data    []*WorkflowListResp `json:"data"`
}

// 工作流详情响应
type WorkflowDetailResp struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	UserID      string    `json:"user_id"`
}

// 全局保存节点
