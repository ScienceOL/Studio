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
	Success       ErrCode = iota // success
	UnDefineErr                  // undefined
	NoPermission                 // no permission
	InvalidateJWT                // invalidate jwt
)

// view layer errors
const (
	ParamErr           ErrCode = iota + 1000 // parse parameter error
	NotPointerErr                            // not pointer err
	NotSlicePointerErr                       // must be a pointer to a slice
	PointerIsNilErr                          // pointer is nil error
)

// login module errors
const (
	LoginConfigErr           ErrCode = iota + 5000 // login configuration error
	LoginSetStateErr                               // set login state error
	RefreshTokenErr                                // refresh token failed
	LoginStateErr                                  // state verification failed
	ExchangeTokenErr                               // exchange token failed
	CallbackParamErr                               // callback parameter error
	LoginGetUserInfoErr                            // get user info failed
	LoginCallbackErr                               // login process user info failed
	UnLogin                                        // not logged in
	LoginFormatErr                                 // login verification format error
	InvalidToken                                   // invalid token
	RefreshTokenParamErr                           // refresh token parameter error
	ParseLoginRedirectURLErr                       // redirect login url error
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
	RegActionNameEmptyErr       ErrCode = iota + 20000 // reg action name empty
	ResourceIsEmptyErr                                 // resource is empty
	ResourceNotExistErr                                // resource not exist
	WorkflowTemplateNotFoundErr                        // can not found workflow template error
	UserIDIsEmptyErr                                   // user id is empty
	LabIDIsEmptyErr                                    // lab id is empty error
	LabNotFound                                        // laboratory not found error
	LabInviteNotFoundErr                               // can not found laboratory invite link error
	InviteExpiredErr                                   // invite expired error
	InvalidateThirdID                                  // invalidate third id error
	LabAlreadyDeletedErr                               // lab already deleted error
)

// material module errors
const (
	ResNotExistErr             ErrCode = iota + 22000 // resource not exist
	EdgeNodeNotExistErr                               // edge node not exist
	EdgeHandleNotExistErr                             // node handle not exist
	UnknownWSActionErr                                // unknown material websocket action
	UnmarshalWSDataErr                                // unmarshal material websocket data error
	CanNotGetLabIDErr                                 // cannot get lab id error
	UpdateNodeErr                                     // update material node error
	ParentNodeNotFoundErr                             // parent node not found error
	TemplateNodeNotFoundErr                           // template node not found error
	InvalidDagErr                                     // invalid dag error
	MaxTplNodeDeepErr                                 // max template node deep error
	CanNotFoundMaterialNodeErr                        // can not found material node error
	MachineAlreadyExistErr                            // machine already exist error
	QueryMachineStatusFailErr                         // query machine status error
	MachineNotExistErr                                // machine not exist error
	MachineReachMaxNumCountErr                        // machine reach max number error
	MachineNodeStoppingErr                            // machine is stopping
	MachineStartUnknownErr                            // start machine unknown error
	CanNotFoundTargetNode                             // can not found target node error
	PathHasEmptyName                                  // path has empty name error
)

// notify module errors
const (
	NotifyActionAlreadyRegistryErr ErrCode = iota + 24000 // notify action already registry
	NotifySubscribeChannelErr                             // notify subscribe channel fail
	NotifySendMsgErr                                      // notify send message error
)

// rpc account module errors
const (
	RPCHttpErr              ErrCode = iota + 26000 // rpc request http error
	RPCHttpCodeErr                                 // rpc request http code error
	RPCHttpCodeRespErr                             // rpc request http code resp error
	CasDoorCreateLabUserErr                        // create lab user error
	CasDoorQueryLabUserErr                         // query lab user error
	BohrBatchQueryErr                              // bhor batch query user error
)

// workflow module errors
const (
	CanNotGetWorkflowUUIDErr ErrCode = iota + 28000 // can not get workflow uuid
	WorkflowNotExistErr                             // workflow not exist
	UpsertWorkflowEdgeErr                           // upsert workflow edge error
	PermissionDenied                                // permission denied
	SaveWorkflowNodeErr                             // batch save nodes error
	SaveWorkflowEdgeErr                             // batch save workflow edge error
	WorkflowNodeNotFoundErr                         // workflow node not found error
	CanNotGetworkflowErr                            // workflow not found error
	FormatCSVTaskErr                                // format csv data error
)

// schedule module errors
const (
	WorkflowTaskAlreadyExistErr     ErrCode = iota + 30000 // workflow task already exist error
	CanNotFoundEdgeSession                                 // can not found edge session
	WorkflowHasCircularErr                                 // workflow has circular error
	EdgeConnectClosedErr                                   // connect closed when node running error
	NodeDataMarshalErr                                     // marshal node data error
	JobRunFailErr                                          // job run fail error
	WorkflowTaskNotFoundErr                                // can not found workflow task error
	WorkflowTaskStatusErr                                  // workflow task status error
	WorkflowTaskFinished                                   // workflow task finished
	WorkflowNodeNoDeviceName                               // workflow node no device name error
	WorkflowNodeNoActionName                               // workflow node no action name error
	WorkflowNodeNoActionType                               // workflow node no action type error
	QueryJobStatusKeyNotExistErr                           // query job status key note exists error
	CallbackJobStatusKeyNotExistErr                        // callback job status key note exists error
	JobTimeoutErr                                          // job timeout error
	JobRetryTimeout                                        // job retry timeout error
	CallbackJobStatusTimeoutErr                            // callback job status timeout error
	JobCanceled                                            // job is canceled
	CanNotGetWorkflowTaskErr                               // can not get workflow task error
	WorkflowTaskStatusNotPendingErr                        // workflow task not in pending status
	CanNotFoundWorkflowHandleErr                           // can not found workflow handle error
	CanNotGetParentJobErr                                  // can not found parent node job error
	ParamDataKeyInvalidateErr                              // param data key invalidate error
	ParamDataValueInvalidateErr                            // param data value invalidate error
	DataNotMapAnyTypeErr                                   // data not map any type error
	ValueSliceOutIndexErr                                  // value slice out index error
	ValueNotExistErr                                       // value not exist error
	SetLabHeartErr                                         // set lab heart error
	TargetDataNotMapAnyTypeErr                             // target data not map any type error
	MarshalTargetDataErr                                   // marshal target data error
	TargetParamInvalidateErr                               // target param invalidate error
	WorkflowNodeScriptEmtpyErr                             // workflow script empty error
	UnknownWorkflowNodeTypeErr                             // unknown workflow node type error
	ExecWorkflowNodeScriptErr                              // exec workflow script error

)
