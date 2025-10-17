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

// @Summary      创建实验室
// @Description  创建一个新的实验室环境
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        req  body      environment.LaboratoryEnvReq  true  "实验室创建请求"
// @Success      200  {object}  common.Resp{data=environment.LaboratoryEnvResp}
// @Failure      400  {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab [post]
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

// @Summary      更新实验室
// @Description  更新实验室环境信息
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        req  body      environment.UpdateEnvReq  true  "实验室更新请求"
// @Success      200  {object}  common.Resp{data=environment.LaboratoryResp}
// @Failure      400  {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab [patch]
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

// @Summary      获取实验室列表
// @Description  获取当前用户的所有实验室
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        page      query     int  false  "页码"
// @Param        page_size query     int  false  "每页大小"
// @Success      200       {object}  common.Resp{data=common.PageMoreResp[[]environment.LaboratoryResp]}
// @Failure      400       {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/list [get]
func (l *EnvHandle) LabList(ctx *gin.Context) {
	req := &common.PageReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	req.Normalize()

	resp, err := l.envService.LabList(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "LabList err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary      创建实验室资源
// @Description  从边缘侧创建资源
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        req  body      environment.ResourceReq  true  "资源创建请求"
// @Success      200  {object}  common.Resp
// @Failure      400  {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/resource [post]
// 创建注册表
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

// @Summary      获取实验室成员
// @Description  根据实验室获取当前实验室成员
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        lab_uuid  path      string  true  "实验室UUID"
// @Param        page      query     int     false  "页码"
// @Param        page_size query     int     false  "每页大小"
// @Success      200       {object}  common.Resp{data=common.PageResp[[]environment.LabMemberResp]}
// @Failure      400       {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/member/{lab_uuid} [get]
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

	resp, err := l.envService.LabMemberList(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "GetLabMemeber err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

// @Summary      删除实验室成员
// @Description  删除实验室成员
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        lab_uuid     path      string  true  "实验室UUID"
// @Param        member_uuid  path      string  true  "成员UUID"
// @Success      200          {object}  common.Resp
// @Failure      400          {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/member/{lab_uuid}/{member_uuid} [delete]
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

// @Summary      创建邀请链接
// @Description  创建实验室邀请链接
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        lab_uuid  path      string  true  "实验室UUID"
// @Success      200       {object}  common.Resp{data=environment.InviteResp}
// @Failure      400       {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/invite/{lab_uuid} [post]
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

// @Summary      接受邀请链接
// @Description  接受实验室邀请链接
// @Tags         Laboratory
// @Accept       json
// @Produce      json
// @Param        uuid  path      string  true  "邀请UUID"
// @Success      200   {object}  common.Resp
// @Failure      400   {object}  common.Resp
// @Security     BearerAuth
// @Router       /lab/invite/{uuid} [get]
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
