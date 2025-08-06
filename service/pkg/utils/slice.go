//nolint:revive // var-naming: common package contains shared utilities
package utils

import "slices"

func AppendUniqSlice[T comparable](slice []T, elem T) []T {
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

func FilterSlice[S any, T any](sources []T, f func(i T) (S, bool)) []S {
	newSlice := make([]S, 0, len(sources))
	for _, item := range sources {
		data, isAdd := f(item)
		if isAdd {
			newSlice = append(newSlice, data)
		}
	}
	return newSlice
}

func SliceToMap[K comparable, V any, T any](sources []T, f func(i T) (K, V)) map[K]V {
	result := make(map[K]V)
	for _, item := range sources {
		key, value := f(item)
		result[key] = value
	}
	return result
}
