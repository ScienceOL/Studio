package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/repo/model"
)

type RegDeviceInfo struct {
	RegName              string
	RegID                int64
	DeviceNodeTemplateID int64
}

type EnvRepo interface {
	CreateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	GetLabByUUID(ctx context.Context, UUID common.BinUUID) (*model.Laboratory, error)
	UpdateLaboratoryEnv(ctx context.Context, data *model.Laboratory) error
	CreateReg(ctx context.Context, data *model.Registry) error
	UpsertRegAction(ctx context.Context, datas []*model.RegAction) error
	UpsertDeviceTemplate(ctx context.Context, data *model.DeviceNodeTemplate) error
	UpsertDeviceHandleTemplate(ctx context.Context, data []*model.DeviceNodeHandleTemplate) error
	UpsertDeviceParamTemplate(ctx context.Context, data []*model.DeviceNodeParamTemplate) error
	GetRegs(ctx context.Context, labID int64, names []string) (map[string]*RegDeviceInfo, error)
	GetDeviceTemplateHandels(ctx context.Context, deviceIDs []int64) (map[int64][]*model.DeviceNodeHandleTemplate, error)
}
