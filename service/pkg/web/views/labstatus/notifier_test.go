package labstatus

import (
	"context"
	"testing"
	"time"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNotifier(t *testing.T) {
	notifier := GetGlobalNotifier()

	// 测试注册处理器
	called := false
	var receivedUUID uuid.UUID
	var receivedOnline bool

	handler := func(ctx context.Context, labUUID uuid.UUID, isOnline bool, lastConnectedAt *time.Time) {
		called = true
		receivedUUID = labUUID
		receivedOnline = isOnline
	}

	notifier.RegisterHandler(handler)

	// 触发通知
	testUUID := uuid.NewV4()
	now := time.Now()
	notifier.Notify(context.Background(), testUUID, true, &now)

	// 等待异步处理
	time.Sleep(100 * time.Millisecond)

	// 验证
	assert.True(t, called, "handler should be called")
	assert.Equal(t, testUUID, receivedUUID, "should receive correct UUID")
	assert.True(t, receivedOnline, "should receive correct online status")
}
