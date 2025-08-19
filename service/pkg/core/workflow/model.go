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

// ======================================websocket============================
type ActionType string

const (
	FetchGraph      ActionType = "fetch_graph"
	FetchTemplate   ActionType = "fetch_template"
	CreateNode      ActionType = "create_node"
	UpdateNode      ActionType = "update_node"
	BatchDelNode    ActionType = "batch_del_nodes"
	BatchCreateEdge ActionType = "batch_create_edges"
	BatchDelEdge    ActionType = "batch_del_edges"
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

type WSNode struct {
	UUID         uuid.UUID                      `json:"uuid"`
	Name         string                         `json:"name"`
	TemplateUUID uuid.UUID                      `json:"template_uuid"`
	ParentUUID   uuid.UUID                      `json:"parent_uuid"`
	UserID       string                         `json:"user_id"`
	Status       string                         `json:"status"`
	Type         string                         `json:"type"`
	Icon         string                         `json:"icon"`
	Pose         datatypes.JSONType[model.Pose] `json:"pose"`
	Param        datatypes.JSON                 `json:"param"`
	Schema       datatypes.JSON                 `json:"schema"`
	Handles      []*WSNodeHandle                `json:"handles"`
	Footer       string                         `json:"footer"`
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
	Type         string                         `json:"type"`
	Icon         string                         `json:"icon"`
	Pose         datatypes.JSONType[model.Pose] `json:"pose"`
	Param        datatypes.JSON                 `json:"param"`
	Footer       string                         `json:"footer"`
	Name         string                         `json:"name"`
}

type WSUpdateNode struct {
	UUID       uuid.UUID                       `json:"uuid"`
	ParentUUID *uuid.UUID                      `json:"parent_uuid"`
	Status     *string                         `json:"status"`
	Type       *string                         `json:"type"`
	Icon       *string                         `json:"icon"`
	Pose       *datatypes.JSONType[model.Pose] `json:"pose"`
	Param      *datatypes.JSON                 `json:"param"`
	Footer     *string                         `json:"footer"`
	Name       *string                         `json:"name"`
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
