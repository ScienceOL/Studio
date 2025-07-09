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

func (l *Lab) GetEnv(ctx *gin.Context) {
	resp, err := l.service.GetEnvs(ctx)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}
	common.ReplyOk(ctx, resp)
}
