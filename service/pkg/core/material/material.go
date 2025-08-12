package material

import (
	"context"

	"github.com/olahol/melody"
)

type Service interface {
	CreateMaterial(ctx context.Context, req *GraphNodeReq) error
	CreateEdge(ctx context.Context, req *GraphEdge) error
	HandleWSMsg(ctx context.Context, s *melody.Session, b []byte) error
	HandleWSConnect(ctx context.Context, s *melody.Session) error
	HandleNotify(ctx context.Context, msg string) error
}
