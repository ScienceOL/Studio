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
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/datatypes"
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
	if len(req.Resources) == 0 {
		return code.ResourceIsEmptyErr
	}

	labInfo := auth.GetCurrentUser(ctx)
	if labInfo == nil {
		return code.UnLogin
	}
	labData, err := lab.envStore.GetLabByAkSk(ctx, labInfo.AccessKey, labInfo.AccessSecret)
	if err != nil {
		return err
	}

	return db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		resDatas := utils.FilterSlice(req.Resources, func(item environment.Resource) (*model.ResourceNodeTemplate, bool) {
			data := &model.ResourceNodeTemplate{
				Name:        item.RegName,
				LabID:       labData.ID,     // 实验室的 id
				UserID:      labData.UserID, // 创建实验室的 user id
				Header:      item.RegName,
				Footer:      "",
				Version:     utils.Or(item.Version, "0.0.1"),
				Icon:        item.Icon,
				Description: item.Description,
				Model:       item.Model,
				Module:      item.Class.Module,
				Language:    item.Language,
				StatusTypes: item.Class.StatusTypes,
				// DataSchema: utils.TernaryLazy(
				// 	item.InitParamSchema == nil || item.InitParamSchema.Data == nil,
				// 	func() datatypes.JSON { return datatypes.JSON{} },
				// 	func() datatypes.JSON { return item.InitParamSchema.Data.Properties }, // 安全！
				// ),
				DataSchema: utils.SafeValue(func() datatypes.JSON {
					return item.InitParamSchema.Data.Properties
				}, datatypes.JSON{}),
				ConfigSchema: utils.SafeValue(
					func() datatypes.JSON { return item.InitParamSchema.Config.Properties },
					datatypes.JSON{}),
				// Labels      :
			}
			return data, true
		})

		resDataMap := utils.SliceToMap(resDatas, func(item *model.ResourceNodeTemplate) (string, *model.ResourceNodeTemplate) {
			return item.Name, item
		})

		if err := lab.envStore.UpsertDeviceTemplate(txCtx, resDatas); err != nil {
			return err
		}

		// device actions
		resDeviceAction, err := utils.FilterSliceWithErr(req.Resources, func(item environment.Resource) ([]*model.DeviceAction, bool, error) {
			resData, ok := resDataMap[item.RegName]
			if !ok {
				return nil, false, code.ResourceNotExistErr
			}
			actions := make([]*model.DeviceAction, 0, len(item.Class.ActionValueMappings))
			for actionName, action := range item.Class.ActionValueMappings {
				if actionName == "" {
					return nil, false, code.RegActionNameEmptyErr
				}

				actions = append(actions, &model.DeviceAction{
					ResNodeID:   resData.ID,
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
			return actions, true, nil
		})
		if err != nil {
			return err
		}

		// device handles
		resDeviceHandles, err := utils.FilterSliceWithErr(req.Resources, func(item environment.Resource) ([]*model.ResourceHandleTemplate, bool, error) {
			resData, ok := resDataMap[item.RegName]
			if !ok {
				return nil, false, code.ResourceNotExistErr
			}
			handles := make([]*model.ResourceHandleTemplate, 0, len(item.Class.ActionValueMappings))
			for _, handle := range item.Handles {
				handles = append(handles, &model.ResourceHandleTemplate{
					NodeID:      resData.ID,
					Name:        handle.HandlerKey,
					DisplayName: handle.Label,
					Type:        handle.DataType,
					IOType:      handle.IoType,
					Source:      handle.DataSource,
					Key:         handle.DataKey,
					Side:        handle.Side,
				})
			}
			return handles, true, nil
		})
		if err != nil {
			return err
		}
		if err := lab.envStore.UpsertDeviceAction(txCtx, resDeviceAction); err != nil {
			return err
		}
		if err := lab.envStore.UpsertDeviceHandleTemplate(txCtx, resDeviceHandles); err != nil {
			return err
		}
		return nil
	})
}
