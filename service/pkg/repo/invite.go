package repo

import (
	"context"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Invite interface {
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
	// 删除单条数据
	DelData(ctx context.Context, tableModel schema.Tabler, condition map[string]any) error
	// 创建单条
	CreateData(ctx context.Context, data schema.Tabler) error
}
