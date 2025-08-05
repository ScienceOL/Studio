//nolint:revive // var-naming: common package contains shared utilities
package utils

import "slices"

func AppendUniqSlice[T string | int | int64](slice []T, elem T) []T {
	if slices.Contains(slice, elem) {
		return slice
	}
	return append(slice, elem) // 添加新元素
}

func Or[T comparable](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}

func RemoveDuplicates[T comparable](slice []T) []T {
	keys := make(map[T]bool)
	result := []T{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
