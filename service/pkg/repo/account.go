package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/repo/model"
)

type Account interface {
	CreateLabUser(ctx context.Context, user *model.LabInfo) error
	GetLabUserInfo(ctx context.Context, req *model.LabAkSk) (*model.UserData, error)
	DelLabUserInfo(ctx context.Context, req *model.LabAkSk) error
}
