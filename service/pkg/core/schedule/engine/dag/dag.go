package dag

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olahol/melody"
	"github.com/panjf2000/ants/v2"
	"github.com/scienceol/studio/service/internal/configs/schedule"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/core/schedule/device"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	"github.com/scienceol/studio/service/pkg/repo/model"
	wfl "github.com/scienceol/studio/service/pkg/repo/workflow"
	"github.com/scienceol/studio/service/pkg/utils"
)

type dagEngine struct {
	deviceManager device.Service
	job           *engine.WorkflowInfo
	Cancel        context.CancelFunc
	session       *melody.Session

	envStore      repo.LaboratoryRepo
	workflowStore repo.WorkflowRepo

	nodes []*model.WorkflowNode
	edges []*model.WorkflowEdge

	dependencies map[*model.WorkflowNode]map[*model.WorkflowNode]struct{}

	pools *ants.Pool
	wg    sync.WaitGroup

	boardEvent notify.MsgCenter
}

func NewDagTask(_ context.Context, param *engine.TaskParam) engine.Task {
	pools, _ := ants.NewPool(5,
		ants.WithExpiryDuration(10*time.Second))

	return &dagEngine{
		deviceManager: param.Devices,
		session:       param.Session,
		Cancel:        param.Cancle,
		envStore:      eStore.New(),
		workflowStore: wfl.New(),
		dependencies:  make(map[*model.WorkflowNode]map[*model.WorkflowNode]struct{}),
		pools:         pools,
		wg:            sync.WaitGroup{},
		boardEvent:    events.NewEvents(),
	}
}

func (d *dagEngine) checkTaskStatus(ctx context.Context) bool {
	tasks := make([]*model.WorkflowTask, 0, 1)
	if err := d.workflowStore.FindDatas(ctx, &tasks, map[string]any{
		"uuid": d.job.TaskUUID,
	}, "id", "uuid", "status"); err != nil {
		logger.Errorf(ctx, "can not found workflow task uuid: %s, err: %+v", d.job.TaskUUID, err)
		return false
	}

	if len(tasks) != 1 {
		logger.Errorf(ctx, "can not found workflow task uuid: %s", d.job.TaskUUID)
		return false
	}

	if tasks[0].Status == model.WorkflowTaskStatusStoped {
		return false
	}

	d.job.TaskID = tasks[0].ID
	return true
}

func (d *dagEngine) loadData(ctx context.Context) error {
	d.boardMsg(ctx, &engine.BoardMsg{
		Header:         "",
		WorkflowStatus: "starting",
		Status:         "pending",
		Type:           "info",
		Msg:            []string{"prepare to run node"},
	})

	wk, err := d.workflowStore.GetWorkflowByUUID(ctx, d.job.WorkflowUUID)
	if err != nil {
		return err
	}

	// 加载所有工作流节点数据
	allNodes, err := d.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": wk.ID,
	})
	if err != nil {
		return err
	}

	// 过滤可执行节点
	nodes := utils.FilterSlice(allNodes, func(node *model.WorkflowNode) (*model.WorkflowNode, bool) {
		if node.Type == model.WorkflowNodeGroup || node.Disabled {
			return nil, false
		}
		return node, true
	})

	// 获取节点UUID用于查询边
	nodeUUIDs := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		return node.UUID, true
	})

	edges, err := d.workflowStore.GetWorkflowEdges(ctx, nodeUUIDs)
	if err != nil {
		return err
	}

	d.nodes = nodes
	d.edges = edges
	return nil
}

func (d *dagEngine) boardMsg(ctx context.Context, msg *engine.BoardMsg) {
	if err := d.boardEvent.Broadcast(ctx, &notify.SendMsg{
		Channel:      notify.WorkflowRun,
		LabUUID:      d.job.LabUUID,
		WorkflowUUID: d.job.WorkflowUUID,
		TaskUUID:     d.job.TaskUUID,
		UserID:       d.job.UserID,
		Data:         msg,
		UUID:         d.job.TaskUUID,
		Timestamp:    time.Now().Unix(),
	}); err != nil {
		logger.Errorf(ctx, "schedule board msg fail err: %+v", err)
	}
}

func (d *dagEngine) buildTask(_ context.Context) error {
	// 构建图关系
	nodeMap := utils.Slice2Map(d.nodes, func(node *model.WorkflowNode) (uuid.UUID, *model.WorkflowNode) {
		return node.UUID, node
	})

	// 修复：构建正确的边关系映射，支持一对多关系
	targetSourcesMap := make(map[uuid.UUID][]uuid.UUID)
	sourceTargetsMap := make(map[uuid.UUID][]uuid.UUID)

	for _, edge := range d.edges {
		// 目标节点的所有源节点
		targetSourcesMap[edge.TargetNodeUUID] = append(targetSourcesMap[edge.TargetNodeUUID], edge.SourceNodeUUID)
		// 源节点的所有目标节点
		sourceTargetsMap[edge.SourceNodeUUID] = append(sourceTargetsMap[edge.SourceNodeUUID], edge.TargetNodeUUID)
	}

	// 先检测循环
	if err := d.detectCycle(sourceTargetsMap); err != nil {
		return err
	}

	for _, node := range d.nodes {
		// d.boardMsg(ctx, &engine.BoardMsg{
		// 	Header:         node.ActionName,
		// 	WorkflowStatus: "running",
		// 	Status:         "pending",
		// 	Type:           "info",
		// 	Msg:            []string{"prepare to run node"},
		// })

		parentNodeMap := make(map[*model.WorkflowNode]struct{})
		d.findAllParents(nodeMap, targetSourcesMap, node, parentNodeMap)
		d.dependencies[node] = parentNodeMap
	}
	return nil
}

// 使用DFS检测循环
func (d *dagEngine) detectCycle(edgeMap map[uuid.UUID][]uuid.UUID) error {
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)

	for _, node := range d.nodes {
		if !visited[node.UUID] {
			if d.dfsDetectCycle(node.UUID, edgeMap, visited, recStack) {
				return code.WorkflowHasCircularErr
			}
		}
	}
	return nil
}

func (d *dagEngine) dfsDetectCycle(nodeUUID uuid.UUID,
	edgeMap map[uuid.UUID][]uuid.UUID, visited, recStack map[uuid.UUID]bool,
) bool {
	visited[nodeUUID] = true
	recStack[nodeUUID] = true

	// 查找所有子节点
	if targets, exists := edgeMap[nodeUUID]; exists {
		for _, targetUUID := range targets {
			if !visited[targetUUID] {
				if d.dfsDetectCycle(targetUUID, edgeMap, visited, recStack) {
					return true
				}
			} else if recStack[targetUUID] {
				return true // 发现循环
			}
		}
	}

	recStack[nodeUUID] = false
	return false
}

// 修复后的findAllParents方法，支持多个父节点
func (d *dagEngine) findAllParents(nodeMap map[uuid.UUID]*model.WorkflowNode,
	targetSourcesMap map[uuid.UUID][]uuid.UUID, node *model.WorkflowNode, parentMap map[*model.WorkflowNode]struct{},
) {
	if node == nil {
		return
	}

	// 查找所有指向当前节点的父节点
	if sources, exists := targetSourcesMap[node.UUID]; exists {
		for _, sourceUUID := range sources {
			parentNode, ok := nodeMap[sourceUUID]
			if !ok {
				continue // 父节点不存在
			}

			// 避免重复访问
			if _, exists := parentMap[parentNode]; !exists {
				parentMap[parentNode] = struct{}{}
				// 递归查找父节点的父节点
				d.findAllParents(nodeMap, targetSourcesMap, parentNode, parentMap)
			}
		}
	}
}

func (d *dagEngine) Run(ctx context.Context, job *engine.WorkflowInfo) error {
	d.job = job
	if !d.checkTaskStatus(ctx) {
		return nil
	}
	if err := d.loadData(ctx); err != nil {
		return err
	}

	d.boardMsg(ctx, &engine.BoardMsg{
		Header:         "",
		WorkflowStatus: "starting",
		Status:         "pending",
		Type:           "info",
		Msg:            []string{"prepare to run node"},
	})
	if err := d.buildTask(ctx); err != nil {
		return err
	}

	err := d.runAllNodes(ctx)

	data := &engine.BoardMsg{
		Header:         "end",
		WorkflowStatus: "finished",
		Type:           "info",
		Status:         "success",
		Msg:            []string{"running node"},
	}
	if err != nil {
		data.Status = "fail"
		data.Msg = append(data.Msg, err.Error())
	}

	d.boardMsg(ctx, data)
	return err
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
			return nil
		default:
		}

		canRunNodes := make([]*model.WorkflowNode, 0, 10)
		nodeJobs := make([]*model.WorkflowNodeJob, 0, 10)
		for node, nodeDependences := range d.dependencies {
			if len(nodeDependences) == 0 {
				canRunNodes = append(canRunNodes, node)
				nodeJobs = append(nodeJobs, &model.WorkflowNodeJob{
					LabID:          d.job.LabData.ID,
					WorkflowTaskID: d.job.TaskID,
					NodeID:         node.ID,
					Status:         model.WorkflowJobPending,
				})
			}
		}

		if err := d.workflowStore.CreateJobs(closeCtx, nodeJobs); err != nil {
			return err
		}

		for index, node := range canRunNodes {
			newNode := node
			newInde := index
			d.wg.Add(1) // Add应该在Submit之前调用
			if err := d.pools.Submit(func() {
				defer d.wg.Done()
				if err := utils.SafelyRun(func() {
					select {
					case <-closeCtx.Done():
						return
					default:
					}
					if err := d.runNode(closeCtx, newNode, nodeJobs[newInde]); err != nil {
						logger.Errorf(closeCtx, "node run fail node id: %d, errr: %+v", newNode.ID, err)
						if !hasError.Load() {
							firstError.Store(err)
							hasError.Store(true)
							cancel()
						}
					}
				}); err != nil {
					logger.Errorf(ctx, "run all node SafelyRun err: %+v", err)
				}
			}); err != nil {
				logger.Errorf(ctx, "run all node submit run node fail err: %+v", err)
			}
		}
		d.wg.Wait()

		if hasError.Load() {
			return firstError.Load().(error)
		}

		for _, runnedNode := range canRunNodes {
			delete(d.dependencies, runnedNode)
			for _, nodeDependences := range d.dependencies {
				delete(nodeDependences, runnedNode)
			}
		}
	}
}

func (d *dagEngine) runNode(ctx context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	d.boardMsg(ctx, &engine.BoardMsg{
		Header:         node.ActionName,
		NodeUUID:       node.UUID,
		WorkflowStatus: "running",
		Status:         "running",
		Type:           "info",
		Msg:            []string{"running node"},
	})
	if err := d.sendAction(ctx, node, job); err != nil {
		return err
	}

	err := d.callbackAction(ctx, job)

	data := &engine.BoardMsg{
		Header:         node.ActionName,
		NodeUUID:       node.UUID,
		WorkflowStatus: "running",
		Type:           "info",
		Msg:            []string{"running node"},
	}
	data.Status = "success"
	if err != nil {
		data.Status = "failed"
		data.Msg = append(data.Msg, err.Error())
	}
	d.boardMsg(ctx, data)

	return err
}

func (d *dagEngine) sendAction(_ context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	param := node.Param

	if d.session.IsClosed() {
		return code.EdgeConnectClosedErr
	}
	// 数据库插入数据

	data := engine.SendActionData{
		DeviceID: utils.SafeValue(func() string {
			return *node.DeviceName
		}, ""),
		Action:     node.ActionName,
		ActionType: node.ActionType,
		ActionArgs: param,
		JobID:      job.UUID.String(),
		NodeID:     node.UUID.String(),
		ServerInfo: engine.ServerInfo{
			SendTimestamp: float64(time.Now().UnixNano()) / 1e9,
		},
	}

	b, err := json.Marshal(data)
	if err != nil {
		return code.NodeDataMarshalErr.WithErr(err)
	}

	return d.session.Write(b)
}

func (d *dagEngine) callbackAction(ctx context.Context, job *model.WorkflowNodeJob) error {
	// FIXME: 测试结束后放开这段代码。
	time.Sleep(5 * time.Second)
	return nil
	// 查询任务状态是否回调成功
	dagConf := schedule.Config().DynamicConfig.DagTask
	retry := dagConf.RetryCount
	var jobs []*model.WorkflowNodeJob
	for retry > 0 {
		time.Sleep(time.Duration(dagConf.Interval) * time.Second)
		err := d.workflowStore.FindDatas(ctx, &jobs, map[string]any{
			"id": job.ID,
		})
		if err != nil {
			return err
		}

		if len(jobs) != 1 {
			return code.RecordNotFound.WithMsgf("can not found job id: %+d", job.ID)
		}

		switch jobs[0].Status {
		case model.WorkflowJobSuccess:
			return nil
		case model.WorkflowJobFailed:
			return code.JobRunFailErr
		}

		retry--
	}

	logger.Warnf(ctx, "schedule job run timeout id: %d", job.ID)
	return nil
}

func (d *dagEngine) Stop(_ context.Context) error {
	d.Cancel()
	if d.pools != nil {
		d.pools.Release()
	}

	d.wg.Wait()
	return nil
}

func (d *dagEngine) GetStatus(_ context.Context) error {
	return nil
}
