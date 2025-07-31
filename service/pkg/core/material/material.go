package material

import "context"

type Service interface {
	CreateMaterial(ctx context.Context, req []*Node) error
	CreateEdge(ctx context.Context, req []*Edge) error
	HandleNotify(ctx context.Context, msg string) error
}
