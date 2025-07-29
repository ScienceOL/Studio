package utils

import "slices"

func AppendUniqSlice[T string | int | int64](slice []T, elem T) []T {
	if slices.Contains(slice, elem) {
		return slice
	}
	return append(slice, elem) // 添加新元素
}
