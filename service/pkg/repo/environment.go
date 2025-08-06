package repo

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type RegDeviceInfo struct {
	RegName              string
	RegID                int64
	DeviceNodeTemplateID int64
	Icon                 string
}

type EnvRepo interface {
	// 创建实验室
	CreateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	// 根据 uuid 获取实验室
	GetLabByUUID(ctx context.Context, UUID uuid.UUID, selectKeys ...string) (*model.Laboratory, error)
	// 根据实验室用户 AK、SK 获取
	GetLabByAkSk(ctx context.Context, accessKey string, accessSecret string) (*model.Laboratory, error)
	// 更新实验室环境
	UpdateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	// 更新或者插入注册表
	CreateReg(ctx context.Context, data *model.Registry) error
	// 更新或者插入设备模板
	UpsertRegAction(ctx context.Context, datas []*model.DeviceAction) error
	// 更新或者插入设备模板
	UpsertDeviceTemplate(ctx context.Context, data *model.ResourceNodeTemplate) error
	// 更新或者插入设备模板 handle
	UpsertDeviceHandleTemplate(ctx context.Context, data []*model.ResourceNodeHandle) error
	// 更新或者插入设备模板参数
	UpsertDeviceParamTemplate(ctx context.Context, data []*model.ResourceNodeParam) error
	// 根据实验室获取所有的注册表信息
	GetRegs(ctx context.Context, labID int64, names []string) (map[string]*RegDeviceInfo, error)
	// 根据 device tempalte node id 获取所有的 handle
	GetDeviceTemplateHandels(ctx context.Context, deviceIDs []int64) (map[int64][]*model.ResourceNodeHandle, error)
	// 根据实验室 id 获取所有的模板信息
	GetAllDeviceTemplateByLabID(ctx context.Context, labID int64, selectKeys ...string) ([]*model.ResourceNodeTemplate, error)
	// 根据 device ids 获取所有的 handles
	GetAllDeviceTemplateHandlesByID(ctx context.Context, templateIDs []int64, selectKeys ...string) ([]*model.ResourceNodeHandle, error)
	// 根据实验室 id 获取所有的模板信息
	GetAllDeviceTemplateParamByID(ctx context.Context, templateIDs []int64, selectKeys ...string) ([]*model.ResourceNodeParam, error)
}
