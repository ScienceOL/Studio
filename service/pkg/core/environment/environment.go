package environment

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common"
)

type EnvService interface {
	CreateLaboratoryEnv(ctx context.Context, req *LaboratoryEnvReq) (*LaboratoryEnvResp, error)
	UpdateLaboratoryEnv(ctx context.Context, req *UpdateEnvReq) (*LaboratoryResp, error)
	DelLab(ctx context.Context, req *DelLabReq) error
	LabInfo(ctx context.Context, req *LabInfoReq) (*LabInfoResp, error)
	CreateResource(ctx context.Context, req *ResourceReq) error
	LabList(ctx context.Context, req *common.PageReq) (*common.PageMoreResp[[]*LaboratoryResp], error)
	LabMemberList(ctx context.Context, req *LabMemberReq) (*common.PageResp[[]*LabMemberResp], error)
	DelLabMember(ctx context.Context, req *DelLabMemberReq) error
	CreateInvite(ctx context.Context, req *InviteReq) (*InviteResp, error)
	AcceptInvite(ctx context.Context, req *AcceptInviteReq) error
	// UserInfo(ctx context.Context) (*model.UserData, error)
}
