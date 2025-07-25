package material

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/core/material"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/datatypes"
)

type materialImpl struct {
	envStore repo.EnvRepo
}

func NewMaterial() material.Service {
	return &materialImpl{
		envStore: eStore.NewEnv(),
	}
}

func (m *materialImpl) CreateMaterial(ctx context.Context, req []*material.Node) error {
	// FIXME: 需要在此逻辑位置创建节点模板，暂时不处理
	/// FIXME: 修复 lab 上传鉴权问题一下 uuid 暂时硬编码
	uuid := common.BinUUID(datatypes.BinUUIDFromString(""))
	labData, err := m.envStore.GetLabByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	mDatas := make([]*model.MaterialNode, 0, len(req))
	for _, data := range req {
		mDatas = append(mDatas, &model.MaterialNode{
			ParentID:             0,
			LabID:                labData.ID,
			Name:                 data.ID,
			DisplayName:          data.Name,
			Description:          data.Description,
			Status:               "idle",
			Type:                 data.Type,
			DeviceNodeTemplateID: 0,
			RegID:                0,
			// FIXME: 从注册表获取
			InitParamData: data.Config,
			// Schema              :
			Data: data.Data,
			// Dirs:
			Position: data.Position,
			// Pose                :
			Model: data.Model,
		})
		_ = mDatas
	}

	return nil
}
