package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/model"
)

type Tags interface {
	UpsertTags(ctx context.Context, tags []*model.Tags) error
	GetAllTags(ctx context.Context, tagType model.TagType) ([]string, error)
}
