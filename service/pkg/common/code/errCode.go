package code

// "golang.org/x/tools/cmd/stringer"

//go:generate stringer -type ErrCode -linecomment -output code_string.go
type ErrCode int

const (
	codeSplit = " &_&_& "
)

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

// 登录模块错误
const (
	LoginConfigErr   ErrCode = iota + 1000 // 登录配置错误
	LoginSetStateErr                       // 设置登录状态错误
	RefreshTokenErr                        // 刷新 token 失败
	LoginStateErr                          // state 验证失败
	ExchangeTokenErr                       // 交换 token 失败
	CallbackParamErr                       // 回调参数错误
)
