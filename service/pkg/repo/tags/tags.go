package tags

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/gorm/clause"
)

type tagImpl struct {
	repo.IDOrUUIDTranslate
}

func NewTag() repo.Tags {
	return &tagImpl{
		IDOrUUIDTranslate: repo.NewBaseDB(),
	}
}

func (t *tagImpl) UpsertTags(ctx context.Context, tags []*model.Tags) error {
	if len(tags) == 0 {
		return nil
	}

	statement := t.DBWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "type"},
			{Name: "name"},
		},
		DoNothing: true,
	}).Create(tags)

	if statement.Error != nil {
		logger.Errorf(ctx, "UpsertTags err: %+v", statement.Error)
		return code.CreateDataErr.WithMsg(statement.Error.Error())
	}

	return nil
}
func (t *tagImpl) GetAllTags(ctx context.Context, tagType model.TagType) ([]string, error) {
	tags := []string{}
	if err := t.DBWithContext(ctx).
		Model(&model.Tags{}).
		Select("name").
		Where("type = ?", tagType).
		Find(&tags).Error; err != nil {
		logger.Errorf(ctx, "GetAllTags fail tag: %+v", tagType)

		return nil, code.QueryRecordErr.WithMsgf("tag: %+v", tagType)
	}

	return tags, nil
}
