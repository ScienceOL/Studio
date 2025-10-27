package utils

import (
	"encoding/json"

	"github.com/google/go-cmp/cmp"
)

func JSONEqual(a, b string) (bool, error) {
	var objA, objB any
	if err := json.Unmarshal([]byte(a), &objA); err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(b), &objB); err != nil {
		return false, err
	}
	return cmp.Equal(objA, objB), nil
}
