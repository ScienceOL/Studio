package events

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/utils"
)

/*
	使用 redis 的发布订阅实现多进程间广播通信
*/

var (
	once   sync.Once
	center *Events
)

type Events struct {
	actions sync.Map
	subs    sync.Map
	client  *r.Client
}

func NewEvents() notify.MsgCenter {
	once.Do(func() {
		center = &Events{
			actions: sync.Map{},
			client:  redis.GetClient(),
		}
	})

	return center
}

func (events *Events) Registry(ctx context.Context, msgName notify.Action, handleFunc notify.HandleFunc) error {
	if _, ok := events.actions.LoadOrStore(msgName, handleFunc); ok {
		return code.NotifyActionAlreadyRegistryErr.WithMsg(string(msgName))
	}

	// 订阅消息
	sub := events.client.Subscribe(ctx, string(msgName))
	events.subs.Store(msgName, sub)
	utils.SafelyGo(func() {
		ch := sub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					logger.Infof(ctx, "exit redis channel name: %s", string(msgName))
					sub.Unsubscribe(ctx, string(msgName))
					events.actions.Delete(msgName)
					return
				}

				if msg == nil {
					continue
				}
				if err := handleFunc(ctx, msg.Payload); err != nil {
					logger.Errorf(ctx, "handle redis msg fail name: %s, err: %+v", msgName, err)
				}
			case <-ctx.Done():
				logger.Infof(ctx, "exit redis channel name: %s", string(msgName))
				sub.Unsubscribe(ctx, string(msgName))
				events.actions.Delete(msgName)
				return
			}
		}
	}, func(err error) {
		logger.Errorf(ctx, "Registry handle msg err: %+v", err)
	})
	return nil
}

func (events *Events) Broadcast(ctx context.Context, msg *notify.SendMsg) error {
	msg.Timestamp = time.Now().Unix()
	if msg.UUID.IsNil() {
		msg.UUID = uuid.NewV4()
	}

	data, _ := json.Marshal(msg)
	// FIXME: ctx 被取消后，通知消息发送直接失败。
	ret := events.client.Publish(ctx, string(msg.Channel), data)
	if ret.Err() != nil {
		logger.Errorf(ctx, "send msg fail action: %s, err: %+v", msg.Channel, ret.Err())
		return code.NotifySendMsgErr
	}

	return nil
}

func (events *Events) Close(ctx context.Context) error {
	events.subs.Range(func(key, value any) bool {
		msgName := key.(notify.Action)
		sub := value.(*r.PubSub)
		if err := sub.Unsubscribe(ctx, string(msgName)); err != nil {
			logger.Errorf(ctx, "unsubscribe redis channel %s err: %+v", msgName, err)
		}
		events.subs.Delete(key)
		events.actions.Delete(key)
		return true
	})
	return nil
}
