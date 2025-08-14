package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type IDOrUUIDTranslate interface {
	// TranslateIDOrUUID(ctx context.Context, data any) error
	DBWithContext(ctx context.Context) *gorm.DB
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
	Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error)
	UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids []uuid.UUID) map[uuid.UUID]int64
	ID2UUID(ctx context.Context, tableModel schema.Tabler, ids []int64) map[int64]uuid.UUID
}

type Base struct {
	*db.Datastore
}

func NewBaseDB() IDOrUUIDTranslate {
	return &Base{
		Datastore: db.DB(),
	}
}

func (b *Base) DBWithContext(ctx context.Context) *gorm.DB {
	return b.Datastore.DBWithContext(ctx)
}

func (b *Base) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return b.Datastore.ExecTx(ctx, fn)
}

func (b *Base) UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids []uuid.UUID) map[uuid.UUID]int64 {
	if len(uuids) == 0 {
		return map[uuid.UUID]int64{}
	}

	datas := make([]*model.BaseModel, 0, len(uuids))
	if err := b.DBWithContext(ctx).Model(tableModel).
		Select("id, uuid").Where("uuid in ?", uuids).
		Find(&datas).Error; err != nil {
		logger.Errorf(ctx, "TranslateUUID2ID fail uuids: %+v, err: %+v", uuids, err)
		return map[uuid.UUID]int64{}
	}

	return utils.SliceToMap(datas, func(item *model.BaseModel) (uuid.UUID, int64) {
		return item.UUID, item.ID
	})
}

func (b *Base) ID2UUID(ctx context.Context, tableModel schema.Tabler, ids []int64) map[int64]uuid.UUID {
	if len(ids) == 0 {
		return map[int64]uuid.UUID{}
	}

	datas := make([]*model.BaseModel, 0, len(ids))
	if err := b.DBWithContext(ctx).Model(tableModel).
		Select("id, uuid").Where("id in ?", ids).
		Find(&datas).Error; err != nil {
		logger.Errorf(ctx, "TranslateID2UUID fail ids: %+v, err: %+v", ids, err)
		return map[int64]uuid.UUID{}
	}

	return utils.SliceToMap(datas, func(item *model.BaseModel) (int64, uuid.UUID) {
		return item.ID, item.UUID
	})
}

// func (b *Base) TranslateIDOrUUID(ctx context.Context, data any) error {
// 	// 使用反射获取数据的类型和值
// 	v := reflect.ValueOf(data)
// 	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
// 		return code.NotPointerErr
// 	}
//
// 	id := int64(0)
// 	uuid := uuid.UUID{}
// 	switch m := data.(type) {
// 	case model.BaseDBModel:
// 		if m.GetID() == 0 || m.GetUUID().IsNil() {
// 			return code.ParamErr
// 		}
// 		id = m.GetID()
// 		uuid = m.GetUUID()
// 	default:
// 		return code.NotBaseDBTypeErr
// 	}
//
// 	// 判断转换方向
// 	if id <= 0 && !uuid.IsNil() {
// 		// UUID 转 ID
// 		err := b.DBWithContext(ctx).Model(data).
// 			Select("id").
// 			Where("uuid = ?", uuid).
// 			First(data).Error
// 		if err != nil {
// 			return code.QueryRecordErr
// 		}
// 	} else if id > 0 && uuid.IsNil() {
// 		err := b.DBWithContext(ctx).Model(data).
// 			Select("uuid").
// 			Where("id = ?", id).
// 			First(data).Error
// 		if err != nil {
// 			return code.QueryRecordErr
// 		}
// 	} else {
// 		return nil
// 	}
//
// 	return nil
// }

func (b *Base) Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error) {
	var count int64
	if err := b.DBWithContext(ctx).Model(tableModel).Where(condition).Count(&count).Error; err != nil {
		logger.Errorf(ctx, "Count fail condition : %+v, err: %+v", condition, err)
		return 0, code.QueryRecordErr
	}

	return count, nil
}
