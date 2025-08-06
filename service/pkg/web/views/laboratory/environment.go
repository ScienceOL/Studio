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
