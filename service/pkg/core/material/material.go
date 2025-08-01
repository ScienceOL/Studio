package material

import (
	"context"

	"github.com/olahol/melody"
)

type Service interface {
	CreateMaterial(ctx context.Context, req []*Node) error
	CreateEdge(ctx context.Context, req []*Edge) error
	HandleWSMsg(ctx context.Context, s *melody.Session, b []byte) error
	HandleNotify(ctx context.Context, msg string)error
}
