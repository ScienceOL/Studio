package laboratory

import (
	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/environment"
	"github.com/scienceol/studio/service/pkg/core/environment/laboratory"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type EnvHandle struct {
	envService environment.EnvService
}

func NewEnvironment() *EnvHandle {
	return &EnvHandle{
		envService: laboratory.NewLab(),
	}
}

// @Summary 	创建实验室环境
// @Description 创建一个新的实验室环境
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		lab body environment.LaboratoryEnvReq true "实验室环境创建请求"
// @Success 	200 {object} common.Resp{data=environment.LaboratoryEnvResp} "创建成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab [post]
func (l *EnvHandle) CreateLabEnv(ctx *gin.Context) {
	req := &environment.LaboratoryEnvReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := l.envService.CreateLaboratoryEnv(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "CreateLaboratoryEnv err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary 	更新实验室环境
// @Description 更新一个已存在的实验室环境
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		lab body environment.UpdateEnvReq true "实验室环境更新请求"
// @Success 	200 {object} common.Resp{data=environment.LaboratoryResp} "更新成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab [patch]
func (l *EnvHandle) UpdateLabEnv(ctx *gin.Context) {
	req := &environment.UpdateEnvReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := l.envService.UpdateLaboratoryEnv(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "CreateLaboratoryEnv err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary 	删除实验室环境
// @Description 删除一个实验室环境
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		lab body environment.DelLabReq true "删除实验室请求"
// @Success 	200 {object} common.Resp{} "删除成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab [delete]
func (l *EnvHandle) DelLabEnv(ctx *gin.Context) {
	req := &environment.DelLabReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	err := l.envService.DelLab(ctx, req)
	common.Reply(ctx, err)
}

// @Summary 	获取实验室列表
// @Description 获取当前用户的所有实验室
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		page query common.PageReq false "分页参数"
// @Success 	200 {object} common.Resp{data=environment.LaboratoryListResp} "获取成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/list [get]
func (l *EnvHandle) LabList(ctx *gin.Context) {
	req := &common.PageReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	// 统一规范化分页参数（设置默认值）
	req.Normalize()

	resp, err := l.envService.LabList(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "LabList err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary 	获取实验室信息
// @Description 获取单个实验室的详细信息
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		lab_uuid path string true "实验室UUID"
// @Param 		with_member query bool false "是否包含成员列表"
// @Success 	200 {object} common.Resp{data=environment.LabInfoResp} "获取成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/{lab_uuid} [get]
func (l *EnvHandle) LabInfo(ctx *gin.Context) {
	req := &environment.LabInfoReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := l.envService.LabInfo(ctx, req)
	common.Reply(ctx, err, resp)
}

// @Summary 	创建实验室资源
// @Description 从边缘端创建实验室资源
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		resource body environment.ResourceReq true "资源请求"
// @Success 	200 {object} common.Resp{} "创建成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/resource [post]
func (l *EnvHandle) CreateLabResource(ctx *gin.Context) {
	req := &environment.ResourceReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	err := l.envService.CreateResource(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "CreateLabResource err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx)
}

// @Summary 	获取实验室成员
// @Description 获取指定实验室的成员列表
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		lab_uuid path string true "实验室UUID"
// @Param 		page query common.PageReq false "分页参数"
// @Success 	200 {object} common.Resp{data=environment.LabMemberListResp} "获取成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/member/{lab_uuid} [get]
func (l *EnvHandle) GetLabMemeber(ctx *gin.Context) {
	req := &environment.LabMemberReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	if err := ctx.ShouldBindQuery(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	// 统一规范化分页参数（设置默认值）
	req.Normalize()

	resp, err := l.envService.LabMemberList(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "GetLabMemeber err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary 	删除实验室成员
// @Description 从实验室中删除一个成员
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		lab_uuid path string true "实验室UUID"
// @Param 		member_uuid path string true "成员UUID"
// @Success 	200 {object} common.Resp{} "删除成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/member/{lab_uuid}/{member_uuid} [delete]
func (l *EnvHandle) DelLabMember(ctx *gin.Context) {
	req := &environment.DelLabMemberReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	err := l.envService.DelLabMember(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "DelLabMember err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx)
}

// @Summary 	创建邀请链接
// @Description 为实验室创建邀请链接
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		lab_uuid path string true "实验室UUID"
// @Success 	200 {object} common.Resp{data=environment.InviteResp} "创建成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/invite/{lab_uuid} [post]
func (l *EnvHandle) CreateInvite(ctx *gin.Context) {
	req := &environment.InviteReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := l.envService.CreateInvite(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "CreateInvite err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary 	接受邀请
// @Description 通过邀请链接加入实验室
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Param 		uuid path string true "邀请链接UUID"
// @Success 	200 {object} common.Resp{} "接受成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/invite/{uuid} [get]
func (l EnvHandle) AcceptInvite(ctx *gin.Context) {
	req := &environment.AcceptInviteReq{}
	if err := ctx.ShouldBindUri(req); err != nil {
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	err := l.envService.AcceptInvite(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "AcceptInvite err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx)
}

// @Summary 	获取用户信息
// @Description 获取当前用户信息
// @Tags 		Laboratory
// @Accept 		json
// @Produce 	json
// @Security 	BearerAuth
// @Success 	200 {object} common.Resp{data=model.UserData} "获取成功"
// @Failure 	200 {object} common.Resp{code=code.ErrCode} "请求参数错误"
// @Router 		/v1/lab/user [get]
func (l *EnvHandle) UserInfo(ctx *gin.Context) {
	resp, err := l.envService.UserInfo(ctx)
	common.Reply(ctx, err, resp)
}
