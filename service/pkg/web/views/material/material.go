package material

import (
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/material"
	impl "github.com/scienceol/studio/service/pkg/core/material/material"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type Handle struct {
	mService material.Service
}

func NewMaterialHandle() *Handle {
	return &Handle{
		mService: impl.NewMaterial(),
	}
}

func (m *Handle) CreateLabMaterial(ctx *gin.Context) {
	reqs := make([]*material.Node, 0, 1)
	if err := ctx.ShouldBindJSON(&reqs); err != nil {
		logger.Errorf(ctx, "parse CreateLabMaterial param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	if err := m.mService.CreateMaterial(ctx, reqs); err != nil {
		logger.Errorf(ctx, "CreateMaterial err: %+v", err)
		common.ReplyErr(ctx, err)

		return
	}

	common.ReplyOk(ctx)
}

func (m *Handle) CreateMaterialEdge(ctx *gin.Context) {
	reqs := make([]*material.Edge, 0, 1)
	if err := ctx.ShouldBindJSON(&reqs); err != nil {
		logger.Errorf(ctx, "parse CreateMaterialEdge param err: %+v", err.Error())
		common.ReplyErr(ctx, code.ParamErr, err.Error())
		return
	}
	if err := m.mService.CreateEdge(ctx, reqs); err != nil {
		logger.Errorf(ctx, "CreateMaterialEdge err: %+v", err)
		common.ReplyErr(ctx, err)

		return
	}

	common.ReplyOk(ctx)
}

func (m *Handle) LabMaterial(ctx *gin.Context) {
	// https://github.com/googollee/go-socket.io
	wsMelody := melody.New()
	wsMelody.BroadcastMultiple
}
