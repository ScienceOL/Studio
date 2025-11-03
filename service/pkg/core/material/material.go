package material

import (
	"context"

	"github.com/olahol/melody"
)

type Service interface {
	CreateMaterial(ctx context.Context, req *GraphNodeReq) error
	SaveMaterial(ctx context.Context, req *SaveGrapReq) error
	LabMaterial(ctx context.Context, req *MaterialReq) ([]*MaterialResp, error)
	BatchUpdateMaterial(ctx context.Context, req *UpdateMaterialReq) error
	BatchUpdateUniqueName(ctx context.Context, req *UpdateMaterialReq) error
	CreateEdge(ctx context.Context, req *GraphEdge) error
	OnWSMsg(ctx context.Context, s *melody.Session, b []byte) error
	OnWSConnect(ctx context.Context, s *melody.Session) error
	OnMaterialNotify(ctx context.Context, msg string) error
	DownloadMaterial(ctx context.Context, req *DownloadMaterial) (*GraphNodeReq, error)
	GetMaterialTemplate(ctx context.Context, req *TemplateReq) (*TemplateResp, error)
	ResourceList(ctx context.Context, req *ResourceReq) (*ResourceResp, error)
	ResourceTemplateList(ctx context.Context, req *ResourceTemplateReq) (*ResourceTemplateResp, error)
	DeviceAction(ctx context.Context, req *DeviceActionReq) (*DeviceActionResp, error)

	// 开发机相关
	// StartMachine(ctx context.Context, req *StartMachineReq) (*StartMachineRes, error)
	// DelMachine(ctx context.Context, req *DelMachineReq) error
	// StopMachine(ctx context.Context, req *StopMachineReq) error
	// MachineStatus(ctx context.Context, req *MachineStatusReq) (*MachineStatusRes, error)

	// edge 相关接口
	EdgeCreateMaterial(ctx context.Context, req *CreateMaterialReq) ([]*CreateMaterialResp, error)
	EdgeUpsertMaterial(ctx context.Context, req *UpsertMaterialReq) ([]*UpsertMaterialResp, error)
	EdgeCreateEdge(ctx context.Context, req *CreateMaterialEdgeReq) error
	EdgeQueryMaterial(ctx context.Context, req *MaterialQueryReq) (*MaterialQueryResp, error)
	EdgeDownloadMaterial(ctx context.Context) (*DownloadMaterialResp, error)
}
