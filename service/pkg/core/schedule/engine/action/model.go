package action

import (
	"fmt"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"gorm.io/datatypes"
)

const (
	ActionKeyPrefix = "workflow_action"
)

func ActionKey(uuid uuid.UUID) string {
	return fmt.Sprintf("%s:%s", ActionKeyPrefix, uuid)
}

func ActionRetKey(uuid uuid.UUID) string {
	return fmt.Sprintf("%s:res:%s", ActionKeyPrefix, uuid)
}

type RunActionReq struct {
	LabUUID    uuid.UUID      `json:"lab_uuid" binding:"required"`
	DeviceID   string         `json:"device_id" binding:"required"`
	Action     string         `json:"action" binding:"required"`
	ActionType string         `json:"action_type" binding:"required"`
	Param      datatypes.JSON `json:"param"`
	UUID       uuid.UUID      `json:"uuid"`
}

type RunActionResp struct {
	*engine.JobData
}
