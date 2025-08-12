package utils

import "reflect"

func Compare(a, b any) bool {
	// 如果都是 nil
	if a == nil && b == nil {
		return true
	}

	// 如果其中一个是 nil
	if a == nil || b == nil {
		return false
	}

	// 获取类型信息
	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)

	// 类型不同直接返回 false
	if aType != bType {
		return false
	}

	// 检查类型是否可比较
	if aType.Comparable() {
		return a == b
	}

	// 不可比较类型使用 DeepEqual
	return reflect.DeepEqual(a, b)
}
