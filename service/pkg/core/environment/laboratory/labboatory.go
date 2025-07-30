package laboratory

import (
	"context"
	"time"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/environment"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type lab struct {
	envStore repo.EnvRepo
}

func NewLab() environment.EnvService {
	return &lab{
		envStore: eStore.NewEnv(),
	}
}

func (lab *lab) CreateLaboratoryEnv(ctx context.Context, req *environment.LaboratoryEnvReq) (*environment.LaboratoryEnvResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	data := &model.Laboratory{
		Name:        req.Name,
		UserID:      userInfo.ID,
		Status:      model.INIT,
		Description: req.Description,
		BaseModel: model.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	if err := lab.envStore.CreateLaboratoryEnv(ctx, data); err != nil {
		return nil, err
	}

	return &environment.LaboratoryEnvResp{
		UUID: data.UUID,
		Name: data.Name,
	}, nil
}

func (lab *lab) UpdateLaboratoryEnv(ctx context.Context, req *environment.UpdateEnvReq) (*environment.UpdateEnvResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	data := &model.Laboratory{
		BaseModel: model.BaseModel{
			UUID:      req.UUID,
			UpdatedAt: time.Now(),
		},
		Name:        req.Name,
		UserID:      userInfo.ID,
		Description: req.Description,
	}

	err := lab.envStore.UpdateLaboratoryEnv(ctx, data)
	if err != nil {
		return nil, err
	}
	return &environment.UpdateEnvResp{
		UUID:        data.UUID,
		Name:        data.Name,
		Description: data.Description,
	}, nil
}

func (lab *lab) CreateReg(ctx context.Context, req *environment.RegistryReq) error {
	userIfo := auth.GetCurrentUser(ctx)
	if userIfo == nil {
		return code.UnLogin
	}
	labData, err := lab.envStore.GetLabByUUID(ctx, req.LabUUID)
	if err != nil {
		return err
	}

	// 处理 action

	return db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		for _, reg := range req.Registries {
			regData := &model.Registry{
				Name:   reg.RegName,
				LabID:  labData.ID,
				Status: model.REGINIT,
				Module: reg.Class.Module,
				// FIXME: 未查询到数据从哪获取
				// Model      : datatypes.JSON(req.Class.Module), // @世xiang
				Type:         reg.Class.Type,
				RegsitryType: reg.RegistryType,
				Version:      reg.Version,
				StatusTypes:  reg.Class.StatusTypes,
				Icon:         reg.Icon,
				Description:  reg.Description,
			}

			if err := lab.envStore.CreateReg(txCtx, regData); err != nil {
				return err
			}

			actions := make([]*model.RegAction, 0, len(reg.Class.ActionValueMappings))
			for actionName, action := range reg.Class.ActionValueMappings {
				if actionName == "" {
					return code.RegActionNameEmptyErr
				}

				actions = append(actions, &model.RegAction{
					RegID:       regData.ID,
					Name:        actionName,
					Goal:        action.Goal,
					GoalDefault: action.GoalDefault,
					Feedback:    action.Feedback,
					Result:      action.Result,
					Schema:      action.Schema,
					Type:        action.Type,
					Handles:     action.Handles,
				})
			}

			if err := lab.envStore.UpsertRegAction(txCtx, actions); err != nil {
				return err
			}

			deviceData := &model.DeviceNodeTemplate{
				Name:        reg.RegName,
				LabID:       labData.ID,
				RegID:       regData.ID,
				UserID:      userIfo.ID,
				Header:      regData.Name,
				Footer:      "",
				Version:     regData.Version,
				Icon:        regData.Icon,
				Description: regData.Description,
			}
			if err := lab.envStore.UpsertDeviceTemplate(txCtx, deviceData); err != nil {
				return err
			}

			handles := make([]*model.DeviceNodeHandleTemplate, 0, len(reg.Handles))
			for _, handle := range reg.Handles {
				handles = append(handles, &model.DeviceNodeHandleTemplate{
					NodeID:      deviceData.ID,
					Name:        handle.HandlerKey,
					DisplayName: handle.Label,
					Type:        handle.DataType,
					IOType:      handle.IoType,
					Source:      handle.DataSource,
					Key:         handle.DataKey,
					Side:        handle.Side,
				})
			}
			if err := lab.envStore.UpsertDeviceHandleTemplate(txCtx, handles); err != nil {
				return err
			}

			deviceSchemas := make([]*model.DeviceNodeParamTemplate, 0, 2)
			if reg.InitParamSchema.Data != nil {
				deviceSchemas = append(deviceSchemas, &model.DeviceNodeParamTemplate{
					NodeID:      deviceData.ID,
					Name:        "data",
					Type:        "DEFAULT",
					Placeholder: "设备初始化参数配置",
					Schema:      reg.InitParamSchema.Data.Properties,
				})
			}

			if reg.InitParamSchema.Config != nil {
				deviceSchemas = append(deviceSchemas, &model.DeviceNodeParamTemplate{
					NodeID:      deviceData.ID,
					Name:        "config",
					Type:        "DEFAULT",
					Placeholder: "设备初始化参数配置",
					Schema:      reg.InitParamSchema.Config.Properties,
				})
			}
			if err := lab.envStore.UpsertDeviceParamTemplate(txCtx, deviceSchemas); err != nil {
				return err
			}
		}
		return nil
	})
}
