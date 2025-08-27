package model

import (
	"testing"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEnvironment(t *testing.T) {
	e := &Laboratory{}
	e.ID = 1
	tmpUUID := uuid.NewV4()
	e.UUID = tmpUUID
	ei := any(e)

	switch m := ei.(type) {
	case BaseDBModel:
		assert.Equal(t, int64(1), m.GetID())
		assert.Equal(t, tmpUUID, m.GetUUID())

	default:
		t.Error("type err")
	}
}
