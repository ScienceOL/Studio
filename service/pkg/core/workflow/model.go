package workflow

import "github.com/scienceol/studio/service/pkg/common/uuid"

type LabWorkflow struct {
	UUID uuid.UUID `json:"uuid" uri:"uuid" form:"uuid"`
}
