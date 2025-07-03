package views

import (
	"github.com/gin-gonic/gin"
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
	// 处理业务参数
	resp, err := l.service.GetEnvs(ctx)
	// 处理返回结果
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, resp)
}
