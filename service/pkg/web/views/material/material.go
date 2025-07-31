package material

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/material"
	impl "github.com/scienceol/studio/service/pkg/core/material/material"
	"github.com/scienceol/studio/service/pkg/core/notify"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type Handle struct {
	mService material.Service
}

func NewMaterialHandle() *Handle {
	mService := impl.NewMaterial()
	events.NewEvents().Registry(context.Background(), notify.MaterialModify, mService.HandleNotify)
	return &Handle{
		mService: mService,
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
	wsMelody := melody.New()
	melody.New()
	_ = wsMelody
}
