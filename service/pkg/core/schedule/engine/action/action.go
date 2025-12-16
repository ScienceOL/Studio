package action

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/olahol/melody"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/schedule"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/utils"
)

type stepFunc func(ctx context.Context) error

type actionEngine struct {
	job     *engine.WorkflowInfo
	cancel  context.CancelFunc
	ctx     context.Context
	session *melody.Session
	data    *RunActionReq
	ret     *RunActionResp

	wg        sync.WaitGroup
	stepFuncs []stepFunc

	boardEvent notify.MsgCenter

	actionStatus sync.Map
	rClient      *r.Client
	sanbox       repo.Sandbox
}

func NewActionTask(ctx context.Context, param *engine.TaskParam) engine.Task {
	d := &actionEngine{
		session:    param.Session,
		cancel:     param.Cancle,
		ctx:        ctx,
		wg:         sync.WaitGroup{},
		rClient:    redis.GetClient(),
		sanbox:     param.Sandbox,
		boardEvent: param.BoardEvent,
	}
	d.stepFuncs = append(d.stepFuncs,
		d.loadData, // 加载运行数据
	)

	return d
}

func (d *actionEngine) loadData(ctx context.Context) error {
	paramKey := ActionKey(d.job.TaskUUID)
	paramRet := d.rClient.Get(ctx, paramKey)
	if paramRet.Err() != nil {
		logger.Errorf(ctx, "actionEngine loadData err: %+v", paramRet.Err())
		return paramRet.Err()
	}

	data := &RunActionReq{}
	err := json.Unmarshal([]byte(paramRet.Val()), data)
	if err != nil {
		logger.Errorf(ctx, "actionEngine loadData Unmarshal err: %+v", err)
		return code.ParamErr.WithErr(err)
	}

	if data.UUID.IsNil() ||
		data.LabUUID.IsNil() ||
		data.Action == "" ||
		data.ActionType == "" ||
		data.DeviceID == "" {
		logger.Errorf(ctx, "actionEngine loadData pararm err: %+v", err)
		return code.ParamErr
	}

	d.data = data
	return nil
}

// 运行入口
func (d *actionEngine) Run(ctx context.Context, job *engine.WorkflowInfo) error {
	d.job = job
	var err error
	defer func() {
		d.setActionRet(ctx)
	}()

	for _, s := range d.stepFuncs {
		if err = s(ctx); err != nil {
			break
		}
	}
	err = d.runNode(ctx)
	return err
}

func (d *actionEngine) Stop(_ context.Context) error {
	return nil
}

func (d *actionEngine) runNode(ctx context.Context) error {
	// 查询 action 是否可以执行
	err := d.queryAction(ctx)
	if err != nil {
		return err
	}

	err = d.sendAction(ctx)
	if err != nil {
		return err
	}

	key := engine.ActionKey{
		Type:       engine.JobCallbackStatus,
		TaskID:     d.job.TaskUUID,
		JobID:      d.job.TaskUUID,
		DeviceID:   d.data.DeviceID,
		ActionName: d.data.Action,
	}

	d.InitDeviceActionStatus(ctx, key, time.Now().Add(20*time.Second), false)
	err = d.callbackAction(ctx, key)

	return err
}

func (d *actionEngine) queryAction(ctx context.Context) error {
	key := engine.ActionKey{
		Type:   engine.QueryActionStatus,
		TaskID: d.job.TaskUUID,
		JobID:  d.job.TaskUUID,
		DeviceID: utils.SafeValue(func() string {
			return d.data.DeviceID
		}, ""),
		ActionName: d.data.Action,
	}
	d.InitDeviceActionStatus(ctx, key, time.Now().Add(time.Second*20), false)
	if err := d.sendQueryAction(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return code.JobCanceled
		default:
		}

		time.Sleep(time.Millisecond * 500)
		value, exist := d.GetDeviceActionStatus(ctx, key)
		if !exist {
			return code.QueryJobStatusKeyNotExistErr
		}

		if value.Free {
			d.DelStatus(ctx, key)
			return nil
		}

		if value.Timestamp.Unix() < time.Now().Unix() {
			return code.JobTimeoutErr
		}
	}
}

func (d *actionEngine) sendQueryAction(_ context.Context) error {
	if d.session.IsClosed() {
		return code.EdgeConnectClosedErr
	}

	data := schedule.SendAction[engine.ActionKey]{
		Action: schedule.QueryActionStatus,
		Data: engine.ActionKey{
			TaskID:     d.job.TaskUUID,
			JobID:      d.job.TaskUUID,
			DeviceID:   d.data.DeviceID,
			ActionName: d.data.Action,
		},
	}

	bData, _ := json.Marshal(data)
	return d.session.Write(bData)
}

func (d *actionEngine) sendAction(_ context.Context) error {
	if d.session.IsClosed() {
		return code.EdgeConnectClosedErr
	}

	data := schedule.SendAction[*engine.SendActionData]{
		Action: schedule.JobStart,
		Data: &engine.SendActionData{
			DeviceID:   d.data.DeviceID,
			Action:     d.data.Action,
			ActionType: d.data.ActionType,
			ActionArgs: d.data.Param,
			JobID:      d.job.TaskUUID,
			TaskID:     d.job.TaskUUID,
			NodeID:     d.job.TaskUUID,
			ServerInfo: engine.ServerInfo{
				SendTimestamp: float64(time.Now().UnixNano()) / 1e9,
			},
		},
	}

	b, err := json.Marshal(data)
	if err != nil {
		return code.NodeDataMarshalErr.WithErr(err)
	}

	return d.session.Write(b)
}

func (d *actionEngine) callbackAction(ctx context.Context, key engine.ActionKey) error {
	for {
		select {
		case <-ctx.Done():
			return code.JobCanceled
		default:
		}

		time.Sleep(time.Millisecond * 500)
		value, exist := d.GetDeviceActionStatus(ctx, key)
		if !exist {
			return code.CallbackJobStatusKeyNotExistErr
		}

		if value.Free {
			d.DelStatus(ctx, key)
			break
		}

		if value.Timestamp.Unix() < time.Now().Unix() {
			return code.JobTimeoutErr
		}
	}

	// 查询任务状态是否回调成功
	if d.ret == nil {
		return code.JobTimeoutErr
	}

	logger.Infof(ctx, "schedule action job run finished: %d", d.job.TaskID)
	switch d.ret.Status {
	case string(model.WorkflowJobSuccess):
		return nil
	case string(model.WorkflowJobFailed):
		return code.JobRunFailErr
	default:
		return code.JobRunFailErr
	}
}

func (d *actionEngine) GetStatus(_ context.Context) error {
	return nil
}

func (d *actionEngine) OnJobUpdate(ctx context.Context, data *engine.JobData) error {
	// 广播状态更新（包括 running 状态）
	d.boardMsg(ctx, data)

	if data.Status == "running" {
		return nil
	}

	d.SetDeviceActionStatus(ctx, engine.ActionKey{
		Type:       engine.JobCallbackStatus,
		TaskID:     data.TaskID,
		JobID:      data.JobID,
		DeviceID:   data.DeviceID,
		ActionName: data.ActionName,
	}, true, 0)

	d.ret = &RunActionResp{
		JobData: data,
	}

	return nil
}

func (d *actionEngine) setActionRet(ctx context.Context) {
	retKey := ActionRetKey(d.job.TaskUUID)
	value := &RunActionResp{}
	if d.ret == nil {
		if d.job != nil {
			value.Status = "fail"
			value.JobID = d.job.TaskUUID
			value.TaskID = d.job.TaskUUID
			value.Status = "fail"
		} else {
			value.Status = "fail"
		}
	} else {
		value = d.ret
	}

	b, _ := json.Marshal(value)
	ret := d.rClient.SetEx(ctx, retKey, b, 1*time.Hour)
	if ret.Err() != nil {
		logger.Errorf(ctx, "setActionRet err: %+v", ret.Err())
	}
}

func (d *actionEngine) GetDeviceActionStatus(ctx context.Context, key engine.ActionKey) (engine.ActionValue, bool) {
	valueI, ok := d.actionStatus.Load(key)
	if !ok {
		return engine.ActionValue{}, false
	}
	return valueI.(engine.ActionValue), true
}

func (d *actionEngine) SetDeviceActionStatus(ctx context.Context, key engine.ActionKey, free bool, needMore time.Duration) {
	valueI, ok := d.actionStatus.Load(key)
	if ok {
		value := valueI.(engine.ActionValue)
		value.Free = free
		value.Timestamp = value.Timestamp.Add(needMore)
		logger.Warnf(ctx, "SetDeviceActionStatus key: %+v, value: %+v, more: %d", key, value, needMore)
		d.actionStatus.Store(key, value)
	} else {
		logger.Warnf(ctx, "SetDeviceActionStatus not found key: %+v", key)
	}
}

func (d *actionEngine) InitDeviceActionStatus(ctx context.Context, key engine.ActionKey, start time.Time, free bool) {
	d.actionStatus.Store(key, engine.ActionValue{
		Timestamp: start,
		Free:      free,
	})
}

func (d *actionEngine) DelStatus(ctx context.Context, key engine.ActionKey) {
	d.actionStatus.Delete(key)
}

func (d *actionEngine) boardMsg(ctx context.Context, jobData *engine.JobData) {
	if d.boardEvent == nil {
		return
	}

	if err := d.boardEvent.Broadcast(context.Background(), &notify.SendMsg{
		Channel:      "action-run",
		TaskUUID:     d.job.TaskUUID,
		LabUUID:      d.job.LabUUID,
		WorkflowUUID: d.job.WorkflowUUID,
		UserID:       d.job.UserID,
		UUID:         d.job.TaskUUID,
		Data:         jobData,
		Timestamp:    time.Now().Unix(),
	}); err != nil {
		logger.Errorf(ctx, "action board msg fail err: %+v", err)
	}
}

func (d *actionEngine) ID(ctx context.Context) uuid.UUID {
	if d.job == nil {
		return uuid.NewNil()
	}

	return d.job.TaskUUID
}
