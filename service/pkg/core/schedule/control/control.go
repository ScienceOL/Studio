package control

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alphadose/haxmap"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	"github.com/panjf2000/ants/v2"
	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/core/schedule"
	"github.com/scienceol/studio/service/pkg/core/schedule/edge"
	edgeImpl "github.com/scienceol/studio/service/pkg/core/schedule/edge/edge"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	mStore "github.com/scienceol/studio/service/pkg/repo/material"
	s "github.com/scienceol/studio/service/pkg/repo/sandbox"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/scienceol/studio/service/pkg/web/views/labstatus"
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
	wsClient      *melody.Melody                // websocket 连接控制
	scheduleName  string                        // 调度器名
	labMap        *haxmap.Map[int64, edge.Edge] // lab 信息
	rClient       *r.Client                     // redis client
	pools         *ants.Pool                    // 任务池
	boardEvent    notify.MsgCenter              // 广播系统
	sandbox       repo.Sandbox                  // 脚本运行沙箱
	labStore      repo.LaboratoryRepo           // 实验室存储
	materialStore repo.MaterialRepo             // 物料调度
}

func NewControl(ctx context.Context) schedule.Control {
	once.Do(func() {
		wsClient := melody.New()
		wsClient.Config.MaxMessageSize = constant.MaxMessageSize
		wsClient.Config.PingPeriod = 10 * time.Second
		scheduleName := fmt.Sprintf("lab-schedule-name-%s", uuid.NewV4().String())
		logger.Infof(ctx, "====================schedule name: %s ======================", scheduleName)

		ctl = &control{
			wsClient:      wsClient,
			scheduleName:  scheduleName,
			rClient:       redis.GetClient(),
			labMap:        haxmap.New[int64, edge.Edge](),
			labStore:      eStore.New(),
			materialStore: mStore.NewMaterialImpl(),
			boardEvent:    events.NewEvents(),
			sandbox:       s.NewSandbox(),
		}
		ctl.pools, _ = ants.NewPool(poolSize)
		ctl.initWebSocket(ctx)
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

	// Set lab heart key in redis
	setSuccess, err := i.rClient.SetNX(ctx,
		utils.LabHeartName(labUser.LabUUID),
		time.Now().UTC(),
		100*utils.LabHeartTime-time.Second).Result()
	if err != nil {
		logger.Errorf(ctx, "schedule control set lab heart fail uuid: %s, err: %+v", labUser.LabUUID, err)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsgf("set lab heart err: %+v", err))
		return
	}

	if !setSuccess {
		logger.Warnf(ctx, "schedule control lab already connect uuid: %s", labUser.LabUUID)
		common.ReplyErr(ginCtx, code.ParamErr.WithMsg("lab already exist"))
		return
	}

	defer func() {
		if _, err := i.rClient.Del(context.Background(),
			utils.LabHeartName(labUser.LabUUID)).Result(); err != nil {
			logger.Errorf(ctx, "schedule control lab already connet uuid: %s", labUser.LabUUID)
		}
	}()

	// 更新数据库：设置实验室为在线状态
	now := time.Now()
	logger.Infof(ctx, "🟢 [Schedule Control] Lab connecting: %s (ID: %d)", lab.UUID, lab.ID)
	if err := i.labStore.UpdateLabOnlineStatus(ctx, lab.ID, true, &now); err != nil {
		logger.Errorf(ctx, "❌ [Schedule Control] Failed to update lab online status err: %+v", err)
	} else {
		logger.Infof(ctx, "✅ [Schedule Control] Lab online status updated in DB, now notifying...")
		// 通知状态变化
		labstatus.GetGlobalNotifier().Notify(ctx, lab.UUID, true, &now)
		logger.Infof(ctx, "📡 [Schedule Control] Global notifier called for lab %s", lab.UUID)
	}

	if err := i.wsClient.HandleRequestWithKeys(ginCtx.Writer, ginCtx.Request, map[string]any{
		schedule.LABINFO: labInfo,
		"ctx":            ctx,
		"lab_uuid":       lab.UUID,
		"lab_id":         lab.ID,
		"lab_user_id":    labUser.ID,
	}); err != nil {
		logger.Errorf(ctx, "schedule control HandleRequestWithKeys fail err: %+v", err)
	}
}

// init websocket
func (i *control) initWebSocket(ctx context.Context) {

	// 连接正式创建
	i.wsClient.HandleConnect(func(s *melody.Session) {
		labID := s.MustGet("lab_id").(int64)
		labUUID := s.MustGet("lab_uuid").(uuid.UUID)
		sessionCtx := s.MustGet("ctx").(*gin.Context)
		labUserID := s.MustGet("lab_user_id").(string)
		labInfo := &edge.LabInfo{
			UUID:      labUUID,
			ID:        labID,
			LabUserID: labUserID,
			Session:   s,
		}

		edgeImpl, err := edgeImpl.NewEdge(sessionCtx, labInfo)
		if err != nil {
			s.CloseWithMsg(fmt.Appendf(nil, "create lab instance fail err: %+v", err))
			return
		}

		if oldEdgeImpl, ok := i.labMap.Get(labID); ok {
			oldEdgeImpl.Close(sessionCtx)
		}

		i.labMap.Set(labID, edgeImpl)
	})

	// edge websocket 断开
	i.wsClient.HandleClose(func(s *melody.Session, _ int, _ string) error {
		// 关闭之后的回调
		labUUID := s.MustGet("lab_uuid").(uuid.UUID)
		labID := s.MustGet("lab_id").(int64)
		ctx := s.MustGet("ctx").(*gin.Context)
		if edgeImpl, ok := i.labMap.GetAndDel(labID); ok && edgeImpl != nil {
			edgeImpl.Close(ctx)
		}

		// 更新数据库：设置实验室为离线状态
		now := time.Now()
		logger.Infof(ctx, "🔴 [Schedule Control] Lab disconnecting (HandleClose): %s (ID: %d)", labUUID, labID)
		if err := i.labStore.UpdateLabOnlineStatus(ctx, labID, false, &now); err != nil {
			logger.Errorf(ctx, "❌ [Schedule Control] Failed to update lab offline status err: %+v", err)
		} else {
			logger.Infof(ctx, "✅ [Schedule Control] Lab offline status updated in DB, now notifying...")
			// 通知状态变化
			labstatus.GetGlobalNotifier().Notify(ctx, labUUID, false, &now)
			logger.Infof(ctx, "📡 [Schedule Control] Global notifier called for lab %s", labUUID)
		}

		return nil
	})

	// edge 资源回收
	i.wsClient.HandleDisconnect(func(s *melody.Session) {
		labUUID := s.MustGet("lab_uuid").(uuid.UUID)
		labID := s.MustGet("lab_id").(int64)
		ctx := s.MustGet("ctx").(*gin.Context)
		if edgeImpl, ok := i.labMap.GetAndDel(labID); ok && edgeImpl != nil {
			edgeImpl.Close(ctx)
		}

		// 更新数据库：设置实验室为离线状态
		now := time.Now()
		logger.Infof(ctx, "🔴 [Schedule Control] Lab disconnecting (HandleDisconnect): %s (ID: %d)", labUUID, labID)
		if err := i.labStore.UpdateLabOnlineStatus(context.Background(), labID, false, &now); err != nil {
			logger.Errorf(ctx, "❌ [Schedule Control] Failed to update lab offline status err: %+v", err)
		} else {
			logger.Infof(ctx, "✅ [Schedule Control] Lab offline status updated in DB, now notifying...")
			// 通知状态变化
			labstatus.GetGlobalNotifier().Notify(context.Background(), labUUID, false, &now)
			logger.Infof(ctx, "📡 [Schedule Control] Global notifier called for lab %s", labUUID)
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

	i.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		labID := s.MustGet("lab_id").(int64)
		sessionCtx := s.MustGet("ctx").(*gin.Context)
		edgeImpl, ok := i.labMap.Get(labID)
		if !ok {
			logger.Errorf(sessionCtx, "can not get lab impl lab id: %d", labID)
			return
		}

		edgeImpl.OnEdgeMessge(sessionCtx, s, b)
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
			labID := s.MustGet("lab_id").(int64)
			sessionCtx := s.MustGet("ctx").(*gin.Context)
			_, ok := i.labMap.Get(labID)
			if !ok {
				logger.Errorf(sessionCtx, "can not get lab impl lab id: %d", labID)
				return
			}
			// edgeImpl.OnPongMessage(sessionCtx)
		}
	})
}

// 关闭清理资源
func (i *control) Close(ctx context.Context) {
	if i.wsClient != nil {
		if err := i.wsClient.CloseWithMsg([]byte("reboot")); err != nil {
			logger.Errorf(ctx, "Close fail CloseWithMsg err: %+v", err)
		}
	}

	i.labMap.ForEach(func(i int64, e edge.Edge) bool {
		e.Close(ctx)
		return true
	})

	if i.pools != nil {
		i.pools.Release()
	}
}
