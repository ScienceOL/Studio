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
	return ErrCodeWithMsg{ErrCode: e, msgs: msgs}
}

func (e ErrCode) WithMsgf(format string, msgs ...any) error {
	return ErrCodeWithMsg{ErrCode: e, msgs: []string{fmt.Sprintf(format, msgs...)}}
}

func (e ErrCode) WithErr(errs ...error) error {
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, e.Error())
	}
	return ErrCodeWithMsg{ErrCode: e, msgs: msgs}
}

func (e ErrCode) Int() int {
	return int(e)
}

func (e ErrCode) Error() string {
	return e.String()
}

const (
	Success     ErrCode = iota // success
	UnDefineErr                // undefined
)

// view layer errors
const (
	ParamErr           ErrCode = iota + 1000 // parse parameter error
	NotPointerErr                            // not pointer err
	NotSlicePointerErr                       // must be a pointer to a slice
)

// login module errors
const (
	LoginConfigErr       ErrCode = iota + 5000 // login configuration error
	LoginSetStateErr                           // set login state error
	RefreshTokenErr                            // refresh token failed
	LoginStateErr                              // state verification failed
	ExchangeTokenErr                           // exchange token failed
	CallbackParamErr                           // callback parameter error
	LoginGetUserInfoErr                        // get user info failed
	LoginCallbackErr                           // login process user info failed
	UnLogin                                    // not logged in
	LoginFormatErr                             // login verification format error
	InvalidToken                               // invalid token
	RefreshTokenParamErr                       // refresh token parameter error
)

// database layer errors
const (
	CreateDataErr              ErrCode = iota + 10000 // database create data error
	UpdateDataErr                                     // database update data error
	RecordNotFound                                    // database record not found
	QueryRecordErr                                    // database query error
	DeleteDataErr                                     // database delete error
	NotBaseDBTypeErr                                  // not base db type error
	ModelNotImplementTablerErr                        // model not implement schema.Tabler
	RedisLuaScriptErr                                 // redis lua script error
	RedisLuaRetErr                                    // redis lua return type error
	RedisAddSetErr                                    // redis add user set error
	RedisRemoveSetErr                                 // redis remove user set error
)

// environment business layer errors
const (
	RegActionNameEmptyErr ErrCode = iota + 20000 // reg action name empty
	ResourceIsEmptyErr                           // resource is empty
	ResourceNotExistErr                          // resource not exist
)

// material module errors
const (
	ResNotExistErr          ErrCode = iota + 22000 // resource not exist
	EdgeNodeNotExistErr                            // edge node not exist
	EdgeHandleNotExistErr                          // node handle not exist
	UnknownWSActionErr                             // unknown material websocket action
	UnmarshalWSDataErr                             // unmarshal material websocket data error
	CanNotGetLabIDErr                              // cannot get lab id error
	UpdateNodeErr                                  // update material node error
	ParentNodeNotFoundErr                          // parent node not found error
	TemplateNodeNotFoundErr                        // template node not found error
	InvalidDagErr                                  // invalid dag error
	MaxTplNodeDeepErr                              // max template node deep error
)

// notify module errors
const (
	NotifyActionAlreadyRegistryErr ErrCode = iota + 24000 // notify action already registry
	NotifySubscribeChannelErr                             // notify subscribe channel fail
	NotifySendMsgErr                                      // notify send message error
)

// rpc casdoor module errors
const (
	CasDoorCreateLabUserErr ErrCode = iota + 26000 // create lab user error
	CasDoorQueryLabUserErr                         // query lab user error
)

// workflow module errors
const (
	CanNotGetWorkflowUUIDErr ErrCode = iota + 28000 // can not get workflow uuid
	WorkflowNotExistErr                             // workflow not exist
	UpsertWorkflowEdgeErr                           // upsert workflow edge error
	PermissionDenied                                // permission denied
	SaveWorkflowNodeErr                             // batch save nodes error
	SaveWorkflowEdgeErr                             // batch save workflow edge error
)

// schedule module errors
const (
	WorkflowTaskAlreadyExistErr ErrCode = iota + 30000 // workflow task already exist error
	CanNotFoundEdgeSession                             // can not found edge session
	WorkflowHasCircularErr                             // workflow has circular error
	EdgeConnectClosedErr                               // connect closed when node running error
	NodeDataMarshalErr                                 // marshal node data error
	JobRetryTimeout                                    // job retry timeout error
	JobRunFailErr                                      // job run fail error
	WorkflowTaskNotFoundErr                            // can not found workflow task error
	WorkflowTaskStatusErr                              // workflow task status error
	WorkflowTaskFinished                               // workflow task finished
)
