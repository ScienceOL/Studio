package events

import (
	"context"
	"sync"

	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/common/code"
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

func (events *Events) Registry(ctx context.Context, msgName notify.Action, handleFunc func(ctx context.Context, msg string) error) error {
	if _, ok := events.actions.LoadOrStore(msgName, handleFunc); ok {
		return code.NotifyActionAlreadyRegistryErr.WithMsg(string(msgName))
	}

	// 订阅消息
	sub := events.client.Subscribe(ctx, string(msgName))
	utils.SafelyGo(func() {
		ch := sub.Channel()
		for {
			select {
			case msg := <-ch:
				if err := handleFunc(ctx, msg.Payload); err != nil {
					logger.Errorf(ctx, "handle redis msg fail name: %s, err: %+v", msgName, err)
				}
			case <-ctx.Done():
				logger.Infof(ctx, "exit redis channel name: %s", string(msgName))
				sub.Unsubscribe(ctx, string(msgName))
				events.actions.Delete(msgName)
			}
		}
	}, func(err error) {
		logger.Errorf(ctx, "Registry handle msg err: %+v", err)
	})
	return nil
}

func (events *Events) Broadcast(ctx context.Context, msg *notify.SendMsg) error {
	ret := events.client.Publish(ctx, string(msg.Action), msg)
	if ret.Err() != nil {
		logger.Errorf(ctx, "send msg fail action: %s", msg.Action)
		return code.NotifySendMsgErr
	}

	return nil
}
