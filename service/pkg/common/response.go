//nolint:revive // var-naming: common package contains shared utilities
package common

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
)

const (
	MaxPageSize = 50
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

type PageResp[T any] struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Data     T     `json:"data"`
}

type PageReq struct {
	Page     int `json:"page" form:"page" uri:"page"`
	PageSize int `json:"page_size" form:"page_size" uri:"page_size"`
}

func (p *PageReq) Normalize() {
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}

	if p.PageSize <= 0 {
		p.PageSize = 1
	}

	if p.Page <= 0 {
		p.Page = 1
	}
}

func (p *PageReq) AddPage(count int) {
	p.Normalize()
	p.Page += count
}

func (p *PageReq) Offest() int {
	return (p.Page - 1) * p.PageSize
}

type WsMsgType struct {
	Action  string    `json:"action"`
	MsgUUID uuid.UUID `json:"msg_uuid"`
}

type WSData[T any] struct {
	WsMsgType
	Data T `json:"data,omitempty"`
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

func ReplyWSOk(s *melody.Session, action string, msgUUID uuid.UUID, data ...any) error {
	if len(data) > 0 {
		d := &Resp{
			Code: code.Success,
			Data: &WSData[any]{
				WsMsgType: WsMsgType{
					Action:  action,
					MsgUUID: msgUUID,
				},
				Data: data[0],
			},
			Timestamp: time.Now().Unix(),
		}
		v, _ := json.Marshal(d)
		return s.Write(v)
	}

	v, _ := json.Marshal(&Resp{
		Code: code.Success,
		Data: &WSData[any]{
			WsMsgType: WsMsgType{
				Action:  action,
				MsgUUID: msgUUID,
			},
		},
		Timestamp: time.Now().Unix(),
	})
	return s.Write(v)
}

func ReplyWSErr(s *melody.Session, action string, msgUUID uuid.UUID, err error) error {
	if errCode, ok := err.(code.ErrCode); ok {
		d := &Resp{
			Code: errCode,
			Error: &Error{
				Msg: errCode.String(),
			},
			Data: &WSData[any]{
				WsMsgType: WsMsgType{
					Action:  action,
					MsgUUID: msgUUID,
				},
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
				Msg: errCode.Msgs(),
			},
			Data: &WSData[any]{
				WsMsgType: WsMsgType{
					Action:  action,
					MsgUUID: msgUUID,
				},
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
		Data: &WSData[any]{
			WsMsgType: WsMsgType{
				Action:  action,
				MsgUUID: msgUUID,
			},
		},
		Timestamp: time.Now().Unix(),
	}
	b, _ := json.Marshal(d)
	return s.Write(b)
}
