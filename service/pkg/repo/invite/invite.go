package invite

import "github.com/scienceol/studio/service/pkg/repo"

type inviteImpl struct {
	repo.IDOrUUIDTranslate
}

func New() repo.Invite {
	return &inviteImpl{
		IDOrUUIDTranslate: repo.NewBaseDB(),
	}
}
