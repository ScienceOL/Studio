package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/gorm/schema"
)

type ResNodeTpl struct {
	Node    *model.ResourceNodeTemplate
	Actions []*model.WorkflowNodeTemplate
}

type LaboratoryRepo interface {
	UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids ...uuid.UUID) map[uuid.UUID]int64
	ID2UUID(ctx context.Context, tableModel schema.Tabler, ids ...int64) map[int64]uuid.UUID
	FindDatas(ctx context.Context, datas any, condition map[string]any, keys ...string) error
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
	Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error)
	DelData(ctx context.Context, tableModel schema.Tabler, condition map[string]any) error

	CreateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	// 根据 uuid 获取实验室
	GetLabByUUID(ctx context.Context, UUID uuid.UUID, selectKeys ...string) (*model.Laboratory, error)
	// 根据实验室ID获取实验室
	GetLabByID(ctx context.Context, labID int64, selectKeys ...string) (*model.Laboratory, error)
	// 根据实验室用户 AK、SK 获取
	GetLabByAkSk(ctx context.Context, accessKey string, accessSecret string) (*model.Laboratory, error)
	// 更新实验室环境
	UpdateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	// 更新或者插入设备模板
	UpsertResourceNodeTemplate(ctx context.Context, datas []*model.ResourceNodeTemplate) error
	// 工作流节点模板
	UpsertWorkflowNodeTemplate(ctx context.Context, datas []*model.WorkflowNodeTemplate) error
	// 更新或者插入设备模板 handle
	UpsertResourceHandleTemplate(ctx context.Context, data []*model.ResourceHandleTemplate) error
	// 根据 device template node id 获取所有的 handle
	GetResourceHandleTemplates(ctx context.Context, resourceNodeIDs []int64) (map[int64][]*model.ResourceHandleTemplate, error)
	// 根据 device template node id 获取所有的 uuid
	GetResourceNodeTemplates(ctx context.Context, ids []int64) ([]*model.ResourceNodeTemplate, error)
	// 根据实验室 id 获取所有的模板信息
	GetAllResourceTemplateByLabID(ctx context.Context, labID int64, selectKeys ...string) ([]*model.ResourceNodeTemplate, error)
	// 获取实验室列表
	GetLabList(ctx context.Context, userIDs []string, req *common.PageReq) (*common.PageResp[[]*model.Laboratory], error)
	// 创建 action handle
	UpsertActionHandleTemplate(ctx context.Context, datas []*model.WorkflowHandleTemplate) error
	// 获取所有驱动名称
	GetAllResourceName(ctx context.Context, labID int64) []string
	// 增加实验室成员
	AddLabMemeber(ctx context.Context, datas ...*model.LaboratoryMember) error
	// 获取该用户的所有实验室
	GetLabByUserID(ctx context.Context, req *common.PageReqT[string]) (*common.PageResp[[]*model.LaboratoryMember], error)
	// 根据实验室获取成员
	GetLabByLabID(ctx context.Context, req *common.PageReqT[int64]) (*common.PageResp[[]*model.LaboratoryMember], error)
}
