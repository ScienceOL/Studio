package edge

import (
	"context"
	"encoding/json"

	"github.com/scienceol/studio/service/pkg/core/schedule/edge"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

//	处理 api 任务类消息

// job 运行工作流消息
func (e *EdgeImpl) OnJobMessage(ctx context.Context, msg string) {
	logger.Infof(ctx, "schedule msg OnJobMessage job msg: %s", msg)
	apiType := &edge.ApiMsg{}
	if err := json.Unmarshal([]byte(msg), apiType); err != nil {
		logger.Errorf(ctx, "OnJobMessage err: %+v, msg: %s", err, msg)
		return
	}

	switch apiType.Action {

	default:
		logger.Errorf(ctx, "EdgeImpl.onJobMessage unknown action: %s", apiType.Action)
	}
}
