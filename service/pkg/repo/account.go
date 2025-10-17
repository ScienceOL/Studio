package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/model"
)

type Account interface {
	CreateLabUser(ctx context.Context, user *model.LabInfo) error
	DelLabUserInfo(ctx context.Context, req *model.LabAkSk) error
	BatchGetUserInfo(ctx context.Context, uesrIDs []string) ([]*model.UserData, error)
	GetUserInfo(ctx context.Context, userID string) (*model.UserData, error)
}

type LabAccount interface {
	GetLabUserInfo(ctx context.Context, req *model.LabAkSk) (*model.UserData, error)
}
