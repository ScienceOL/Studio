package workflow

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/scienceol/studio/service/pkg/core/workflow"
	impl "github.com/scienceol/studio/service/pkg/core/workflow/workflow"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type Handle struct {
	wsClient *melody.Melody
	wService workflow.Service
}

func NewWorkflowHandle(ctx context.Context) *Handle {
	wsClient := melody.New()
	wsClient.Config.MaxMessageSize = constant.MaxMessageSize
	// mService := impl.NewMaterial(wsClient)
	// 注册集群通知

	h := &Handle{
		wsClient: wsClient,
		wService: impl.New(ctx, wsClient),
	}

	h.initMaterialWebSocket()
	return h
}

// @Summary 节点模板列表
// @Description 获取实验室下的节点模板列表，支持按名称、标签和分页过滤
// @Tags Workflow
// @Accept json
// @Produce json
// @Param req query workflow.TplPageReq false "查询与分页参数"
// @Success 200 {object} common.Resp{data=TemplateListPage} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/node/template/list [get]
func (w *Handle) TemplateList(ctx *gin.Context) {

	req := workflow.TplPageReq{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	if res, err := w.wService.TemplateList(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 节点模板标签
// @Description 获取指定实验室的节点模板标签列表
// @Tags Workflow
// @Accept json
// @Produce json
// @Param lab_uuid path string true "实验室UUID"
// @Success 200 {object} common.Resp{data=[]string} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/node/template/tags/{lab_uuid} [get]
func (w *Handle) TemplateTags(ctx *gin.Context) {

	req := workflow.TemplateTagsReq{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	if res, err := w.wService.TemplateTags(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// 工作流模板详情
func (w *Handle) TemplateDetail(ctx *gin.Context) {}

// @Summary 工作流模板标签
// @Description 获取全局工作流模板标签列表
// @Tags Workflow
// @Accept json
// @Produce json
// @Success 200 {object} common.Resp{data=[]string} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/template/tags [get]
func (w *Handle) WorkflowTemplateTags(ctx *gin.Context) {

	if res, err := w.wService.WorkflowTemplateTags(ctx); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// WorkflowTemplateTagsByLab 按实验室获取工作流模板标签
func (w *Handle) WorkflowTemplateTagsByLab(ctx *gin.Context) {
	// @Summary 按实验室获取工作流模板标签
	// @Description 根据实验室UUID获取该实验室下可用的工作流模板标签
	// @Tags Workflow
	// @Accept json
	// @Produce json
	// @Param lab_uuid path string true "实验室UUID"
	// @Success 200 {object} common.Resp{data=[]string} "获取成功"
	// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
	// @Router /v1/lab/workflow/template/tags/{lab_uuid} [get]
	req := &workflow.TemplateTagsReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	if res, err := w.wService.WorkflowTemplateTagsByLab(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 工作流模板列表
// @Description 获取工作流模板列表，支持按标签与分页过滤
// @Tags Workflow
// @Accept json
// @Produce json
// @Param req query workflow.TemplateListReq false "查询与分页参数"
// @Success 200 {object} common.Resp{data=WorkflowTemplateListPage} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/template/list [get]
func (w *Handle) WorkflowTemplateList(ctx *gin.Context) {

	req := workflow.TemplateListReq{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.WorkflowTemplateList(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary Fork 工作流模板
// @Description 将已有工作流模板 Fork 到目标实验室
// @Tags Workflow
// @Accept json
// @Produce json
// @Param req query workflow.ForkReq true "Fork 请求参数"
// @Success 200 {object} common.Resp{} "操作成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/template/fork [put]
func (w *Handle) ForkTemplate(ctx *gin.Context) {

	req := &workflow.ForkReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if err := w.wService.ForkWrokflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx)
	}
}

// @Summary 工作流任务列表
// @Description 获取指定工作流的任务列表（滚动分页）
// @Tags Workflow
// @Accept json
// @Produce json
// @Param uuid path string true "工作流UUID"
// @Param req query common.PageReq false "分页参数"
// @Success 200 {object} common.Resp{data=TaskPageMore} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/task/{uuid} [get]
func (w *Handle) TaskList(ctx *gin.Context) {

	req := workflow.TaskReq{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.WorkflowTaskList(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 下载工作流任务
// @Description 下载指定工作流的任务列表为 CSV 文件
// @Tags Workflow
// @Accept json
// @Produce text/csv
// @Param uuid path string true "工作流UUID"
// @Success 200 {file} file "CSV 文件"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/task/download/{uuid} [get]
func (w *Handle) DownloadTask(ctx *gin.Context) {

	req := workflow.TaskDownloadReq{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.TaskDownload(ctx, &req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		ctx.Header("Content-Disposition", "attachment; filename=task.csv")
		ctx.Header("Content-Type", "text/csv")
		ctx.Header("Pragma", "public")
		ctx.Header("Content-Length", fmt.Sprintf("%d", len(res.Bytes())))

		// 发送文件数据
		ctx.Data(http.StatusOK, "text/csv", res.Bytes())
	}
}

// 节点模板列表，节点模板分类

// @Summary 节点模板详情
// @Description 获取节点模板的详细信息
// @Tags Workflow
// @Accept json
// @Produce json
// @Param uuid path string true "模板UUID"
// @Success 200 {object} common.Resp{data=workflow.NodeTemplateDetailResp} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/node/template/detail/{uuid} [get]
func (w *Handle) NodeTemplateDetail(ctx *gin.Context) {

	req := &workflow.NodeTemplateReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("template uuid is empty"))
		return
	}

	if res, err := w.wService.NodeTemplateDetail(ctx, req.UUID); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 更新工作流
// @Description 更新我创建的工作流（名称、发布状态、描述等）
// @Tags Workflow
// @Accept json
// @Produce json
// @Param workflow body workflow.UpdateReq true "工作流更新请求"
// @Success 200 {object} common.Resp{} "更新成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/owner [patch]
func (w *Handle) UpdateWorkflow(ctx *gin.Context) {

	req := &workflow.UpdateReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("workflow uuid is empty"))
		return
	}

	if err := w.wService.UpdateWorkflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx)
	}
}

// @Summary 删除工作流
// @Description 删除我创建的工作流
// @Tags Workflow
// @Accept json
// @Produce json
// @Param uuid path string true "工作流UUID"
// @Success 200 {object} common.Resp{} "删除成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/owner/{uuid} [delete]
func (w *Handle) DelWorkflow(ctx *gin.Context) {

	req := &workflow.DelReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if err := w.wService.DelWorkflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx)
	}
}

// @Summary 创建工作流
// @Description 在指定实验室创建新的工作流
// @Tags Workflow
// @Accept json
// @Produce json
// @Param workflow body workflow.CreateReq true "工作流创建请求"
// @Success 200 {object} common.Resp{data=workflow.CreateResp} "创建成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/owner [post]
func (w *Handle) Create(ctx *gin.Context) {

	req := &workflow.CreateReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.Create(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 工作流列表
// @Description 获取我创建的工作流列表（滚动加载）
// @Tags Workflow
// @Accept json
// @Produce json
// @Param req query workflow.ListReq false "查询与分页参数"
// @Success 200 {object} common.Resp{data=workflow.ListResult} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/owner/list [get]
func (w *Handle) GetWorkflowList(ctx *gin.Context) {

	req := &workflow.ListReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if res, err := w.wService.GetWorkflowList(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 工作流详情
// @Description 获取工作流的节点与边详情
// @Tags Workflow
// @Accept json
// @Produce json
// @Param uuid path string true "工作流UUID"
// @Success 200 {object} common.Resp{data=workflow.DetailResp} "获取成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/template/detail/{uuid} [get]
func (w *Handle) GetWorkflowDetail(ctx *gin.Context) {

	req := &workflow.DetailReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("workflow uuid is empty"))
		return
	}

	if res, err := w.wService.GetWorkflowDetail(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 导出工作流
// @Description 导出工作流为可跨实验室导入的 JSON 数据
// @Tags Workflow
// @Accept json
// @Produce json
// @Param req query workflow.ExportReq true "导出请求参数"
// @Success 200 {object} common.Resp{data=workflow.ExportData} "导出成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/owner/export [get]
func (w *Handle) Export(ctx *gin.Context) {

	req := &workflow.ExportReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	if req.UUID.IsNil() {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("workflow uuid is empty"))
		return
	}
	if res, err := w.wService.ExportWorkflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

// @Summary 导入工作流
// @Description 将导出的工作流 JSON 导入到目标实验室
// @Tags Workflow
// @Accept json
// @Produce json
// @Param workflow body workflow.ImportReq true "导入请求"
// @Success 200 {object} common.Resp{data=workflow.CreateResp} "导入成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/owner/import [post]
func (w *Handle) Import(ctx *gin.Context) {

	req := &workflow.ImportReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	if req.TargetLabUUID.IsNil() || req.Data == nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("target_lab_uuid or data is empty"))
		return
	}
	if res, err := w.wService.ImportWorkflow(ctx, req); err != nil {
		common.ReplyErr(ctx, err)
	} else {
		common.ReplyOk(ctx, res)
	}
}

func (w *Handle) initMaterialWebSocket() {
	w.wsClient.HandlePong(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "==================== pong=====================")
		}
	})
	w.wsClient.HandleClose(func(s *melody.Session, _ int, _ string) error {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "client close keys: %+v", s.Keys)
		}
		return nil
	})

	w.wsClient.HandleDisconnect(func(s *melody.Session) {
		if ctx, ok := s.Get("ctx"); ok {
			logger.Infof(ctx.(context.Context), "client closed keys: %+v", s.Keys)
		}
	})

	w.wsClient.HandleError(func(s *melody.Session, err error) {
		if errors.Is(err, melody.ErrMessageBufferFull) {
			return
		}

		if closeErr, ok := err.(*websocket.CloseError); ok {
			if closeErr.Code == websocket.CloseGoingAway {
				return
			}
		}

		if strings.Contains(err.Error(), "use of closed network connection") {
			return
		}

		if ctx, ok := s.Get("ctx"); ok {
			logger.Errorf(ctx.(context.Context), "websocket find keys: %+v, err: %+v", s.Keys, err)
		}
	})

	w.wsClient.HandleConnect(func(s *melody.Session) {
		c, _ := s.Get("ctx")
		if err := w.wService.OnWSConnect(c.(*gin.Context), s); err != nil {
			logger.Errorf(c.(*gin.Context), "check param err: %+v", err)
			if err := s.CloseWithMsg([]byte(err.Error())); err != nil {
				logger.Errorf(c.(*gin.Context), "workflow HandleConnect CloseWithMsg err: %+v", err)
			}
		}
	})

	w.wsClient.HandleMessage(func(s *melody.Session, b []byte) {
		c, _ := s.Get("ctx")
		if err := w.wService.OnWSMsg(c.(*gin.Context), s, b); err != nil {
			logger.Errorf(c.(*gin.Context), "material handle msg err: %+v", err)
		}
	})

	w.wsClient.HandleSentMessage(func(_ *melody.Session, _ []byte) {
		// 发送完消息后的回调
	})

	w.wsClient.HandleSentMessageBinary(func(_ *melody.Session, _ []byte) {
		// 发送完二进制消息后的回调
		// 如果发送的是字符串消息，上面的 HandleSentMessage 也会被回调
	})
}

// @Summary 工作流 WebSocket
// @Description 连接到指定工作流的 WebSocket 会话，用于实时编辑与运行
// @Tags Workflow
// @Accept json
// @Produce json
// @Param uuid path string true "工作流UUID"
// @Success 200 {object} common.Resp{} "连接成功（协议升级）"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/ws/workflow/{uuid} [get]
func (w *Handle) LabWorkflow(ctx *gin.Context) {

	req := &workflow.NodeTemplateReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		logger.Errorf(ctx, "unmarshal uuid err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	if req.UUID.IsNil() {
		logger.Errorf(ctx, "unmarshal uuid is empty")
		common.ReplyErr(ctx, code.ParamErr.WithMsg("uuid is empty"))
		return
	}

	// 阻塞运行
	if err := w.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		"uuid": req.UUID,
		"ctx":  ctx,
	}); err != nil {
		logger.Errorf(ctx, "workflow HandleRequestWithKeys err: %+v", err)
	}
}

// @Summary 复制工作流
// @Description 在目标实验室中复制现有工作流（自动匹配节点模板）
// @Tags Workflow
// @Accept json
// @Produce json
// @Param workflow body workflow.DuplicateReq true "复制请求"
// @Success 200 {object} common.Resp{data=workflow.DuplicateRes} "复制成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/workflow/owner/duplicate [put]
func (w *Handle) Duplicate(ctx *gin.Context) {

	req := &workflow.DuplicateReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	res, err := w.wService.DuplicateWorkflow(ctx, req)
	common.Reply(ctx, err, res)
}

// @Summary 启动工作流（无鉴权）
// @Description 通过 HTTP 启动工作流任务，返回任务 UUID
// @Tags Workflow
// @Accept json
// @Produce json
// @Param workflow body workflow.RunReq true "启动请求"
// @Success 200 {object} common.Resp{data=uuid.UUID} "启动成功"
// @Failure 200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router /v1/lab/run/workflow [put]
func (w *Handle) RunWorkflow(ctx *gin.Context) {

	req := &workflow.RunReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}
	taskUUID, err := w.wService.HttpRunWorkflow(ctx, req)
	common.Reply(ctx, err, taskUUID)
}
