package laboratory

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/core/environment"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/casdoor"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type lab struct {
	envStore      repo.EnvRepo
	accountClient repo.Account
}

func NewLab() environment.EnvService {
	return &lab{
		envStore:      eStore.NewEnv(),
		accountClient: casdoor.NewCasClient(),
	}
}

func (lab *lab) CreateLaboratoryEnv(ctx context.Context, req *environment.LaboratoryEnvReq) (*environment.LaboratoryEnvResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	ak := uuid.Must(uuid.NewV4()).String()
	sk := uuid.Must(uuid.NewV4()).String()
	err := lab.accountClient.CreateLabUser(ctx, &model.LabInfo{
		AccessKey:         ak,
		AccessSecret:      sk,
		Name:              fmt.Sprintf("%s-%s", req.Name, ak),
		DisplayName:       req.Name,
		Avatar:            "https://cdn.casbin.org/img/casbin.svg",
		Owner:             "scienceol",
		Type:              model.LABTYPE,
		Password:          "lab-user",
		SignupApplication: "scienceol",
	})
	if err != nil {
		return nil, err
	}

	data := &model.Laboratory{
		Name:         req.Name,
		UserID:       userInfo.ID,
		Status:       model.INIT,
		AccessKey:    ak,
		AccessSecret: sk,
		Description:  req.Description,
		BaseModel: model.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	if err := lab.envStore.CreateLaboratoryEnv(ctx, data); err != nil {
		// FIXME: 如果创建实验室失败，则删除对应的实验室用户
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

func (lab *lab) CreateResource(ctx context.Context, req *environment.ResourceReq) error {
	labInfo := auth.GetCurrentUser(ctx)
	if labInfo == nil {
		return code.UnLogin
	}
	labData, err := lab.envStore.GetLabByAkSk(ctx, labInfo.AccessKey, labInfo.AccessSecret)
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

			actions := make([]*model.DeviceAction, 0, len(reg.Class.ActionValueMappings))
			for actionName, action := range reg.Class.ActionValueMappings {
				if actionName == "" {
					return code.RegActionNameEmptyErr
				}

				actions = append(actions, &model.DeviceAction{
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

			deviceData := &model.ResourceNodeTemplate{
				Name:        reg.RegName,
				LabID:       labData.ID,
				RegID:       regData.ID,
				UserID:      labInfo.ID,
				Header:      regData.Name,
				Footer:      "",
				Version:     regData.Version,
				Icon:        regData.Icon,
				Description: regData.Description,
			}
			if err := lab.envStore.UpsertDeviceTemplate(txCtx, deviceData); err != nil {
				return err
			}

			handles := make([]*model.ResourceNodeHandle, 0, len(reg.Handles))
			for _, handle := range reg.Handles {
				handles = append(handles, &model.ResourceNodeHandle{
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

			deviceSchemas := make([]*model.ResourceNodeParam, 0, 2)
			if reg.InitParamSchema.Data != nil {
				deviceSchemas = append(deviceSchemas, &model.ResourceNodeParam{
					NodeID:      deviceData.ID,
					Name:        "data",
					Type:        "DEFAULT",
					Placeholder: "设备初始化参数配置",
					Schema:      reg.InitParamSchema.Data.Properties,
				})
			}

			if reg.InitParamSchema.Config != nil {
				deviceSchemas = append(deviceSchemas, &model.ResourceNodeParam{
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
