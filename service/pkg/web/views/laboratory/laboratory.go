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

func (l *EnvHandle) LabList(ctx *gin.Context) {
	req := &common.PageReq{}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Errorf(ctx, "parse body err: %+v", err)
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}

	resp, err := l.envService.LabList(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "LabList err: %+v", err)
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, resp)
}

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
