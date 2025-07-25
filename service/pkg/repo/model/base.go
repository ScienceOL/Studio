package model

import (
	"time"

	"github.com/scienceol/studio/service/pkg/common"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      common.BinUUID `gorm:"type:bytea;default:decode(replace(gen_random_uuid()::text,'-',''),'hex');uniqueIndex;not null" json:"uuid"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (b *BaseModel) BeforeUpdate(_ *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

type BaseModelNoUUID struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (b *BaseModelNoUUID) BeforeUpdate(_ *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}
