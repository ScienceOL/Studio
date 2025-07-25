//nolint:revive // var-naming: common package contains shared utilities
package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/common/code"
)

type Error struct {
	Msg  string   `json:"msg"`
	Info []string `json:"info,omitempty"`
}

type Resp struct {
	Code  code.ErrCode `json:"code"`
	Error *Error       `json:"error,omitempty"`
	Data  any          `json:"data,omitempty"`
}

func ReplyErr(ctx *gin.Context, err error, msg ...string) {
	if errCode, ok := err.(code.ErrCode); ok {
		ctx.JSON(http.StatusOK, &Resp{
			Code: errCode,
			Error: &Error{
				Msg:  errCode.String(),
				Info: msg,
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, &Resp{
		Code: code.UnDefineErr,
		Error: &Error{
			Msg: err.Error(),
		},
	})
}

// 禁止 data 直接返回数组，不方便接口拓展
func ReplyOk(ctx *gin.Context, data ...any) {
	if len(data) > 0 {
		ctx.JSON(http.StatusOK, &Resp{
			Code: code.Success,
			Data: data[0],
		})
		return
	}

	ctx.JSON(http.StatusOK, &Resp{
		Code: code.Success,
	})
}
