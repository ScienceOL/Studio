package material

import "context"

type Service interface {
	CreateMaterial(ctx context.Context, req []*Node) error
}
