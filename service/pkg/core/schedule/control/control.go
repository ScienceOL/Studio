package control

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
	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/material"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/core/schedule"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine/action"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine/dag"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	mStore "github.com/scienceol/studio/service/pkg/repo/material"
	s "github.com/scienceol/studio/service/pkg/repo/sandbox"
	wfl "github.com/scienceol/studio/service/pkg/repo/workflow"
	"github.com/scienceol/studio/service/pkg/utils"
)

var (
	ctl  *control
	once sync.Once
)

const (
	registryPeriod = 1 * time.Second
	poolSize       = 200
)

type control struct {
	wsClient     *melody.Melody
	scheduleName string
	// deviceManager device.Service
	rClient    *r.Client
	labClient  sync.Map               // 实验室端 map
	tasks      sync.Map               // 任务 map
	pools      *ants.Pool             // 任务池
	consumer   *redis.MessageConsumer // 任务队列消费
	boardEvent notify.MsgCenter       // 广播系统
	sandbox    repo.Sandbox           // 脚本运行沙箱

	labStore      repo.LaboratoryRepo // 实验室存储
	workflowStore repo.WorkflowRepo   // 工作流存储
	materialStore repo.MaterialRepo   // 物料调度
}

func NewControl(ctx context.Context) schedule.Control {
	once.Do(func() {
		wsClient := melody.New()
		wsClient.Config.MaxMessageSize = constant.MaxMessageSize
		wsClient.Config.PingPeriod = 10 * time.Second
		scheduleName := fmt.Sprintf("lab-schedule-name-%s", uuid.NewV4().String())
		logger.Infof(ctx, "====================schedule name: %s ======================", scheduleName)

		ctl = &control{
			wsClient:     wsClient,
			scheduleName: scheduleName,
			// deviceManager: impl.NewDeviceManager(ctx),
			rClient:       redis.GetClient(),
			labClient:     sync.Map{},
			labStore:      eStore.New(),
			workflowStore: wfl.New(),
			materialStore: mStore.NewMaterialImpl(),
			boardEvent:    events.NewEvents(),
			sandbox:       s.NewSandbox(),
		}
		ctl.initControl(ctx)
	})

	return ctl
}

// edge 连接 websocket，第一时间接收到连接消息
func (i *control) Connect(ctx context.Context) {
	// edge 侧用户 websocket 连接
	ginCtx := ctx.(*gin.Context)
	// 使用 Lab 鉴权（AK/SK），而不是普通用户鉴权
	labUser := auth.GetLabUser(ctx)
	if labUser == nil || labUser.AccessKey == "" || labUser.AccessSecret == "" {
		logger.Warnf(ctx, "schedule control missing lab user or ak/sk")
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("invalid ak/sk"))
		return
	}
	lab, err := i.labStore.GetLabByAkSk(ctx, labUser.AccessKey, labUser.AccessSecret)
	if err != nil {
		logger.Warnf(ctx, "schedule control can not get lab access key: %s", labUser.AccessKey)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("can not get lab"))
		return
	}

	labInfo := &schedule.LabInfo{
		LabUser: labUser,
		LabData: lab,
	}

	// 检查 redis 该调度器 set 是否已经存在该连接？
	if exist, err := i.consumer.HasUser(ctx, lab.UUID.String()); err != nil {
		logger.Errorf(ctx, "schedule control check edge exist err: %+v", err)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("schedule control check edge exist err"))
		return
	} else if exist {
		logger.Warnf(ctx, "schedule control edge already exist access key: %s", labUser.AccessKey)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("can not get lab"))
		return
	}

	if _, exist := i.labClient.LoadOrStore(lab.UUID, labInfo); exist {
		logger.Errorf(ctx, "schedule control check edge in schedule map already exist access key: %s", labUser.AccessKey)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("schedule already exist"))
		return
	}
	defer func() {
		i.labClient.Delete(lab.UUID)
	}()

	if err := i.consumer.AddUser(ctx, lab.UUID.String()); err != nil {
		logger.Errorf(ctx, "schedule control add edge to redis set access key: %s, err: %+v,", labUser.AccessKey, err)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("schedule add to redis set err"))
		return
	}

	defer func() {
		i.consumer.RemoveUser(ctx, lab.UUID.String())
	}()

	if err := i.wsClient.HandleRequestWithKeys(ginCtx.Writer, ginCtx.Request, map[string]any{
		schedule.LABINFO: labInfo,
		"ctx":            ctx,
		"lab_uuid":       lab.UUID,
		"lab_id":         lab.ID,
	}); err != nil {
		i.labClient.Delete(lab.UUID)
		if err := i.consumer.RemoveUser(ctx, lab.UUID.String()); err != nil {
			logger.Errorf(ctx, "schedule control remove user from lab uuid: %s, redis set err: %+v", lab.UUID.String(), err)
		}

		logger.Errorf(ctx, "schedule control HandleRequestWithKeys fail err: %+v", err)
	}
}

// edge 侧发送消息
func (i *control) OnEdgeMessge(ctx context.Context, s *melody.Session, b []byte) {
	logger.Infof(ctx, "schedule msg OnEdgeMessge job msg: %s", string(b))
	msgType := &common.WsMsgType{}
	err := json.Unmarshal(b, msgType)
	if err != nil {
		logger.Errorf(ctx, "OnEdgeMessge job msg Unmarshal err: %+v", err)
		return
	}

	switch schedule.ActionType(msgType.Action) {
	case schedule.JobStatus:
		i.onJobStatus(ctx, s, b)
	case schedule.DeviceStatus:
		i.onDeviceStatus(ctx, s, b)
	case schedule.Ping:
		i.onActionPong(ctx, s, b)
	case schedule.ReportActionState:
		i.onActionState(ctx, s, b)
	default:
		logger.Errorf(ctx, "OnEdgeMessge unknow msg: %s", string(b))
		return
	}
}

// 接受 redis queue 发送过来启动任务或者关闭任务的消息
func (i *control) OnJobMessage(ctx context.Context, msg []byte) {
	logger.Infof(ctx, "schedule msg OnJobMessage job msg: %s", msg)
	action := &engine.WorkflowInfo{}
	if err := json.Unmarshal(msg, action); err != nil {
		logger.Errorf(ctx, "OnJobMessage err: %+v", err)
		return
	}

	switch action.Action {
	case engine.StartJob:
		if err := i.createWrokflowTask(ctx, action); err != nil {
			logger.Errorf(ctx, "createWrokflowTask action: %+v, err: %+v", action, err)
		}
	case engine.StopJob:
		if err := i.stopWorkflowTask(ctx, action); err != nil {
			logger.Errorf(ctx, "stopWorkflowTask action: %+v, err: %+v", action, err)
		}
	case engine.StatusJob:
		if err := i.getWorkflowTask(ctx, action); err != nil {
			logger.Errorf(ctx, "getWorkflowTask action: %+v, err: %+v", action, err)
		}
	case engine.StartAction:
		if err := i.createActionTask(ctx, action); err != nil {
			logger.Errorf(ctx, "createWrokflowTask action: %+v, err: %+v", action, err)
		}
	case engine.AddMaterial, engine.UpdateMaterial, engine.RemoveMaterial:
		if err := i.notifyAddMaterial(ctx, action); err != nil {
			logger.Errorf(ctx, "notify add material action: %+v, err: %+v", action, err)
		}

	default:
		logger.Errorf(ctx, "OnJobMessage unknow action msg: %+s", string(msg))
		return
	}
}

// edge 侧发送 ping 消息
func (i *control) onActionPong(ctx context.Context, s *melody.Session, b []byte) {
	res := schedule.SendAction[ActionPong]{}
	if err := json.Unmarshal(b, &res); err != nil {
		logger.Errorf(ctx, "onActionState err: %+v", err)
		return
	}

	res.Data.ServerTimestamp = float64(time.Now().UnixMilli()) / 1000
	action := schedule.Pong
	i.sendAction(ctx, s, &schedule.SendAction[any]{
		Action: action,
		Data:   res.Data,
	})
}

// edge 更新 action 状态
func (i *control) onActionState(ctx context.Context, _ *melody.Session, b []byte) {
	// 处理任务状态
	res := schedule.SendAction[ActionStatus]{}
	if err := json.Unmarshal(b, &res); err != nil {
		logger.Errorf(ctx, "onActionState err: %+v", err)
		return
	}

	if res.Data.Type == "" ||
		res.Data.TaskID.IsNil() ||
		res.Data.JobID.IsNil() ||
		res.Data.DeviceID == "" ||
		res.Data.ActionName == "" {
		logger.Warnf(ctx, "onActionState param err: %+v", res)
		return
	}

	task, ok := i.tasks.Load(res.Data.TaskID)
	if ok {
		task.(*schedule.ControlTask).Task.SetDeviceActionStatus(ctx, res.Data.ActionKey, res.Data.ActionValue.Free, res.Data.NeedMore*time.Second)
	}
}

// edge 侧更新 job 的 status
func (i *control) onJobStatus(ctx context.Context, _ *melody.Session, b []byte) {
	// 处理任务状态
	res := schedule.SendAction[*engine.JobData]{}
	if err := json.Unmarshal(b, &res); err != nil {
		logger.Errorf(ctx, "onJobStatus err: %+v", err)
		return
	}

	task, ok := i.tasks.Load(res.Data.TaskID)
	if ok {
		task.(*schedule.ControlTask).Task.OnJobUpdate(ctx, res.Data)
	}
}

// 获取device 状态的 key
// edge 侧更新设备状态
func (i *control) onDeviceStatus(ctx context.Context, s *melody.Session, b []byte) {
	res := schedule.SendAction[schedule.DeviceData]{}
	if err := json.Unmarshal(b, &res); err != nil {
		logger.Errorf(ctx, "onJobStatus err: %+v", err)
		return
	}

	if res.Data.DeviceID == "" {
		logger.Errorf(ctx, "can not get device name: %s", string(b))
		return
	}

	valueI, ok := s.Get("lab_uuid")
	if !ok {
		logger.Warnf(ctx, "onDeviceStatus can not found uuid")
		return
	}
	labUUID, _ := valueI.(uuid.UUID)

	valueIDI, ok := s.Get("lab_id")
	if !ok {
		logger.Warnf(ctx, "onDeviceStatus can not found uuid")
		return
	}

	labID, _ := valueIDI.(int64)

	nodes, err := i.materialStore.UpdateMaterialNodeDataKey(ctx, labID,
		res.Data.DeviceID, res.Data.Data.PropertyName,
		res.Data.Data.Status)
	if err != nil {
		logger.Errorf(ctx, "onDeviceStatus update material data err: %+v", err)
		return
	}

	data := utils.FilterSlice(nodes, func(n *model.MaterialNode) (*material.UpdateMaterialData, bool) {
		return &material.UpdateMaterialData{
			UUID: n.UUID,
			Data: n.Data,
		}, true
	})

	d := material.UpdateMaterialDeviceNotify{
		Action: string(material.UpdateNodeData),
		Data:   data,
	}

	i.boardEvent.Broadcast(ctx, &notify.SendMsg{
		Channel:   notify.MaterialModify,
		LabUUID:   labUUID,
		UUID:      uuid.NewV4(),
		Data:      d,
		Timestamp: time.Now().Unix(),
	})
}

// 向 edge 侧发送通知消息
func (i *control) sendAction(ctx context.Context, s *melody.Session, data any) {
	bData, _ := json.Marshal(data)
	if err := s.Write(bData); err != nil {
		logger.Errorf(ctx, "sendAction err: %+v", err)
	}
}

// 创建工作流
func (i *control) createWrokflowTask(ctx context.Context, info *engine.WorkflowInfo) error {
	if info.TaskUUID.IsNil() ||
		info.WorkflowUUID.IsNil() ||
		info.LabUUID.IsNil() ||
		info.UserID == "" {
		return code.ParamErr.WithMsg("task uuid is empty")
	}

	allSession, _ := i.wsClient.Sessions()

	targetSession := utils.FindTarget(allSession, func(s *melody.Session) (*melody.Session, bool) {
		uuidI, ok := s.Get("lab_uuid")
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

	_, ok := i.tasks.Load(info.TaskUUID)
	if ok {
		return code.WorkflowTaskAlreadyExistErr.WithMsgf("task uuid: %s", info.TaskUUID.String())
	}

	taskCtx, cancle := context.WithCancel(ctx)
	task := dag.NewDagTask(taskCtx, &engine.TaskParam{
		Session: targetSession,
		Cancle:  cancle,
		Sandbox: i.sandbox,
	})

	controlTask := &schedule.ControlTask{
		Task:   task,
		Cancle: cancle,
		Ctx:    taskCtx,
	}

	v, ok := i.labClient.Load(info.LabUUID)
	if !ok {
		return code.EdgeConnectClosedErr.WithMsgf("can not found lab uuid: %+s", info.LabUUID)
	}

	info.LabData = v.(*schedule.LabInfo).LabData
	i.tasks.Store(info.TaskUUID, controlTask)

	if err := i.pools.Submit(func() {
		if err := task.Run(taskCtx, info); err != nil {
			logger.Warnf(ctx, "task run fail err: %+v", err)
		}

		// 停止任务
		controlTask.Cancle()
		// 删除记录
		i.tasks.Delete(info.TaskUUID)
	}); err != nil {
		logger.Errorf(ctx, "Submit job fail err: %+v", err)
		return err
	}

	return nil
}

// redis 停止工作流
func (i *control) stopWorkflowTask(ctx context.Context, info *engine.WorkflowInfo) error {
	value, ok := i.tasks.Load(info.TaskUUID)
	if !ok {
		logger.Warnf(ctx, "stopWorkflowTask task not exist task uuid: %+s", info.TaskUUID)

		return nil
	}

	task := value.(*schedule.ControlTask)
	err := task.Task.Stop(ctx)
	if err != nil {
		logger.Warnf(ctx, "stopWorkflowTask fail err: %+v", err)
	}

	i.tasks.Delete(info.TaskUUID)
	return err
}

// edge 侧获取工作流任务状态
func (i *control) getWorkflowTask(ctx context.Context, info *engine.WorkflowInfo) error {
	value, ok := i.tasks.Load(info.TaskUUID)
	if !ok {
		logger.Warnf(ctx, "stopWorkflowTask task not exist task uuid: %+s", info.TaskUUID)

		return nil
	}

	// FIXME: 需要获取到什么？
	task := value.(*schedule.ControlTask)
	if err := task.Task.GetStatus(ctx); err != nil {
		logger.Errorf(ctx, "stopWorkflowTask fail err: %+v", err)
	}

	return nil
}

// init websocket
func (i *control) initWebSocket(ctx context.Context) {
	// edge websocket 断开
	i.wsClient.HandleClose(func(s *melody.Session, _ int, _ string) error {
		// 关闭之后的回调
		ctxI, _ := s.Get("ctx")
		gCtx := ctxI.(*gin.Context)
		labUUID := s.MustGet("lab_uuid").(uuid.UUID)
		logger.Infof(gCtx, "client close keys: %+v", s.Keys)

		i.labClient.Delete(labUUID)
		if err := i.consumer.RemoveUser(gCtx, labUUID.String()); err != nil {
			logger.Errorf(ctx, "schedule control initWebSocket RemoveUser fail uuid: %s, err: %+v", labUUID.String(), err)
			return err
		}
		return nil
	})

	// client 读写全部退出
	i.wsClient.HandleDisconnect(func(s *melody.Session) {
		// melody client 端口后的最后一个回调事件
		ctx, _ := s.Get("ctx")
		gCtx := ctx.(*gin.Context)
		labUUID := s.MustGet("lab_uuid").(uuid.UUID)
		logger.Infof(gCtx, "client close keys: %+v", s.Keys)

		i.labClient.Delete(labUUID)
		if err := i.consumer.RemoveUser(context.Background(), labUUID.String()); err != nil {
			logger.Errorf(gCtx, "schedule control initWebSocket HandleDisconnect  uuid: %s, fail err: %+v", labUUID.String(), err)
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
			logger.Infof(ctx.(context.Context), "schedule control initWebSocket websocket find HandleError keys: %+v, err: %+v", s.Keys, err)
		}
	})

	i.wsClient.HandleConnect(func(s *melody.Session) {
		// melody 开始连接时报错, 忽略，已经在 connect 处理
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "schedule control initWebSocket HandleConnect websocket connect keys: %+v", s.Keys)
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
		i.OnEdgeMessge(ctxI.(context.Context), s, b)
	})

	i.wsClient.HandleSentMessage(func(_ *melody.Session, _ []byte) {
		// 发送完字符串消息后的回调
	})

	i.wsClient.HandleSentMessageBinary(func(_ *melody.Session, _ []byte) {
		// 发送完二进制消息后的回调
	})

	count := 0
	i.wsClient.HandlePong(func(s *melody.Session) {
		count++
		if count%500 == 0 {
			labUUIDI, ok := s.Get("lab_uuid")
			if !ok {
				return
			}
			labUUID := labUUIDI.(uuid.UUID)
			if ctx, ok := s.Get("ctx"); ok {
				logger.Infof(ctx.(context.Context), "==================== pong ===================== lab uuid: %s", labUUID.String())
			}
		}
	})
}

// 初始化 control
func (i *control) initControl(ctx context.Context) {
	i.pools, _ = ants.NewPool(poolSize)
	i.consumer = redis.NewMessageConsumer(i.rClient, i.scheduleName)
	i.initWebSocket(ctx)
	i.startConsumeJob(ctx)
}

// 异步启动获取 redis 消息
func (i *control) startConsumeJob(ctx context.Context) {
	conf := config.Global().Job
	count := 0
	utils.SafelyGo(func() {
		time.Sleep(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				logger.Infof(ctx, "======================startConsumeJob exit")
				return
			case <-time.Tick(registryPeriod):
				count++
				count = count % 200
				if count == 0 {
					logger.Infof(ctx, "========================= consumer message")
				}
				i.consumer.Message(ctx, conf.JobQueueName, func(msg []byte) {
					i.OnJobMessage(ctx, msg)
				})
			}
		}
	}, func(err error) {
		logger.Errorf(ctx, "consumer redis message fail err:%+v", err)
	})
}

// 关闭清理资源
func (i *control) Close(ctx context.Context) {
	if i.wsClient != nil {
		if err := i.wsClient.CloseWithMsg([]byte("reboot")); err != nil {
			logger.Errorf(ctx, "Close fail CloseWithMsg err: %+v", err)
		}
	}

	if i.consumer != nil {
		if err := i.consumer.Cleanup(context.Background()); err != nil {
			logger.Errorf(ctx, "Close fail Cleanup err: %+v", err)
		}
	}

	if i.pools != nil {
		i.pools.Release()
	}
}

func (i *control) createActionTask(ctx context.Context, info *engine.WorkflowInfo) error {
	if info.TaskUUID.IsNil() ||
		info.LabUUID.IsNil() ||
		info.UserID == "" {
		return code.ParamErr.WithMsg("task uuid is empty")
	}

	allSession, _ := i.wsClient.Sessions()

	targetSession := utils.FindTarget(allSession, func(s *melody.Session) (*melody.Session, bool) {
		uuidI, ok := s.Get("lab_uuid")
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

	_, ok := i.tasks.Load(info.TaskUUID)
	if ok {
		return code.WorkflowTaskAlreadyExistErr.WithMsgf("task uuid: %s", info.TaskUUID.String())
	}

	taskCtx, cancle := context.WithCancel(ctx)
	task := action.NewActionTask(taskCtx, &engine.TaskParam{
		Session: targetSession,
		Cancle:  cancle,
		Sandbox: i.sandbox,
	})

	controlTask := &schedule.ControlTask{
		Task:   task,
		Cancle: cancle,
		Ctx:    taskCtx,
	}

	v, ok := i.labClient.Load(info.LabUUID)
	if !ok {
		return code.EdgeConnectClosedErr.WithMsgf("can not found lab uuid: %+s", info.LabUUID)
	}

	info.LabData = v.(*schedule.LabInfo).LabData
	i.tasks.Store(info.TaskUUID, controlTask)

	if err := i.pools.Submit(func() {
		if err := task.Run(taskCtx, info); err != nil {
			logger.Warnf(ctx, "action task run fail err: %+v", err)
		}

		// 停止任务
		controlTask.Cancle()
		// 删除记录
		i.tasks.Delete(info.TaskUUID)
	}); err != nil {
		logger.Errorf(ctx, "Submit action job fail err: %+v", err)
		return err
	}

	return nil
}

func (c *control) notifyAddMaterial(ctx context.Context, info *engine.WorkflowInfo) error {
	// 异步通知，防止注册主消息
	if info.LabUUID.IsNil() {
		logger.Warnf(ctx, "notifyAddMaterial lab uuid is empty uuid: %+v", info.LabUUID)
		return nil
	}

	allSession, _ := c.wsClient.Sessions()
	targetSession := utils.FindTarget(allSession, func(s *melody.Session) (*melody.Session, bool) {
		uuidI, ok := s.Get("lab_uuid")
		if !ok {
			return nil, ok
		}
		if uuidI.(uuid.UUID) == info.LabUUID {
			return s, true
		}
		return nil, false
	})

	if targetSession == nil {
		return nil
	}

	data := map[string]any{
		"action": info.Action,
		"data":   info.Data,
	}

	dataB, _ := json.Marshal(data)
	if err := targetSession.Write(dataB); err != nil {
		logger.Errorf(ctx, "notifyAddMaterial data: %s, err: %+v", string(dataB), err)
	}

	return nil
}
