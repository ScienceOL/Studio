package repo

import (
	"context"
	"reflect"

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
	DBWithContext(ctx context.Context) *gorm.DB
	// 开启事务
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
	// 计数
	Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error)
	// 批量 uuid 转 id
	UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids ...uuid.UUID) map[uuid.UUID]int64
	// 批量 id 转 uuid
	ID2UUID(ctx context.Context, tableModel schema.Tabler, ids ...int64) map[int64]uuid.UUID
	// 获取数据
	FindDatas(ctx context.Context, datas any, condition map[string]any, keys ...string) error
	// 更新数据
	UpdateData(ctx context.Context, data any, condition map[string]any, keys ...string) error
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

func (b *Base) UUID2ID(ctx context.Context, tableModel schema.Tabler, uuids ...uuid.UUID) map[uuid.UUID]int64 {
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

func (b *Base) ID2UUID(ctx context.Context, tableModel schema.Tabler, ids ...int64) map[int64]uuid.UUID {
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

func (b *Base) Count(ctx context.Context, tableModel schema.Tabler, condition map[string]any) (int64, error) {
	var count int64
	if err := b.DBWithContext(ctx).Model(tableModel).Where(condition).Count(&count).Error; err != nil {
		logger.Errorf(ctx, "Count fail condition : %+v, err: %+v", condition, err)
		return 0, code.QueryRecordErr
	}

	return count, nil
}

func (b *Base) FindDatas(ctx context.Context, datas any, condition map[string]any, keys ...string) error {
	if datas == nil {
		return code.NotSlicePointerErr.WithMsg("datas cannot be nil")
	}

	// 1. 判断 datas 是否是一个指向 slice 的指针
	datasVal := reflect.ValueOf(datas)
	if datasVal.Kind() != reflect.Ptr || datasVal.Elem().Kind() != reflect.Slice {
		return code.NotSlicePointerErr
	}

	// 2. 判断 slice 的元素类型是否实现了 schema.Tabler
	sliceType := datasVal.Type().Elem()
	elemType := sliceType.Elem()
	tablerType := reflect.TypeOf((*schema.Tabler)(nil)).Elem()

	// 检查元素类型或其指针类型是否实现了 Tabler 接口
	// GORM 通常在指针接收器上定义 TableName() 方法
	var tableModel schema.Tabler
	if elemType.Kind() == reflect.Ptr {
		// 元素是 *User
		if !elemType.Implements(tablerType) {
			return code.ModelNotImplementTablerErr.WithMsgf("model %s not implement schema.Tabler", elemType.String())
		}
		tableModel = reflect.New(elemType.Elem()).Interface().(schema.Tabler)
	} else {
		// 元素是 User
		if !reflect.PointerTo(elemType).Implements(tablerType) {
			return code.ModelNotImplementTablerErr.WithMsgf("model *%s not implement schema.Tabler", elemType.String())
		}
		tableModel = reflect.New(elemType).Interface().(schema.Tabler)
	}

	// 3. 执行查询
	db := b.DBWithContext(ctx).Model(tableModel)
	if len(keys) > 0 {
		db = db.Select(keys)
	}

	if err := db.Where(condition).Find(datas).Error; err != nil {
		logger.Errorf(ctx, "FindDatas fail table name: %s, condition: %+v, err: %+v", tableModel.TableName(), condition, err)
		return code.QueryRecordErr.WithErr(err)
	}

	return nil
}

func (b *Base) UpdateData(ctx context.Context, data any, condition map[string]any, keys ...string) error {
	dataType := reflect.TypeOf(data)
	dataKind := dataType.Kind()
	if dataKind != reflect.Ptr {
		return code.NotPointerErr
	}

	dataValue := reflect.ValueOf(data)
	if dataValue.IsNil() {
		return code.NotPointerErr
	}

	query := b.DBWithContext(ctx).Where(condition)
	if len(keys) > 0 {
		query = query.Select(keys)
	}
	if err := query.Updates(data).Error; err != nil {
		return code.UpdateDataErr.WithErr(err)
	}

	return nil
}
