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
