package model

import (
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"gorm.io/datatypes"
)

type Workflow struct {
	BaseModel
	UserID      string  `gorm:"type:varchar(120);not null" json:"user_id"`
	LabID       int64   `gorm:"type:bigint;not null" json:"lab_id"`
	Name        string  `gorm:"type:text;not null;default:'Untitled'" json:"name"`
	Description *string `gorm:"type:text" json:"description"`
}

func (*Workflow) TableName() string {
	return "workflow"
}

type WorkflowNode struct {
	BaseModel
	WorkflowID int64                    `gorm:"type:bigint;not null;index:idx_workflow_id" json:"workflow_id"`
	ActionID   int64                    `gorm:"type:bigint;not null" json:"action_id"`
	ParentID   int64                    `gorm:"type:bigint;not null" json:"parent_id"`
	UserID     string                   `gorm:"type:varchar(120);not null" json:"user_id"`
	Status     string                   `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	Type       string                   `gorm:"type:varchar(20);not null" json:"type"`
	Icon       string                   `gorm:"type:text" json:"icon"`
	Pose       datatypes.JSONType[Pose] `gorm:"type:jsonb" json:"pose"`
	Param      datatypes.JSON           `gorm:"type:jsonb" json:"param"`
	Footer     string                   `gorm:"type:text" json:"footer"`
}

func (*WorkflowNode) TableName() string {
	return "workflow_node"
}

type WorkflowConsole struct {
	BaseModel
}

func (*WorkflowConsole) TableName() string {
	return "workflow_console"
}

type WorkflowEdge struct {
	BaseModel
	SourceNodeUUID   uuid.UUID `gorm:"type:uuid;not null;index:idx_we_source_node;uniqueIndex:idx_we_stst,priority:1" json:"source_node_uuid"`
	TargetNodeUUID   uuid.UUID `gorm:"type:uuid;not null;index:idx_we_target_node;uniqueIndex:idx_we_stst,priority:2" json:"target_node_uuid"`
	SourceHandleUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_we_stst,priority:3" json:"source_handle_uuid"`
	TargetHandleUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_we_stst,priority:4" json:"target_handle_uuid"`
}

func (*WorkflowEdge) TableName() string {
	return "workflow_edge"
}
