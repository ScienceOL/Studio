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

	if len(values) > 0 {
		return values[len(values)-1]
	}

	return zero
}

// TernaryLazy 延迟计算版本的三元运算符，避免预先计算所有参数
// 只有在需要时才会调用相应的函数来计算值
func TernaryLazy[T any](condition bool, trueFn, falseFn func() T) T {
	if condition {
		return trueFn()
	}
	return falseFn()
}

func SafeValue[T any](f func() T, defaultVal T) (res T) {
	defer func() {
		if r := recover(); r != nil {
			res = defaultVal
		}
	}()

	return f()
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

func FilterSliceWithErr[S any, T any](sources []T, f func(i T) ([]S, bool, error)) ([]S, error) {
	newSlice := make([]S, 0, len(sources))
	for _, item := range sources {
		datas, isAdd, err := f(item)
		if err != nil {
			return nil, err
		}
		if isAdd {
			newSlice = append(newSlice, datas...)
		}
	}
	return newSlice, nil
}

func SliceToMap[K comparable, V any, T any](sources []T, f func(i T) (K, V)) map[K]V {
	result := make(map[K]V)
	for _, item := range sources {
		key, value := f(item)
		result[key] = value
	}
	return result
}

func SliceToMapSlice[K comparable, V any, T any](sources []T, f func(i T) (K, V, bool)) map[K][]V {
	result := make(map[K][]V)
	for _, item := range sources {
		if key, value, isAdd := f(item); isAdd {
			result[key] = append(result[key], value)
		}
	}
	return result
}

func MapToSlice[K comparable, V any, T any](sources map[K]V, f func(key K, value V) (T, bool)) []T {
	result := make([]T, 0, len(sources))
	for k, v := range sources {
		data, add := f(k, v)
		if add {
			result = append(result, data)
		}
	}

	return result
}
