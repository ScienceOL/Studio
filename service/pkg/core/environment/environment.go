package environment

import "context"

type EnvService interface {
	CreateLaboratoryEnv(ctx context.Context, req *LaboratoryEnvReq) (*LaboratoryEnvResp, error)
	UpdateLaboratoryEnv(ctx context.Context, req *UpdateEnvReq) (*UpdateEnvResp, error)
	CreateReg(ctx context.Context, req *RegistryReq) error
}
