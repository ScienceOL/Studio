package model

import (
	"time"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"gorm.io/datatypes"
)

type Workflow struct {
	BaseModel
	UserID      string                      `gorm:"type:varchar(120);not null;index:idx_workflow_lu,priority:2" json:"user_id"`
	LabID       int64                       `gorm:"type:bigint;not null;index:idx_workflow_lu,priority:1" json:"lab_id"`
	Name        string                      `gorm:"type:text;not null;default:'Untitled'" json:"name"`
	Published   bool                        `gorm:"type:bool;not null;default:false" json:"published"`
	Tags        datatypes.JSONSlice[string] `gorm:"type:jsonb" json:"tags"`
	Description *string                     `gorm:"type:text" json:"description"`
}

func (*Workflow) TableName() string {
	return "workflow"
}

type WorkflowNodeType string

const (
	WorkflowNodeGroup WorkflowNodeType = "Group"
	WorkflowNodeILab  WorkflowNodeType = "ILab"
)

type WorkflowNode struct {
	BaseModel
	WorkflowID     int64                    `gorm:"type:bigint;not null;index:idx_workflow_id" json:"workflow_id"` // 工作流 id
	WorkflowNodeID int64                    `gorm:"type:bigint;not null" json:"workflow_node_id"`                  // 模板 id
	ParentID       int64                    `gorm:"type:bigint;not null" json:"parent_id"`
	Name           string                   `gorm:"type:varchar(200);not null;default:'unknow'" json:"name"`
	UserID         string                   `gorm:"type:varchar(120);not null" json:"user_id"`
	Status         string                   `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	Type           WorkflowNodeType         `gorm:"type:varchar(20);not null" json:"type"`
	LabNodeType    string                   `gorm:"type:varchar(20);not null;default:'Device'" json:"lab_node_type"` // 节点类型，默认DEFAULT
	Icon           string                   `gorm:"type:text" json:"icon"`
	Pose           datatypes.JSONType[Pose] `gorm:"type:jsonb" json:"pose"`
	Param          datatypes.JSON           `gorm:"type:jsonb" json:"param"`
	Footer         string                   `gorm:"type:text" json:"footer"`
	DeviceName     *string                  `gorm:"type:varchar(255)" json:"device_name"`
	ActionName     string                   `gorm:"type:varchar(255)" json:"action_name"`
	ActionType     string                   `gorm:"type:text" json:"action_type"`
	Disabled       bool                     `gorm:"type:bool;not null;default:false" json:"disabled"`
	Minimized      bool                     `gorm:"type:bool;not null;default:false" json:"minimized"`

	OldNode *WorkflowNode `gorm:"-"` // 复制的节点
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

type WorkflowJobStatus string

const (
	WorkflowJobDraft   WorkflowJobStatus = "draft"
	WorkflowJobSkipped WorkflowJobStatus = "skipped"
	WorkflowJobSuccess WorkflowJobStatus = "success"
	WorkflowJobFailed  WorkflowJobStatus = "failed"
	WorkflowJobRunning WorkflowJobStatus = "running"
	WorkflowJobPending WorkflowJobStatus = "pending"
)

type WorkflowNodeJob struct {
	BaseModel
	LabID          int64             `gorm:"type:bigint;not null;uniqueIndex:idx_workflownodejob_lwn,priority:1" json:"lab_id"`
	WorkflowTaskID int64             `gorm:"type:bigint;not null;uniqueIndex:idx_workflownodejob_lwn,priority:2" json:"workflow_task_id"`
	NodeID         int64             `gorm:"type:bigint;not null;uniqueIndex:idx_workflownodejob_lwn,priority:3" json:"node_id"`
	Status         WorkflowJobStatus `gorm:"type:varchar(50);not null" json:"status"`
	Data           datatypes.JSON    `gorm:"type:jsonb" json:"data"`
}

func (*WorkflowNodeJob) TableName() string {
	return "workflow_node_job"
}

type WorkflowTaskStatus string

const (
	WorkflowTaskStatusPending   WorkflowTaskStatus = "pending"
	WorkflowTaskStatusRunnig    WorkflowTaskStatus = "running"
	WorkflowTaskStatusStoped    WorkflowTaskStatus = "stoped"
	WorkflowTaskStatusFiled     WorkflowTaskStatus = "failed"
	WorkflowTaskStatusSuccessed WorkflowTaskStatus = "successed"
)

type WorkflowTask struct {
	BaseModel
	LabID        int64              `gorm:"type:bigint;not null;index:idx_workflowtask_lwu,priority:1" json:"lab_id"`
	WorkflowID   int64              `gorm:"type:bigint;not null;index:idx_workflowtask_lwu,priority:2" json:"workflow_id"`
	UserID       string             `gorm:"type:varchar(120);not null;index:idx_workflowtask_lwu,priority:3" json:"user_id"`
	Status       WorkflowTaskStatus `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	FinishedTime time.Time          `gorm:"column:finished_time" json:"finished_at"`
}

func (*WorkflowTask) TableName() string {
	return "workflow_task"
}
