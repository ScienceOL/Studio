package model

import (
	"time"

	"github.com/gofrs/uuid/v5"
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
