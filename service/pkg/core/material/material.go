package material

import "context"

type MaterialService interface {
	CreateMaterial(ctx context.Context, req []*MaterialNode) error
}
