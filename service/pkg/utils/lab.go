package utils

import (
	"fmt"
	"time"

	"github.com/scienceol/studio/service/pkg/common/uuid"
)

const (
	LabTaskPrefix    = "lab_task_queue_%s"
	LabControlPrefix = "lab_control_queue_%s"
	LabHeartPrefix   = "lab_heart_key_%s"

	LabHeartTime = 5 * time.Second
)

func LabTaskName(labUUID uuid.UUID) string {
	return fmt.Sprintf(LabTaskPrefix, labUUID.String())
}

func LabControlName(labUUID uuid.UUID) string {
	return fmt.Sprintf(LabControlPrefix, labUUID.String())
}

func LabHeartName(labUUID uuid.UUID) string {
	return fmt.Sprintf(LabHeartPrefix, labUUID.String())
}
