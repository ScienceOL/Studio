//nolint:revive // var-naming: common package contains shared utilities
package common

import (
	"testing"

	"gorm.io/datatypes"
)

func TestBinUUID(t *testing.T) {
	uuid := "56d92f9fa1c54ac284a0a753f0fbf65b"
	_ = uuid
	uuid1 := "56d92f9f-a1c5-4ac2-84a0-a753f0fbf65b"
	uuidT := datatypes.BinUUIDFromString(uuid1)
	t.Log(uuidT)
}
