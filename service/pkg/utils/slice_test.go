// notlint:revive
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeValue(t *testing.T) {
	assert.Equal(t, SafeValue(func() int {
		panic("test")
	}, 2), 2)

	assert.Equal(t, SafeValue(func() int {
		return 3
	}, 2), 3)
}
