package edge

import (
	"context"
	"encoding/json"
	"time"

	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/material"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/schedule/edge"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/utils"
)

// 处理 edge 侧消息
// edge 侧发送消息
func (e *EdgeImpl) OnEdgeMessge(ctx context.Context, s *melody.Session, b []byte) {
	logger.Infof(ctx, "schedule msg OnEdgeMessge job msg: %s", string(b))
	edgeType := &edge.EdgeMsg{}
	err := json.Unmarshal(b, edgeType)
	if err != nil {
		logger.Errorf(ctx, "OnEdgeMessge job msg Unmarshal err: %+v", err)
		return
	}

	switch edgeType.Action {
	case edge.JobStatus:
		e.onJobStatus(ctx, s, b)
	case edge.DeviceStatus:
		e.onDeviceStatus(ctx, s, b)
	case edge.Ping:
		e.onPing(ctx, s, b)
	case edge.ReportActionState:
		e.onActionState(ctx, s, b)
	case edge.HostNodeReady:
		e.onEdgeReady(ctx, s, b)
	case edge.NormalExist:
		e.onNormalExit(ctx, s, b)
	default:
		logger.Errorf(ctx, "EdgeImpl.OnEdgeMessge unknow action: %s", edgeType.Action)
	}
}

// Edge Update Job Status
func (e *EdgeImpl) onJobStatus(ctx context.Context, s *melody.Session, b []byte) {
	res := edge.EdgeData[*engine.JobData]{}
	if err := json.Unmarshal(b, &res); err != nil {
		logger.Errorf(ctx, "onJobStatus err: %+v", err)
		return
	}

	e.onActionTask(ctx, &res)
	e.onJobTask(ctx, &res)
}

func (e *EdgeImpl) onActionTask(ctx context.Context, updateData *edge.EdgeData[*engine.JobData]) {
	if updateData == nil || updateData.Data == nil {
		return
	}

	if e.isTaskNil(ctx, e.actionTask) {
		return
	}

	if e.actionTask.ID(ctx) == updateData.Data.TaskID {
		e.actionTask.OnJobUpdate(ctx, updateData.Data)
	}
}

func (e *EdgeImpl) onJobTask(ctx context.Context, updateData *edge.EdgeData[*engine.JobData]) {
	if updateData == nil || updateData.Data == nil {
		return
	}

	if e.isTaskNil(ctx, e.jobTask) {
		return
	}

	if e.jobTask.ID(ctx) == updateData.Data.TaskID {
		e.jobTask.OnJobUpdate(ctx, updateData.Data)
	}
}

// Edge Device Status Update
func (e *EdgeImpl) onDeviceStatus(ctx context.Context, s *melody.Session, b []byte) {
	res := edge.EdgeData[edge.DeviceData]{}
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

	nodes, err := e.materialStore.UpdateMaterialNodeDataKey(ctx, labID,
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

	e.boardEvent.Broadcast(ctx, &notify.SendMsg{
		Channel:   notify.MaterialModify,
		LabUUID:   labUUID,
		UUID:      uuid.NewV4(),
		Data:      d,
		Timestamp: time.Now().Unix(),
	})
}

func (e *EdgeImpl) onPing(ctx context.Context, s *melody.Session, b []byte) {
	req := edge.EdgeData[edge.ActionPong]{}
	if err := json.Unmarshal(b, &req); err != nil {
		logger.Errorf(ctx, "onActionState err: %+v", err)
		return
	}

	req.Data.ServerTimestamp = float64(time.Now().UnixMilli()) / 1000
	e.sendAction(ctx, s, &edge.EdgeData[any]{
		EdgeMsg: edge.EdgeMsg{
			Action: edge.Pong,
		},
		Data: req.Data,
	})
}

func (e *EdgeImpl) onActionState(ctx context.Context, _ *melody.Session, b []byte) {
	// 处理任务状态
	res := edge.EdgeData[edge.ActionStatus]{}
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

	e.onAction(ctx, &res)
	e.onJob(ctx, &res)
}

func (e *EdgeImpl) onAction(ctx context.Context, data *edge.EdgeData[edge.ActionStatus]) {
	if data.Data.TaskID.IsNil() {
		return
	}

	if e.isTaskNil(ctx, e.actionTask) {
		return
	}

	if e.actionTask.ID(ctx) != data.Data.TaskID {
		return
	}

	e.actionTask.SetDeviceActionStatus(ctx, data.Data.ActionKey, data.Data.ActionValue.Free, data.Data.NeedMore*time.Second)
}

func (e *EdgeImpl) onJob(ctx context.Context, data *edge.EdgeData[edge.ActionStatus]) {
	if data.Data.TaskID.IsNil() {
		return
	}

	if e.isTaskNil(ctx, e.jobTask) {
		return
	}

	if e.jobTask.ID(ctx) != data.Data.TaskID {
		return
	}

	e.jobTask.SetDeviceActionStatus(ctx, data.Data.ActionKey, data.Data.ActionValue.Free, data.Data.NeedMore*time.Second)
}

func (e *EdgeImpl) onEdgeReady(ctx context.Context, _ *melody.Session, b []byte) {
	res := edge.EdgeData[edge.EdgeReady]{}
	if err := json.Unmarshal(b, &res); err != nil {
		logger.Errorf(ctx, "onActionState err: %+v", err)
		return
	}

	logger.Infof(ctx,
		"onEdgeReady lab id: %d, status: %s, timestamp: %f",
		e.labInfo.ID, res.Data.Status, res.Data.Timestamp)
	e.startTaskConsumer(e.ctx)
	e.startControlConsumer(e.ctx)
	e.wait.Add(2)
}

func (e *EdgeImpl) onNormalExit(ctx context.Context, _ *melody.Session, _ []byte) {
	logger.Infof(ctx, "EdgeImpl.onNormalExit starting lab id: %d", e.labInfo.ID)
	e.Close(ctx)
}
