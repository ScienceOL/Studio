package dag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olahol/melody"
	"github.com/panjf2000/ants/v2"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/core/schedule"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	wfl "github.com/scienceol/studio/service/pkg/repo/workflow"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"
)

type stepFunc func(ctx context.Context) error

type dagEngine struct {
	job     *engine.WorkflowInfo
	cancel  context.CancelFunc
	ctx     context.Context
	session *melody.Session

	envStore      repo.LaboratoryRepo
	workflowStore repo.WorkflowRepo

	nodes   []*model.WorkflowNode           // 所有节点
	edges   []*model.WorkflowEdge           // 所有边
	handles []*model.WorkflowHandleTemplate // 所有 handles

	jobMap          map[uuid.UUID]*model.WorkflowNodeJob // 所有的 job map
	nodeMap         map[int64]*model.WorkflowNodeJob     // 所有的 node 对应的运行结果
	nodeParentEdges map[int64][]*engine.HandlePair       // 节点对应的所有 parent edge

	dependencies map[*model.WorkflowNode]map[*model.WorkflowNode]struct{} // dag 图依赖关系

	pools     *ants.Pool
	wg        sync.WaitGroup
	stepFuncs []stepFunc

	boardEvent notify.MsgCenter
	sandbox    repo.Sandbox

	actionStatus sync.Map
}

func NewDagTask(ctx context.Context, param *engine.TaskParam) engine.Task {
	pools, _ := ants.NewPool(5,
		ants.WithExpiryDuration(10*time.Second))

	d := &dagEngine{
		session:         param.Session,
		cancel:          param.Cancle,
		ctx:             ctx,
		envStore:        eStore.New(),
		workflowStore:   wfl.New(),
		dependencies:    make(map[*model.WorkflowNode]map[*model.WorkflowNode]struct{}),
		pools:           pools,
		wg:              sync.WaitGroup{},
		boardEvent:      events.NewEvents(),
		jobMap:          make(map[uuid.UUID]*model.WorkflowNodeJob),
		nodeMap:         make(map[int64]*model.WorkflowNodeJob),
		nodeParentEdges: make(map[int64][]*engine.HandlePair),
		sandbox:         param.Sandbox,
	}
	d.stepFuncs = append(d.stepFuncs,
		d.checkTaskStatus, // 检查任务状态
		d.loadData,        // 加载运行数据
		d.buildTask,       // 构建任务
		d.runAllNodes,     // 运行任务
	)

	return d
}

func (d *dagEngine) ID(ctx context.Context) uuid.UUID {
	if d.job == nil {
		return uuid.NewNil()
	}

	return d.job.TaskUUID
}

func (d *dagEngine) checkTaskStatus(ctx context.Context) error {
	task := &model.WorkflowTask{}
	if err := d.workflowStore.GetData(ctx, task, map[string]any{
		"uuid": d.job.TaskUUID,
	}, "id", "uuid", "status"); err != nil {
		logger.Errorf(ctx, "can not found workflow task uuid: %s, err: %+v", d.job.TaskUUID, err)
		return code.CanNotGetWorkflowTaskErr
	}

	if task.Status != model.WorkflowTaskStatusPending {
		return code.WorkflowTaskStatusNotPendingErr
	}

	d.job.TaskID = task.ID
	return nil
}

func (d *dagEngine) loadData(ctx context.Context) error {
	// 获取工作流
	wk, err := d.workflowStore.GetWorkflowByUUID(ctx, d.job.WorkflowUUID)
	if err != nil {
		return err
	}

	// 加载所有工作流节点数据
	allNodes, err := d.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": wk.ID,
		"type": []model.WorkflowNodeType{
			model.WorkflowNodeILab,
			model.WorkflowPyScript,
		},
	})
	if err != nil {
		return err
	}

	// 过滤检查可执行节点
	nodes, err := utils.FilterSliceWithErr(allNodes, func(node *model.WorkflowNode) ([]*model.WorkflowNode, bool, error) {
		if node.Type == model.WorkflowNodeGroup || node.Disabled {
			return nil, false, nil
		}

		if node.Type == model.WorkflowNodeILab {
			if node.DeviceName == nil || *node.DeviceName == "" {
				return nil, false, code.WorkflowNodeNoDeviceName
			}

			if node.ActionName == "" {
				return nil, false, code.WorkflowNodeNoActionName
			}

			if node.ActionType == "" {
				return nil, false, code.WorkflowNodeNoActionType
			}
		} else {
			// 计算类型
			if node.Script == nil || *node.Script == "" {
				return nil, false, code.WorkflowNodeScriptEmtpyErr
			}
		}

		return []*model.WorkflowNode{node}, true, nil
	})
	if err != nil {
		return err
	}

	// 节点UUID查询边
	nodeUUIDs := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		return node.UUID, true
	})

	edges, err := d.workflowStore.GetWorkflowEdges(ctx, nodeUUIDs)
	if err != nil {
		return err
	}

	edgeHandleUUIDs := make([]uuid.UUID, 0, 2*len(edges))
	utils.Range(edges, func(_ int, e *model.WorkflowEdge) bool {
		edgeHandleUUIDs = utils.AppendUniqSlice(edgeHandleUUIDs, e.SourceHandleUUID)
		edgeHandleUUIDs = utils.AppendUniqSlice(edgeHandleUUIDs, e.TargetHandleUUID)
		return true
	})

	handleTpls := make([]*model.WorkflowHandleTemplate, 0, len(edgeHandleUUIDs))
	if err := d.workflowStore.FindDatas(ctx, &handleTpls, map[string]any{
		"uuid": edgeHandleUUIDs,
	}); err != nil {
		return err
	}

	d.nodes = nodes
	d.edges = edges
	d.handles = handleTpls
	return nil
}

func (d *dagEngine) buildTask(ctx context.Context) error {
	// 构建图关系
	nodeMap := utils.Slice2Map(d.nodes, func(node *model.WorkflowNode) (uuid.UUID, *model.WorkflowNode) {
		return node.UUID, node
	})

	nodeParentUUIDMap := make(map[uuid.UUID][]uuid.UUID)
	nodeChildrenUUIDMap := make(map[uuid.UUID][]uuid.UUID)

	for _, edge := range d.edges {
		// 目标节点的所有源节点
		nodeParentUUIDMap[edge.TargetNodeUUID] = append(nodeParentUUIDMap[edge.TargetNodeUUID], edge.SourceNodeUUID)
		// 源节点的所有目标节点
		nodeChildrenUUIDMap[edge.SourceNodeUUID] = append(nodeChildrenUUIDMap[edge.SourceNodeUUID], edge.TargetNodeUUID)
	}

	// 先检测循环
	if err := d.detectCycle(nodeChildrenUUIDMap); err != nil {
		return err
	}
	handleMap := utils.Slice2Map(d.handles, func(h *model.WorkflowHandleTemplate) (uuid.UUID, *model.WorkflowHandleTemplate) {
		return h.UUID, h
	})

	for _, node := range d.nodes {
		parentNodeMap := make(map[*model.WorkflowNode]struct{})
		d.findAllParents(nodeMap, nodeParentUUIDMap, node, parentNodeMap)
		d.dependencies[node] = parentNodeMap

		// 找出该节点的所有前向边
		leftEdges := utils.FilterSlice(d.edges, func(e *model.WorkflowEdge) (*model.WorkflowEdge, bool) {
			if node.UUID == e.TargetNodeUUID {
				return e, true
			}
			return nil, false
		})

		// if config.Global().Dynamic().Schedule.TranslateNodeParam {
		if true {
			var err error
			d.nodeParentEdges[node.ID], err = utils.FilterSliceErr(leftEdges, func(e *model.WorkflowEdge) (*engine.HandlePair, bool, error) {
				sourceHandle, ok := handleMap[e.SourceHandleUUID]
				if !ok {
					return nil, false, code.CanNotFoundWorkflowHandleErr.WithMsg(fmt.Sprintf("node id: %d, source uuid: %s", node.ID, e.SourceHandleUUID))
				}

				targetHandle, ok := handleMap[e.TargetHandleUUID]
				if !ok {
					return nil, false, code.CanNotFoundWorkflowHandleErr.WithMsg(fmt.Sprintf("node id: %d, target uuid: %s", node.ID, e.TargetHandleUUID))
				}

				pair := &engine.HandlePair{
					SourceHandle: sourceHandle,
					TargetHandle: targetHandle,
				}

				sourceNode, ok := nodeMap[e.SourceNodeUUID]
				if ok {
					pair.SourceNode = sourceNode
				}
				// 不存在的情况是父节点被禁用了

				return pair, true, nil
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 使用DFS检测循环
func (d *dagEngine) detectCycle(nodeChildrenUUIDMap map[uuid.UUID][]uuid.UUID) error {
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)

	for _, node := range d.nodes {
		if !visited[node.UUID] {
			if d.dfsDetectCycle(node.UUID, nodeChildrenUUIDMap, visited, recStack) {
				return code.WorkflowHasCircularErr
			}
		}
	}

	return nil
}

func (d *dagEngine) dfsDetectCycle(nodeUUID uuid.UUID,
	nodeChildrenUUIDMap map[uuid.UUID][]uuid.UUID, visited, recStack map[uuid.UUID]bool,
) bool {
	visited[nodeUUID] = true
	recStack[nodeUUID] = true

	// 查找所有子节点
	if children, exists := nodeChildrenUUIDMap[nodeUUID]; exists {
		for _, child := range children {
			if !visited[child] {
				if d.dfsDetectCycle(child, nodeChildrenUUIDMap, visited, recStack) {
					return true
				}
			} else if recStack[child] {
				return true // 发现循环
			}
		}
	}

	recStack[nodeUUID] = false
	return false
}

// 修复后的findAllParents方法，支持多个父节点
func (d *dagEngine) findAllParents(nodeMap map[uuid.UUID]*model.WorkflowNode,
	nodeParentUUIDMap map[uuid.UUID][]uuid.UUID, node *model.WorkflowNode, parentMap map[*model.WorkflowNode]struct{},
) {
	if node == nil {
		return
	}

	// 查找该节点的所有父节点
	if sources, exists := nodeParentUUIDMap[node.UUID]; exists {
		for _, sourceUUID := range sources {
			parentNode, ok := nodeMap[sourceUUID]
			if !ok {
				continue // 父节点不存在
			}

			// 避免重复访问
			if _, exists := parentMap[parentNode]; !exists {
				parentMap[parentNode] = struct{}{}
				// 递归查找父节点的父节点
				d.findAllParents(nodeMap, nodeParentUUIDMap, parentNode, parentMap)
			}
		}
	}
}

// 运行入口
func (d *dagEngine) Run(ctx context.Context, job *engine.WorkflowInfo) error {
	d.job = job
	var err error
	data := &engine.BoardMsg{
		TaskStatus: "starting",
		JobStatus:  "pending",
		Header:     "",
		Type:       "info",
		Msg:        "prepare to run node",
		StackTrace: nil,
	}
	d.boardMsg(ctx, data)

	// 执行所有步骤
	for _, step := range d.stepFuncs {
		if err = step(ctx); err != nil {
			break
		}
	}

	taskStatus := model.WorkflowTaskStatusFailed
	data.TaskStatus = "end"
	data.Timestamp = time.Now()
	if err != nil {
		if errors.Is(err, code.JobTimeoutErr) {
			taskStatus = model.WorkflowTaskStatusTimeout
			data.Msg = "job timeout"
			data.Type = "warning"
		} else if errors.Is(err, code.JobCanceled) {
			taskStatus = model.WorkflowTaskStatusCanceled
			data.Msg = "job canceled"
			data.Type = "warning"
		} else {
			taskStatus = model.WorkflowTaskStatusFailed
			data.Msg = "job failed"
			data.Type = "error"
		}
	} else {
		taskStatus = model.WorkflowTaskStatusSuccessed
		data.Msg = "finished"
	}

	d.updateTaskStatus(ctx, taskStatus, d.job.TaskID)
	d.boardMsg(ctx, data)

	d.wg.Wait()
	return err
}

func (d *dagEngine) Stop(_ context.Context) error {
	data := schedule.SendAction[*engine.CancelTask]{
		Action: schedule.CancelTask,
		Data: &engine.CancelTask{
			TaskID: d.job.TaskUUID,
		},
	}
	b, _ := json.Marshal(data)
	d.session.Write(b)

	d.cancel()
	d.wg.Wait()

	if d.pools != nil {
		d.pools.Release()
	}

	return nil
}

func (d *dagEngine) runAllNodes(ctx context.Context) error {
	var hasError atomic.Bool
	var firstError atomic.Value
	closeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		if len(d.dependencies) == 0 {
			return nil
		}

		select {
		case <-closeCtx.Done():
			return code.JobCanceled
		default:
		}

		noDepNodes := make([]*model.WorkflowNode, 0, 10)
		nodeJobs := make([]*model.WorkflowNodeJob, 0, 10)
		for node, nodeDependences := range d.dependencies {
			if len(nodeDependences) > 0 {
				continue
			}

			noDepNodes = append(noDepNodes, node)
			nodeJobs = append(nodeJobs, &model.WorkflowNodeJob{
				LabID:          d.job.LabData.ID,
				WorkflowTaskID: d.job.TaskID,
				NodeID:         node.ID,
				Status:         model.WorkflowJobPending,
			})
		}

		if err := d.workflowStore.CreateJobs(closeCtx, nodeJobs); err != nil {
			return err
		}

		for _, job := range nodeJobs {
			d.jobMap[job.UUID] = job
			d.nodeMap[job.NodeID] = job
		}

		for index, node := range noDepNodes {
			newNode := node
			newIndex := index
			d.wg.Add(1)
			if err := d.pools.Submit(func() {
				defer d.wg.Done()

				if err := utils.SafelyRun(func() {
					select {
					case <-closeCtx.Done():
						return
					default:
					}

					if err := d.runNode(closeCtx, newNode, nodeJobs[newIndex]); err != nil {
						if !errors.Is(err, code.JobCanceled) {
							logger.Errorf(closeCtx, "node run fail node id: %d, err: %+v", newNode.ID, err)
						}

						if !hasError.Load() {
							firstError.Store(err)
							hasError.Store(true)
							cancel()
						}
					}
				}); err != nil {
					if !errors.Is(err, code.JobCanceled) {
						logger.Errorf(ctx, "run all node SafelyRun err: %+v", err)
					}
				}
			}); err != nil {
				logger.Errorf(ctx, "run all node submit run node fail err: %+v", err)
			}

		}

		d.wg.Wait()

		if hasError.Load() {
			return firstError.Load().(error)
		}

		// 移除依赖关系
		for _, runnedNode := range noDepNodes {
			delete(d.dependencies, runnedNode)
			for _, nodeDependences := range d.dependencies {
				delete(nodeDependences, runnedNode)
			}
		}
	}
}

func (d *dagEngine) parsePreNodeParam(ctx context.Context, node *model.WorkflowNode) error {
	// dynamicConf := config.Global().Dynamic()
	// if !dynamicConf.Schedule.TranslateNodeParam {
	// 	return nil
	// }

	pairs, ok := d.nodeParentEdges[node.ID]
	if !ok || len(pairs) == 0 {
		return nil
	}

	for _, p := range pairs {
		// 无父节点
		if p.SourceNode == nil {
			continue
		}

		if p.SourceHandle == nil || p.SourceHandle.DataKey == "" {
			continue
		}

		if p.TargetHandle == nil || p.TargetHandle.DataKey == "" {
			continue
		}

		if p.SourceHandle.DataSource != "executor" ||
			p.SourceHandle.HandleKey == "ready" {
			continue
		}

		job, ok := d.nodeMap[p.SourceNode.ID]
		if !ok {
			return code.CanNotGetParentJobErr.WithMsg(
				fmt.Sprintf("parent node id: %d, node id: %d", p.SourceNode.ID, node.ID))
		}

		var err error
		retValueB, err := json.Marshal(job.ReturnInfo.Data().ReturnValue)
		if err != nil {
			return code.DataNotMapAnyTypeErr
		}
		res := gjson.Get(string(retValueB), p.SourceHandle.DataKey)
		if !res.Exists() {
			return code.ValueNotExistErr
		}

		jsonStr, err := sjson.Set(string(node.Param), p.TargetHandle.DataKey, res.Value())
		if err != nil {
			return code.UpdateNodeErr
		}

		node.Param = datatypes.JSON(jsonStr)
	}

	return nil
}

func (d *dagEngine) runNode(ctx context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	if err := d.parsePreNodeParam(ctx, node); err != nil {
		return err
	}

	data := &engine.BoardMsg{
		TaskStatus: "running",
		JobStatus:  "running",
		Header:     node.ActionName,
		NodeUUID:   node.UUID,
		Type:       "info",
		Msg:        "running",
	}

	d.boardMsg(ctx, data)

	var err error
	defer func() {
		jobStatus := model.WorkflowJobFailed
		data.Msg = "failed"
		data.Timestamp = time.Now()
		data.ReturnInfos = job.ReturnInfo
		if err != nil {
			jobStatus = model.WorkflowJobFailed
			if errors.Is(err, code.JobCanceled) {
				data.Msg = "job canceled"
				data.Type = "warning"
				data.JobStatus = "failed"
				jobStatus = model.WorkflowJobCanceled
			}

			if errors.Is(err, code.JobTimeoutErr) {
				data.Msg = "job timeout"
				data.Type = "warning"
				data.JobStatus = "failed"
				jobStatus = model.WorkflowJobTimeout
			} else {
				data.Msg = "job failed"
				data.Type = "warning"
				data.JobStatus = "failed"
				data.StackTrace = append(data.StackTrace, err.Error())
				jobStatus = model.WorkflowJobFailed
			}
		} else {
			data.Msg = "success"
			data.JobStatus = "success"
			jobStatus = model.WorkflowJobSuccess
		}

		d.boardMsg(ctx, data)
		d.updateJob(ctx, jobStatus, job.ID)
	}()

	// 查询 action 是否可以执行
	if node.Type == model.WorkflowNodeILab {
		err = d.queryAction(ctx, node, job)
		if err != nil {
			return err
		}
	}

	err = d.execNodeAction(ctx, node, job)
	if err != nil {
		return err
	}

	if node.Type == model.WorkflowPyScript {
		return err
	}

	key := engine.ActionKey{
		Type:       engine.JobCallbackStatus,
		TaskID:     d.job.TaskUUID,
		JobID:      job.UUID,
		DeviceID:   *node.DeviceName,
		ActionName: node.ActionName,
	}

	d.InitDeviceActionStatus(ctx, key, time.Now().Add(20*time.Second), false)
	err = d.callbackAction(ctx, key, job)

	return err
}

func (d *dagEngine) queryAction(ctx context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	if node.Type == model.WorkflowPyScript {
		return nil
	}

	key := engine.ActionKey{
		Type:   engine.QueryActionStatus,
		TaskID: d.job.TaskUUID,
		JobID:  job.UUID,
		DeviceID: utils.SafeValue(func() string {
			return *node.DeviceName
		}, ""),
		ActionName: node.ActionName,
	}
	d.InitDeviceActionStatus(ctx, key, time.Now().Add(time.Second*20), false)
	if err := d.sendQueryAction(ctx, node, job); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return code.JobCanceled
		default:
		}

		time.Sleep(time.Millisecond * 500)
		value, exist := d.GetDeviceActionStatus(ctx, key)
		if !exist {
			return code.QueryJobStatusKeyNotExistErr
		}

		if value.Free {
			d.DelStatus(ctx, key)
			return nil
		}

		if value.Timestamp.Unix() < time.Now().Unix() {
			return code.JobTimeoutErr
		}
	}
}

func (d *dagEngine) sendQueryAction(_ context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	if d.session.IsClosed() {
		return code.EdgeConnectClosedErr
	}

	data := schedule.SendAction[engine.ActionKey]{
		Action: schedule.QueryActionStatus,
		Data: engine.ActionKey{
			TaskID:     d.job.TaskUUID,
			JobID:      job.UUID,
			DeviceID:   *node.DeviceName,
			ActionName: node.ActionName,
		},
	}

	bData, _ := json.Marshal(data)
	return d.session.Write(bData)
}

func (d *dagEngine) execNodeAction(ctx context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	switch node.Type {
	case model.WorkflowNodeILab:
		return d.sendAction(ctx, node, job)
	case model.WorkflowPyScript:
		return d.execScript(ctx, node, job)
	default:
		return code.UnknownWorkflowNodeTypeErr
	}
}

func (d *dagEngine) execScript(ctx context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	inputs := map[string]any{}
	err := json.Unmarshal(node.Param, &inputs)
	returnInfo := model.ReturnInfo{
		Suc:         false,
		Error:       "",
		ReturnValue: nil,
	}

	ret, errMsg, err := d.sandbox.ExecCode(ctx, *node.Script, inputs)
	returnInfo.Error = errMsg
	returnInfo.ReturnValue = ret
	if err != nil {
		returnInfo.Suc = false
	}

	if err != nil || errMsg != "" {
		job.Status = model.WorkflowJobFailed
	}

	job.ReturnInfo = datatypes.NewJSONType(returnInfo)
	job.UpdatedAt = time.Now()

	if err := d.workflowStore.UpdateData(ctx, job, map[string]any{
		"uuid": job.UUID,
	}, "status", "feedback_data", "return_info", "updated_at"); err != nil {
		logger.Errorf(ctx, "onJobStatus update job fail uuid: %s, err: %+v", job.UUID, err)
	}

	return err
}

func (d *dagEngine) sendAction(_ context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	if d.session.IsClosed() {
		return code.EdgeConnectClosedErr
	}

	data := schedule.SendAction[*engine.SendActionData]{
		Action: schedule.JobStart,
		Data: &engine.SendActionData{
			DeviceID:   *node.DeviceName,
			Action:     node.ActionName,
			ActionType: node.ActionType,
			ActionArgs: node.Param,
			JobID:      job.UUID,
			TaskID:     d.job.TaskUUID,
			NodeID:     node.UUID,
			ServerInfo: engine.ServerInfo{
				SendTimestamp: float64(time.Now().UnixNano()) / 1e9,
			},
		},
	}

	b, err := json.Marshal(data)
	if err != nil {
		return code.NodeDataMarshalErr.WithErr(err)
	}

	return d.session.Write(b)
}

func (d *dagEngine) callbackAction(ctx context.Context, key engine.ActionKey, job *model.WorkflowNodeJob) error {
	for {
		select {
		case <-ctx.Done():
			return code.JobCanceled
		default:
		}

		time.Sleep(time.Millisecond * 500)
		value, exist := d.GetDeviceActionStatus(ctx, key)
		if !exist {
			return code.CallbackJobStatusKeyNotExistErr
		}

		if value.Free {
			d.DelStatus(ctx, key)
			break
		}

		if value.Timestamp.Unix() < time.Now().Unix() {
			return code.JobTimeoutErr
		}
	}

	// 查询任务状态是否回调成功
	err := d.workflowStore.GetData(ctx, job, map[string]any{
		"id": job.ID,
	})
	if err != nil {
		return err
	}

	logger.Infof(ctx, "schedule job run finished: %d", job.ID)
	switch job.Status {
	case model.WorkflowJobSuccess:
		return nil
	case model.WorkflowJobFailed:
		return code.JobRunFailErr
	default:
		return code.JobRunFailErr
	}
}

func (d *dagEngine) GetStatus(_ context.Context) error {
	return nil
}

func (d *dagEngine) OnJobUpdate(ctx context.Context, data *engine.JobData) error {
	if data.Status == "running" {
		return nil
	}

	d.SetDeviceActionStatus(ctx, engine.ActionKey{
		Type:       engine.JobCallbackStatus,
		TaskID:     data.TaskID,
		JobID:      data.JobID,
		DeviceID:   data.DeviceID,
		ActionName: data.ActionName,
	}, true, 0)

	if job, ok := d.jobMap[data.JobID]; ok {
		job.ReturnInfo = data.ReturnInfo
		job.FeedbackData = data.FeedbackData
		job.Status = model.WorkflowJobStatus(data.Status)
	}

	if err := d.workflowStore.UpdateData(ctx, &model.WorkflowNodeJob{
		Status:       model.WorkflowJobStatus(data.Status),
		FeedbackData: data.FeedbackData,
		ReturnInfo:   data.ReturnInfo,
		BaseModel: model.BaseModel{
			UpdatedAt: time.Now(),
		},
		// Timestamp:    data.Timestamp,
	}, map[string]any{
		"uuid": data.JobID,
	}, "status", "feedback_data", "return_info", "updated_at"); err != nil {
		logger.Errorf(ctx, "onJobStatus update job fail uuid: %s, err: %+v", data.JobID, err)
	}

	return nil
}

func (d *dagEngine) updateTaskStatus(ctx context.Context, status model.WorkflowTaskStatus, taskID int64) {
	data := &model.WorkflowTask{
		Status:       status,
		FinishedTime: time.Now(),
	}
	data.UpdatedAt = time.Now()
	if err := d.workflowStore.UpdateData(context.Background(), data, map[string]any{
		"id": taskID,
	}, "status", "updated_at", "finished_time"); err != nil {
		logger.Errorf(ctx, "engine dag updateTask id: %d, err: %+v", taskID, err)
	}
}

func (d *dagEngine) updateJob(ctx context.Context, status model.WorkflowJobStatus, jobID int64) {
	data := &model.WorkflowNodeJob{
		Status: status,
	}
	data.UpdatedAt = time.Now()

	if err := d.workflowStore.UpdateData(context.Background(), data, map[string]any{
		"id": jobID,
	}, "status", "updated_at"); err != nil {
		logger.Errorf(ctx, "engine dag updateJob job id: %+v, err: %+v", jobID, err)
	}
}

func (d *dagEngine) boardMsg(ctx context.Context, msg *engine.BoardMsg) {
	if err := d.boardEvent.Broadcast(context.Background(), &notify.SendMsg{
		Channel:      notify.WorkflowRun,
		TaskUUID:     d.job.TaskUUID,
		LabUUID:      d.job.LabUUID,
		WorkflowUUID: d.job.WorkflowUUID,
		UserID:       d.job.UserID,
		UUID:         d.job.TaskUUID,
		Data:         msg,
		Timestamp:    time.Now().Unix(),
	}); err != nil {
		logger.Errorf(ctx, "schedule board msg fail err: %+v", err)
	}
}

func (d *dagEngine) GetDeviceActionStatus(ctx context.Context, key engine.ActionKey) (engine.ActionValue, bool) {
	valueI, ok := d.actionStatus.Load(key)
	if !ok {
		return engine.ActionValue{}, false
	}
	return valueI.(engine.ActionValue), true
}

func (d *dagEngine) SetDeviceActionStatus(ctx context.Context, key engine.ActionKey, free bool, needMore time.Duration) {
	valueI, ok := d.actionStatus.Load(key)
	if ok {
		value := valueI.(engine.ActionValue)
		value.Free = free
		value.Timestamp = value.Timestamp.Add(needMore)
		logger.Warnf(ctx, "SetDeviceActionStatus key: %+v, value: %+v, more: %d", key, value, needMore)
		d.actionStatus.Store(key, value)
	} else {
		logger.Warnf(ctx, "SetDeviceActionStatus not found key: %+v", key)
	}
}

func (d *dagEngine) InitDeviceActionStatus(ctx context.Context, key engine.ActionKey, start time.Time, free bool) {
	d.actionStatus.Store(key, engine.ActionValue{
		Timestamp: start,
		Free:      free,
	})
}

func (d *dagEngine) DelStatus(ctx context.Context, key engine.ActionKey) {
	d.actionStatus.Delete(key)
}
