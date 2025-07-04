package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/common/code"
)

type Error struct {
	Msg string `json:"msg"`
}

type Resp struct {
	Code  int    `json:"code"`
	Error *Error `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func ReplyErr(ctx *gin.Context, err error) {
	if errCode, ok := err.(*code.ErrCode); ok {
		ctx.JSON(http.StatusOK, &Resp{
			Code: errCode.Int(),
			Error: &Error{
				Msg: errCode.String(),
			},
		})
	}

	ctx.JSON(http.StatusOK, &Resp{
		Code: code.UnDefineErr.Int(),
		Error: &Error{
			Msg: err.Error(),
		},
	})
}

func ReplyOk(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, &Resp{
		Code: code.Success.Int(),
		Data: data,
	})
}
