package code

import (
	"fmt"
	"strings"
)

// "golang.org/x/tools/cmd/stringer"

//go:generate stringer -type ErrCode -linecomment -output code_string.go
type ErrCode int

type ErrCodeWithMsg struct {
	ErrCode
	msgs []string
}

func (e ErrCodeWithMsg) String() string {
	return fmt.Sprintf("code: %d, msgs: %+v", e.ErrCode, e.msgs)
}

func (e ErrCodeWithMsg) Msgs() string {
	return strings.Join(e.msgs, "\t\t\t")
}

// const (
// 	codeSplit = " &_&_& "
// )

func (e ErrCode) WithMsg(msgs ...string) error {
	return &ErrCodeWithMsg{ErrCode: e, msgs: msgs}
}

func (e ErrCode) Int() int {
	return int(e)
}

func (e ErrCode) Error() string {
	return e.String()
}

const (
	Success     ErrCode = 0 // 成功
	UnDefineErr             // 未定义
)

// view 展示层错误
const (
	ParamErr ErrCode = iota + 1000 // parse parameter error
)

// 登录模块错误
const (
	LoginConfigErr       ErrCode = iota + 5000 // 登录配置错误
	LoginSetStateErr                           // 设置登录状态错误
	RefreshTokenErr                            // 刷新 token 失败
	LoginStateErr                              // state 验证失败
	ExchangeTokenErr                           // 交换 token 失败
	CallbackParamErr                           // 回调参数错误
	LoginGetUserInfoErr                        // 获取用户信息失败
	LoginCallbackErr                           // 登录处理用户信息失败
	UnLogin                                    // 未登录状态
	LoginFormatErr                             // 登录验证格式错误
	InvalidToken                               // 无效 token
	RefreshTokenParamErr                       // 刷新 token 参数错误
)

// 数据库层错误
const (
	CreateDataErr  ErrCode = iota + 10000 // dababase create data err
	UpdateDataErr                         // database update data err
	RecordNotFound                        // database record not found
	QueryRecordErr                        // database query err
)

// environment 业务层错误
const (
	RegActionNameEmptyErr ErrCode = iota + 20000 // reg action name empty
)

// material 业务错误
const (
	RegNotExistErr        ErrCode = iota + 22000 // registry not exist
	EdgeNodeNotExistErr                          // edge node not exist
	EdgeHandleNotExistErr                        // node handle not exist
)
