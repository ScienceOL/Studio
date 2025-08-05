//nolint:revive // var-naming: common package contains shared utilities
package common

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common/code"
)

type Error struct {
	Msg  string   `json:"msg"`
	Info []string `json:"info,omitempty"`
}

type Resp struct {
	Code      code.ErrCode `json:"code"`
	Error     *Error       `json:"error,omitempty"`
	Data      any          `json:"data,omitempty"`
	Timestamp int64        `json:"timestamp,omitempty"`
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

	if errCode, ok := err.(code.ErrCodeWithMsg); ok {
		ctx.JSON(http.StatusOK, &Resp{
			Code: errCode.ErrCode,
			Error: &Error{
				Msg:  errCode.Msgs(),
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

func ReplyWSOk(s *melody.Session, data ...any) error {
	if len(data) > 0 {
		data := &Resp{
			Code:      code.Success,
			Data:      data[0],
			Timestamp: time.Now().Unix(),
		}
		v, _ := json.Marshal(data)
		return s.Write(v)
	}

	v, _ := json.Marshal(&Resp{
		Code:      code.Success,
		Timestamp: time.Now().Unix(),
	})
	return s.Write(v)
}

func ReplyWSErr(s *melody.Session, err error, msg ...string) error {
	if errCode, ok := err.(code.ErrCode); ok {
		d := &Resp{
			Code: errCode,
			Error: &Error{
				Msg:  errCode.String(),
				Info: msg,
			},
			Timestamp: time.Now().Unix(),
		}

		b, _ := json.Marshal(d)
		return s.Write(b)
	}

	if errCode, ok := err.(code.ErrCodeWithMsg); ok {
		d := &Resp{
			Code: errCode.ErrCode,
			Error: &Error{
				Msg:  errCode.Msgs(),
				Info: msg,
			},
			Timestamp: time.Now().Unix(),
		}
		b, _ := json.Marshal(d)
		return s.Write(b)
	}

	d := &Resp{
		Code: code.UnDefineErr,
		Error: &Error{
			Msg: err.Error(),
		},
		Timestamp: time.Now().Unix(),
	}
	b, _ := json.Marshal(d)
	return s.Write(b)
}
