package edge

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/scienceol/studio/service/pkg/core/schedule/edge"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine/action"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/utils"
)

// 处理控制类消息

// control 消息
func (e *EdgeImpl) onControlMessage(ctx context.Context, msg string) {
	logger.Infof(ctx, "schedule msg OnJobMessage job msg: %s", msg)
	apiType := &edge.ApiControlMsg{}
	if err := json.Unmarshal([]byte(msg), apiType); err != nil {
		logger.Errorf(ctx, "onControlMessage err: %+v, msg: %s", err, msg)
		return
	}

	switch apiType.Action {
	case edge.StartAction:
		e.onStartAction(ctx, msg)
	case edge.StopJob:
		e.onStopJob(ctx, msg)
	case edge.StatusJob:
		e.onStatusJob(ctx, msg)
	case edge.AddMaterial, edge.UpdateMaterial, edge.RemoveMaterial:
		e.onMaterial(ctx, msg)
	default:
		logger.Errorf(ctx, "EdgeImpl.onControlMessage unknown action: %s", apiType.Action)
	}
}

func (e *EdgeImpl) onStartAction(ctx context.Context, msg string) {
	apiMsg := &edge.ApiControlData[engine.WorkflowInfo]{}
	if err := json.Unmarshal([]byte(msg), apiMsg); err != nil {
		logger.Errorf(ctx, "EdgeImpl.onWorkflowJob unmarshal err: %+v", err)
		return
	}

	defer func() { e.actionTask = nil }()
	e.actionTask = action.NewActionTask(ctx, &engine.TaskParam{
		Session:    e.labInfo.Session,
		Cancle:     e.cancel,
		Sandbox:    e.labInfo.Sandbox,
		BoardEvent: e.boardEvent,
	})

	if err := utils.SafelyRun(func() {
		if err := e.actionTask.Run(ctx, &apiMsg.Data); err != nil {
			logger.Errorf(ctx, "EdgeImpl.onStartAction run err: %+v", err)
		}
	}); err != nil {
		logger.Errorf(ctx, "EdgeImpl.onNotebookJob err: %+v", err)
	}
}

func (e *EdgeImpl) onStopJob(ctx context.Context, msg string) {
	// 停止 workflow 、notebook
	if e.jobTask == nil {
		return
	} else {
		v := reflect.ValueOf(e.jobTask)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return
		}
	}

	apiControlData := &edge.ApiControlData[edge.StopJobReq]{}
	if err := json.Unmarshal([]byte(msg), apiControlData); err != nil {
		logger.Errorf(ctx, "EdgeImpl.onAddMaterial unmarshal err: %+v", err)
		return
	}

	if apiControlData.Data.UUID == e.jobTask.ID(ctx) {
		if err := e.jobTask.Stop(ctx); err != nil {
			logger.Errorf(ctx, "EdgeImpl.onStopJob stop err: %+v", err)
		}
	}
}

func (e *EdgeImpl) onStatusJob(ctx context.Context, msg string) {
	panic("not implements")
}

func (e *EdgeImpl) onMaterial(ctx context.Context, msg string) {
	apiControlData := &edge.ApiControlData[any]{}
	if err := json.Unmarshal([]byte(msg), apiControlData); err != nil {
		logger.Errorf(ctx, "EdgeImpl.onAddMaterial unmarshal err: %+v", err)
		return
	}

	data := map[string]any{
		"action": apiControlData.Action,
		"data":   apiControlData.Data,
	}

	dataB, _ := json.Marshal(data)
	if err := e.labInfo.Session.Write(dataB); err != nil {
		logger.Errorf(ctx, "EdgeImpl.onAddMaterial notifyAddMaterial data: %s, err: %+v", string(dataB), err)
	}
}
