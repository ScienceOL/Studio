package labstatus

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	r "github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
)

const (
	RedisChannelLabStatus = "lab_status_change"
)

var (
	globalNotifier *Notifier
	once           sync.Once
)

// StatusChangeEvent 状态变化事件
type StatusChangeEvent struct {
	LabUUID         uuid.UUID  `json:"lab_uuid"`
	IsOnline        bool       `json:"is_online"`
	LastConnectedAt *time.Time `json:"last_connected_at"`
}

// Notifier 全局状态通知器（使用 Redis Pub/Sub 实现跨进程通信）
type Notifier struct {
	handlers      []StatusChangeHandler
	mu            sync.RWMutex
	rClient       *r.Client
	pubsub        *r.PubSub
	stopChan      chan struct{}
	isSubscribing bool
}

// StatusChangeHandler 状态变化处理函数
type StatusChangeHandler func(ctx context.Context, labUUID uuid.UUID, isOnline bool, lastConnectedAt *time.Time)

// GetGlobalNotifier 获取全局通知器实例
func GetGlobalNotifier() *Notifier {
	once.Do(func() {
		rClient := redis.GetClient()
		globalNotifier = &Notifier{
			handlers:      make([]StatusChangeHandler, 0),
			rClient:       rClient,
			stopChan:      make(chan struct{}),
			isSubscribing: false,
		}
		logger.Infof(context.Background(), "🚀 [Global Notifier] Initialized with Redis client")
	})
	return globalNotifier
}

// RegisterHandler 注册状态变化处理器（只在 service 进程中调用）
func (n *Notifier) RegisterHandler(handler StatusChangeHandler) {
	n.mu.Lock()
	n.handlers = append(n.handlers, handler)
	handlerCount := len(n.handlers)
	n.mu.Unlock()

	logger.Infof(context.Background(), "✅ [Global Notifier] Handler registered, total handlers: %d", handlerCount)

	// 第一次注册 handler 时，启动 Redis 订阅
	if handlerCount == 1 && !n.isSubscribing {
		n.startSubscription()
	}
}

// startSubscription 启动 Redis 订阅（只在 service 进程中运行）
func (n *Notifier) startSubscription() {
	n.mu.Lock()
	if n.isSubscribing {
		n.mu.Unlock()
		return
	}
	n.isSubscribing = true
	n.mu.Unlock()

	ctx := context.Background()
	logger.Infof(ctx, "🎧 [Global Notifier] Starting Redis subscription on channel: %s", RedisChannelLabStatus)

	n.pubsub = n.rClient.Subscribe(ctx, RedisChannelLabStatus)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf(ctx, "❌ [Global Notifier] Subscription panic: %v", r)
			}
			if n.pubsub != nil {
				n.pubsub.Close()
			}
		}()

		ch := n.pubsub.Channel()
		logger.Infof(ctx, "✅ [Global Notifier] Redis subscription started, waiting for messages...")

		for {
			select {
			case <-n.stopChan:
				logger.Infof(ctx, "🛑 [Global Notifier] Subscription stopped")
				return
			case msg := <-ch:
				if msg == nil {
					continue
				}

				logger.Infof(ctx, "📨 [Global Notifier] Received Redis message: %s", msg.Payload)

				// 解析事件
				var event StatusChangeEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					logger.Errorf(ctx, "❌ [Global Notifier] Failed to unmarshal event: %v", err)
					continue
				}

				logger.Infof(ctx, "🔔 [Global Notifier] Processing event: lab=%s, online=%v", event.LabUUID, event.IsOnline)

				// 调用所有 handler
				n.mu.RLock()
				handlers := make([]StatusChangeHandler, len(n.handlers))
				copy(handlers, n.handlers)
				n.mu.RUnlock()

				for i, handler := range handlers {
					go func(h StatusChangeHandler, index int) {
						defer func() {
							if r := recover(); r != nil {
								logger.Errorf(ctx, "❌ [Global Notifier] Handler %d panic: %v", index, r)
							}
						}()
						logger.Infof(ctx, "📤 [Global Notifier] Calling handler %d...", index)
						h(ctx, event.LabUUID, event.IsOnline, event.LastConnectedAt)
						logger.Infof(ctx, "✅ [Global Notifier] Handler %d completed", index)
					}(handler, i)
				}
			}
		}
	}()
}

// Stop 停止订阅
func (n *Notifier) Stop() {
	close(n.stopChan)
}

// Notify 触发状态变化通知（通过 Redis Pub/Sub 发布事件，支持跨进程）
func (n *Notifier) Notify(ctx context.Context, labUUID uuid.UUID, isOnline bool, lastConnectedAt *time.Time) {
	logger.Infof(ctx, "🔔 [Global Notifier] Notify called: lab=%s, online=%v", labUUID, isOnline)

	// 构建事件
	event := StatusChangeEvent{
		LabUUID:         labUUID,
		IsOnline:        isOnline,
		LastConnectedAt: lastConnectedAt,
	}

	// 序列化事件
	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Errorf(ctx, "❌ [Global Notifier] Failed to marshal event: %v", err)
		return
	}

	logger.Infof(ctx, "📦 [Global Notifier] Publishing event to Redis: %s", string(eventBytes))

	// 发布到 Redis
	if err := n.rClient.Publish(ctx, RedisChannelLabStatus, eventBytes).Err(); err != nil {
		logger.Errorf(ctx, "❌ [Global Notifier] Failed to publish to Redis: %v", err)
		return
	}

	logger.Infof(ctx, "✅ [Global Notifier] Event published successfully to channel: %s", RedisChannelLabStatus)
}
