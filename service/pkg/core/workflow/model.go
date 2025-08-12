package workflow

import "github.com/gofrs/uuid/v5"

type LabWorkflow struct {
	UUID uuid.UUID `json:"uuid" uri:"uuid" form:"uuid"`
}
