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

func NewDagTask(ctx context.Context, param *engine.TaskParam) engine.Task {
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

func (d *dagEngine) loadData(ctx context.Context) error {
	wk, err := d.workflowStore.GetWorkflowByUUID(ctx, d.job.WorkflowUUID)
	if err != nil {
		return err
	}

	// 加载所有工作流节点数据
	nodes, err := d.workflowStore.GetWorkflowNodes(ctx, map[string]any{
		"workflow_id": wk.ID,
	})

	if err != nil {
		return err
	}

	nodeUUIDs := utils.FilterSlice(nodes, func(node *model.WorkflowNode) (uuid.UUID, bool) {
		if node.Type == model.WorkflowNodeGroup {
			return node.UUID, false
		}

		if node.Disabled {
			return node.UUID, false
		}

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

func (d *dagEngine) buildTask(ctx context.Context) error {
	// 构建图关系
	nodeMap := utils.SliceToMap(d.nodes, func(node *model.WorkflowNode) (uuid.UUID, *model.WorkflowNode) {
		return node.UUID, node
	})

	targetSourceMap := utils.SliceToMap(d.edges, func(edge *model.WorkflowEdge) (uuid.UUID, uuid.UUID) {
		return edge.TargetNodeUUID, edge.SourceNodeUUID
	})

	for _, node := range d.nodes {
		d.boardMsg(ctx, &engine.BoardMsg{
			Header: node.ActionName,
			Status: "pending",
			Type:   "info",
			Msg:    []string{"prepare to run node"},
		})

		parentNodeMap := make(map[*model.WorkflowNode]struct{})
		if err := d.findParent(ctx,
			nodeMap,
			targetSourceMap,
			node,
			parentNodeMap); err != nil {

			return err
		}
		d.dependencies[node] = parentNodeMap
	}
	return nil
}

func (d *dagEngine) findParent(ctx context.Context,
	nodeMap map[uuid.UUID]*model.WorkflowNode,
	edgeMap map[uuid.UUID]uuid.UUID,
	node *model.WorkflowNode, parentMap map[*model.WorkflowNode]struct{}) error {
	if node == nil {
		return nil
	}

	sourceNodeUUID, ok := edgeMap[node.UUID]
	if !ok {
		return nil
	}

	parentNode, ok := nodeMap[sourceNodeUUID]
	if !ok {
		return nil
	}

	if _, ok := parentMap[parentNode]; ok {
		return code.WorkflowHasCircularErr
	}

	parentMap[parentNode] = struct{}{}
	return d.findParent(ctx, nodeMap, edgeMap, parentNode, parentMap)
}

func (d *dagEngine) Run(ctx context.Context, job *engine.WorkflowInfo) error {
	d.job = job

	if err := d.loadData(ctx); err != nil {
		return err
	}

	if err := d.buildTask(ctx); err != nil {
		return err
	}

	return d.runAllNodes(ctx)
}

func (d *dagEngine) runAllNodes(ctx context.Context) error {
	var hasError atomic.Bool
	var firstError atomic.Value
	closeCtx, cancle := context.WithCancel(ctx)
	defer cancle()

	for {
		if len(d.dependencies) == 0 {
			return nil
		}

		select {
		case <-closeCtx.Done():
			return nil
		default:
			time.Sleep(100 * time.Millisecond)
		}

		canRunNodes := make([]*model.WorkflowNode, 0, 10)
		nodeJobs := make([]*model.WorkflowNodeJob, 0, 10)
		for node, nodeDependences := range d.dependencies {
			if len(nodeDependences) == 0 {
				canRunNodes = append(canRunNodes, node)
				nodeJobs = append(nodeJobs, &model.WorkflowNodeJob{
					LabID:  d.job.LabData.ID,
					NodeID: node.ID,
					UserID: d.job.UserID,
					Status: model.WorkflowJobPending,
				})
			}
		}

		if err := d.workflowStore.CreateJobs(closeCtx, nodeJobs); err != nil {
			return err
		}

		for index, node := range canRunNodes {
			newNode := node
			d.pools.Submit(func() {
				defer d.wg.Done()
				d.wg.Add(1)

				utils.SafelyRun(func() {
					if err := d.runNode(closeCtx, newNode, nodeJobs[index]); err != nil {
						logger.Errorf(closeCtx, "node run fail node id: %d, errr: %+v", newNode.ID, err)
						if hasError.Load() {
							firstError.Store(err)
							hasError.Store(true)
							cancle()
						}
					}
				})

			})
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
		Header: node.ActionName,
		Status: "running",
		Type:   "info",
		Msg:    []string{"running node"},
	})
	if err := d.sendAction(ctx, node, job); err != nil {
		return err
	}

	return d.callbackAction(ctx, job)
}

func (d *dagEngine) sendAction(_ context.Context, node *model.WorkflowNode, job *model.WorkflowNodeJob) error {
	param := node.Param

	if d.session.IsClosed() {
		return code.EdgeConnectClosedErr
	}
	// 数据库插入数据

	data := engine.SendActionData{
		DeviceID:   *node.DeviceName,
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

func (d *dagEngine) Stop(ctx context.Context) error {
	d.Cancel()
	if d.pools != nil {
		d.pools.Release()
	}

	d.wg.Wait()
	return nil
}

func (d *dagEngine) GetStatus(ctx context.Context) error {

	return nil
}
