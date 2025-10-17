package model

import (
	"time"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null" json:"uuid"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (b *BaseModel) BeforeCreate(*gorm.DB) error {
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	return nil
}

func (b *BaseModel) BeforeUpdate(_ *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

type BaseDBModel interface {
	GetID() int64
	GetUUID() uuid.UUID
}

func (b BaseModel) GetID() int64 {
	return b.ID
}

func (b BaseModel) GetUUID() uuid.UUID {
	return b.UUID
}
