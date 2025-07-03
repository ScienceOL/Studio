package utils

import (
	"fmt"
	"runtime/debug"
)

func SafelyRun(function func()) (err error) {
	// 返回是否成功执行
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("%w\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("unknown panic\n%s", string(debug.Stack()))
			}
		}
	}()

	function()

	return nil
}

func SafelyGo(function func(), handleError func(error)) {
	go func() {
		err := SafelyRun(function)
		if err != nil {
			handleError(err)
		}
	}()
}
