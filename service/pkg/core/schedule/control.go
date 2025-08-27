package schedule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	"github.com/panjf2000/ants/v2"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/internal/configs/schedule"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/schedule/device"
	"github.com/scienceol/studio/service/pkg/core/schedule/device/impl"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine/dag"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	"github.com/scienceol/studio/service/pkg/utils"
)

/*
1. edge 实验室 websocket 连接之后双向发送消息，运行指令，上报状态。
2. redis 接受消息。
	a. 工作流启动的队列消息，哪个服务抢到队列的消息，哪个服务运行工作流。
		1. 问题如何解决，两个 schedule pod，edge 连接到 a pod，结果 b 抢到了调度策略，如何把消息发送到 b 让 b 下发任务？
			解决方案，如果 a 抢到了，a 把命令广播出去，b 收到后，直接下发消息。b 收到状态回报消息收，b 修改数据库，广播通知消息，web 服务
			收到后直接发送给 web 侧。
	b. 工作流运行时接收到 websocket 消息上报之后，广播通知所有的客户端终端。

*/

var (
	ctl  *control
	once sync.Once
)

const (
	registryPeriod = 1 * time.Second
	poolSize       = 200
)

type control struct {
	wsClient      *melody.Melody
	scheduleName  string
	deviceManager device.Service
	rClient       *r.Client
	labClient     sync.Map
	tasks         sync.Map
	pools         *ants.Pool
	labStore      repo.LaboratoryRepo
	consumer      *redis.MessageConsumer
}

func NewControl(ctx context.Context) Control {
	once.Do(func() {
		wsClient := melody.New()
		wsClient.Config.MaxMessageSize = constant.MaxMessageSize
		wsClient.Config.PingPeriod = 10 * time.Second

		ctl = &control{
			wsClient:      wsClient,
			scheduleName:  fmt.Sprintf("lab-schedule-name-%s", uuid.NewV4().String()),
			deviceManager: impl.NewDeviceManager(ctx),
			rClient:       redis.GetClient(),
			labClient:     sync.Map{},
			labStore:      eStore.New(),
		}
		ctl.initControl(ctx)
	})

	return ctl
}

func (i *control) Connect(ctx context.Context) {
	// edge 侧用户 websocket 连接
	ginCtx := ctx.(*gin.Context)
	labUser := auth.GetLabUser(ctx)
	lab, err := i.labStore.GetLabByAkSk(ctx, labUser.AccessKey, labUser.AccessSecret)
	if err != nil {
		logger.Errorf(ctx, "can not get lab access key: %s", labUser.AccessKey)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("can not get lab"))
		return
	}

	labInfo := &LabInfo{
		LabUser: labUser,
		LabData: lab,
	}

	if exist, err := i.consumer.HasUser(ctx, lab.UUID.String()); err != nil {
		logger.Errorf(ctx, "connect edge websocket err: %+v", err)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("can not get lab"))
		return
	} else if exist {
		logger.Errorf(ctx, "can not get lab access key: %s", labUser.AccessKey)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("can not get lab"))
		return
	}

	if _, exist := i.labClient.LoadOrStore(lab.UUID, labInfo); exist {
		logger.Errorf(ctx, "can not get lab access key: %s", labUser.AccessKey)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("can not get lab"))
		return
	}

	if err := i.consumer.AddUser(ctx, lab.UUID.String()); err != nil {
		logger.Errorf(ctx, "can not get lab access key: %s", labUser.AccessKey)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("can not get lab"))
		return
	}

	if err := i.wsClient.HandleRequestWithKeys(ginCtx.Writer, ginCtx.Request, map[string]any{
		LABINFO: labInfo,
		"ctx":   ctx,
		"uuid":  lab.UUID,
	}); err != nil {
		logger.Errorf(ctx, "HandleRequestWithKeys fail err: %+v", err)
	}
}

// 接受 edge 发送过来的任务通知消息或者设备状态通知消息
func (i *control) OnEdgeMessge(ctx context.Context, msg string) {
	// FIXME: 待完善
	logger.Infof(ctx, "OnEdgeMessge job msg: %s", msg)
	// 任务状态消息

	// 设备状态消息
}

// 接受 redis queue 发送过来启动任务或者关闭任务的消息
func (i *control) OnJobMessage(ctx context.Context, msg []byte) {
	logger.Infof(ctx, "OnJobMessage job msg: %s", msg)
	action := &engine.WorkflowInfo{}
	if err := json.Unmarshal(msg, action); err != nil {
		logger.Errorf(ctx, "OnJobMessage err: %+v", err)
		return
	}

	switch action.Action {
	case engine.StartJob:
		if err := i.createWrokflowTask(ctx, action); err != nil {
			logger.Errorf(ctx, "createWrokflowTask err: %+v", err)
		}
	case engine.StopJob:
		if err := i.stopWorkflowTask(ctx, action); err != nil {
			logger.Errorf(ctx, "stopWorkflowTask err: %+v", err)
		}
	case engine.StatusJob:
		if err := i.getWorkflowTask(ctx, action); err != nil {
			logger.Errorf(ctx, "getWorkflowTask err: %+v", err)
		}
	default:
		logger.Errorf(ctx, "OnJobMessage action is empty: %+s", msg)
		return
	}
}

func (i *control) createWrokflowTask(ctx context.Context, info *engine.WorkflowInfo) error {
	if info.TaskUUID.IsNil() || info.WorkflowUUID.IsNil() || info.LabUUID.IsNil() {
		return code.ParamErr.WithMsg("task uuid is empty")
	}

	allSession, err := i.wsClient.Sessions()
	if err != nil {
		return code.CanNotFoundEdgeSession
	}

	targetSession := utils.FindTarget(allSession, func(s *melody.Session) (*melody.Session, bool) {
		uuidI, ok := s.Get("uuid")
		if !ok {
			return nil, ok
		}
		if uuidI.(uuid.UUID) == info.LabUUID {
			return s, true
		}
		return nil, false
	})

	if targetSession == nil {
		return code.CanNotFoundEdgeSession
	}

	key := engine.WorkflowTaskKey{
		TaskUUID:     info.TaskUUID,
		WorkflowUUID: info.WorkflowUUID,
	}

	taskCtx, cancle := context.WithCancel(ctx)
	task := dag.NewDagTask(ctx, &engine.TaskParam{
		Devices: i.deviceManager,
		Session: targetSession,
		Cancle:  cancle,
	})

	controlTask := &ControlTask{
		Task:   task,
		Cancle: cancle,
		ctx:    taskCtx,
	}

	_, ok := i.tasks.LoadOrStore(key, controlTask)
	if ok {
		return code.WorkflowTaskAlreadyExistErr.WithMsgf("task info: %+v", *info)
	}

	v, ok := i.labClient.Load(info.LabUUID)
	if !ok {
		return code.EdgeConnectClosedErr.WithMsgf("can not found lab uuid: %+s", info.LabUUID)
	}

	info.LabData = v.(*LabInfo).LabData
	if err := i.pools.Submit(func() {
		if err := task.Run(taskCtx, info); err != nil {
			logger.Errorf(ctx, "task run fail")
		}

		// 停止任务
		controlTask.Cancle()
		// 删除记录
		i.tasks.Delete(key)
	}); err != nil {
		logger.Errorf(ctx, "Submit job fail err: %+v", err)
		return err
	}

	return nil
}

func (i *control) stopWorkflowTask(ctx context.Context, info *engine.WorkflowInfo) error {
	key := engine.WorkflowTaskKey{
		TaskUUID:     info.TaskUUID,
		WorkflowUUID: info.WorkflowUUID,
	}
	value, ok := i.tasks.Load(key)
	if !ok {
		logger.Warnf(ctx, "stopWorkflowTask task not exist key: %+v", key)

		return nil
	}

	task := value.(*ControlTask)
	err := task.Task.Stop(ctx)
	if err != nil {
		logger.Errorf(ctx, "stopWorkflowTask fail err: %+v", err)
	}

	i.tasks.Delete(key)
	return err
}

// nolint: unparam
func (i *control) getWorkflowTask(ctx context.Context, info *engine.WorkflowInfo) error {
	key := engine.WorkflowTaskKey{
		TaskUUID:     info.TaskUUID,
		WorkflowUUID: info.WorkflowUUID,
	}
	value, ok := i.tasks.Load(key)
	if !ok {
		logger.Warnf(ctx, "stopWorkflowTask task not exist key: %+v", key)

		return nil
	}

	// FIXME: 需要获取到什么？
	task := value.(*ControlTask)
	if err := task.Task.GetStatus(ctx); err != nil {
		logger.Errorf(ctx, "stopWorkflowTask fail err: %+v", err)
	}

	return nil
}

func (i *control) initWebSocket(ctx context.Context) {
	i.wsClient.HandleClose(func(s *melody.Session, _ int, _ string) error {
		// 关闭之后的回调
		ctxI, _ := s.Get("ctx")
		gCtx := ctxI.(*gin.Context)
		uuid := s.MustGet("uuid").(uuid.UUID)
		logger.Infof(ctxI.(context.Context), "client close keys: %+v", s.Keys)

		i.labClient.Delete(uuid)
		if err := i.consumer.RemoveUser(gCtx, uuid.String()); err != nil {
			logger.Errorf(ctx, "initWebSocket RemoveUser fail err: %+v", err)
			return err
		}
		return nil
	})

	i.wsClient.HandleDisconnect(func(s *melody.Session) {
		// melody client 端口后的最后一个回调事件
		ctx, _ := s.Get("ctx")
		gCtx := ctx.(*gin.Context)
		uuid := s.MustGet("uuid").(uuid.UUID)
		i.labClient.Delete(uuid)
		if err := i.consumer.RemoveUser(gCtx, uuid.String()); err != nil {
			logger.Errorf(gCtx, "initWebSocket HandleDisconnect fail err: %+v", err)
		}
	})

	i.wsClient.HandleError(func(s *melody.Session, err error) {
		// 读或写或写 buf 满了出错
		if errors.Is(err, melody.ErrMessageBufferFull) {
			return
		}
		if closeErr, ok := err.(*websocket.CloseError); ok {
			if closeErr.Code == websocket.CloseGoingAway {
				return
			}
		}

		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "websocket find keys: %+v, err: %+v", s.Keys, err)
		}
	})

	i.wsClient.HandleConnect(func(s *melody.Session) {
		// melody 开始连接时报错
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "websocket connect keys: %+v", s.Keys)
			// m.mService.OnWSConnect(ctx.(context.Context), s)
		}
	})

	i.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		// client 发送消息回调
		ctxI, ok := s.Get("ctx")
		if !ok {
			if err := s.CloseWithMsg([]byte("no ctx")); err != nil {
				logger.Errorf(ctxI.(context.Context), "initWebSocket HandleMessage fail err: %+v", err)
			}
			return
		}
		i.OnEdgeMessge(ctxI.(context.Context), string(b))
	})

	i.wsClient.HandleSentMessage(func(_ *melody.Session, _ []byte) {
		// 发送完字符串消息后的回调
	})

	i.wsClient.HandleSentMessageBinary(func(_ *melody.Session, _ []byte) {
		// 发送完二进制消息后的回调
	})
}

func (i *control) initControl(ctx context.Context) {
	i.pools, _ = ants.NewPool(poolSize)
	i.consumer = redis.NewMessageConsumer(i.rClient, i.scheduleName)
	i.initWebSocket(ctx)
	i.startConsumeJob(ctx)
}

func (i *control) startConsumeJob(ctx context.Context) {
	conf := schedule.Config().Job
	utils.SafelyGo(func() {
		time.Sleep(5 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(registryPeriod):
				i.consumer.Message(ctx, conf.JobQueueName, func(msg []byte) {
					i.OnJobMessage(ctx, msg)
				})
			}
		}
	}, func(err error) {
		logger.Errorf(ctx, "consumer redis message fail err:%+v", err)
	})
}

func (i *control) Close(ctx context.Context) {
	if i.wsClient != nil {
		if err := i.wsClient.CloseWithMsg([]byte("reboot")); err != nil {
			logger.Errorf(ctx, "Close fail CloseWithMsg err: %+v", err)
		}
	}

	if i.consumer != nil {
		if err := i.consumer.Cleanup(ctx); err != nil {
			logger.Errorf(ctx, "Close fail Cleanup err: %+v", err)
		}
	}

	if i.pools != nil {
		i.pools.Release()
	}
}
