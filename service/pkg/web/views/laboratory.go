package views

import (
	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/service/laboratory"
)

type Lab struct {
	service *laboratory.Laboratory
}

func NewLabHandle() *Lab {
	return &Lab{
		service: laboratory.NewLaboratory(),
	}
}

// GetEnv 获取环境信息
// @Summary 获取环境信息
// @Description 获取实验室环境配置信息
// @Tags 实验室
// @Accept json
// @Produce json
// @Success 200 {object} common.Resp{data=laboratory.LaboratoryEnv}
// @Failure 200 {object} common.Resp
// @Router /api/v1/lab/env [get]
func (l *Lab) GetEnv(ctx *gin.Context) {
	resp, err := l.service.GetEnvs(ctx)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}
	common.ReplyOk(ctx, resp)
}
