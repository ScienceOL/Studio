package environment

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
)

type EnvService interface {
	CreateLaboratoryEnv(ctx context.Context, req *LaboratoryEnvReq) (*LaboratoryEnvResp, error)
	UpdateLaboratoryEnv(ctx context.Context, req *UpdateEnvReq) (*UpdateEnvResp, error)
	CreateResource(ctx context.Context, req *ResourceReq) error
	LabList(ctx context.Context, req *common.PageReq) (*common.PageResp, error)
}
