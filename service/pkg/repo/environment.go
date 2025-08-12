package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type ResNodeTpl struct {
	Node    *model.ResourceNodeTemplate
	Actions []*model.DeviceAction
}

type LaboratoryRepo interface {
	// 创建实验室
	CreateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	// 根据 uuid 获取实验室
	GetLabByUUID(ctx context.Context, UUID uuid.UUID, selectKeys ...string) (*model.Laboratory, error)
	// 根据实验室用户 AK、SK 获取
	GetLabByAkSk(ctx context.Context, accessKey string, accessSecret string) (*model.Laboratory, error)
	// 更新实验室环境
	UpdateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	// 更新或者插入设备模板
	UpsertDeviceAction(ctx context.Context, datas []*model.DeviceAction) error
	// 更新或者插入设备模板
	UpsertDeviceTemplate(ctx context.Context, datas []*model.ResourceNodeTemplate) error
	// 更新或者插入设备模板 handle
	UpsertDeviceHandleTemplate(ctx context.Context, data []*model.ResourceHandleTemplate) error
	// 根据实验室获取所有的注册表信息
	GetResourceTemplate(ctx context.Context, labID int64, names []string) (map[string]*ResNodeTpl, error)
	// 根据 device template node id 获取所有的 handle
	GetResourceHandleTemplates(ctx context.Context, resIDs []int64) (map[int64][]*model.ResourceHandleTemplate, error)
	// 根据 device template node id 获取所有的 uuid
	GetResourceNodeTemplateUUID(ctx context.Context, resIDs []int64) (map[int64]uuid.UUID, error)
	// 根据实验室 id 获取所有的模板信息
	GetAllResourceTemplateByLabID(ctx context.Context, labID int64, selectKeys ...string) ([]*model.ResourceNodeTemplate, error)
	// 根据 device ids 获取所有的 handles
	GetAllDeviceTemplateHandlesByID(ctx context.Context, templateIDs []int64, selectKeys ...string) ([]*model.ResourceHandleTemplate, error)
	// 根据 uuid 获取 template 数据
	GetResourceTemplateByUUD(ctx context.Context, uuid uuid.UUID, selectKeys ...string) (*model.ResourceNodeTemplate, error)
	// 获取实验室列表
	GetLabList(ctx context.Context, userIDs []string, req *common.PageReq) (*common.PageResp[[]*model.Laboratory], error)
}
