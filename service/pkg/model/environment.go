package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type EnvironmentStatus string

const (
	INIT    EnvironmentStatus = "INIT"
	DELETED EnvironmentStatus = "DELETED"
)

// 实验室环境表
type Laboratory struct {
	BaseModel
	Name            string            `gorm:"type:varchar(120);not null;index:idx_laboratory_user_status_name,priority:3" json:"name"`
	UserID          string            `gorm:"type:varchar(120);not null;index:idx_laboratory_user_status_name,priority:1" json:"user_id"`
	Status          EnvironmentStatus `gorm:"type:varchar(20);not null;index:idx_laboratory_user_status_name,priority:2" json:"status"`
	AccessKey       string            `gorm:"type:varchar(120);not null;uniqueIndex:idx_laboratory_lab_id_ak_sk,priority:1" json:"access_key"`
	AccessSecret    string            `gorm:"type:varchar(120);not null;uniqueIndex:idx_laboratory_lab_id_ak_sk,priority:2" json:"access_secret"`
	Description     *string           `gorm:"type:text" json:"description"`
	IsOnline        bool              `gorm:"type:boolean;not null;default:false;index:idx_laboratory_online" json:"is_online"`
	LastConnectedAt *time.Time        `gorm:"type:timestamp" json:"last_connected_at"`
}

func (*Laboratory) TableName() string {
	return "laboratory"
}

type LaboratoryMemberRole string

const (
	LaboratoryMemberAdmin  LaboratoryMemberRole = "admin"
	LaboratoryMemberNormal LaboratoryMemberRole = "normal"
)

type LaboratoryMember struct {
	BaseModel
	UserID string               `gorm:"type:varchar(120);not null;uniqueIndex:idx_laboratorymemeber_lu,priority:1" json:"user_id"`
	LabID  int64                `gorm:"type:bigint;not null;index:idx_laboratorymemeber_ld;uniqueIndex:idx_laboratorymemeber_lu,priority:2" json:"lab_id"`
	Role   LaboratoryMemberRole `gorm:"type:varchar(120);not null" json:"role"`
}

func (*LaboratoryMember) TableName() string {
	return "laboratory_member"
}

// BeforeSave GORM hook, to validate data before saving
func (m *LaboratoryMember) BeforeSave(tx *gorm.DB) (err error) {
	switch m.Role {
	case LaboratoryMemberAdmin, LaboratoryMemberNormal:
		return nil
	default:
		return errors.New("invalid laboratory member role")
	}
}

type InvitationType string

const (
	InvitationTypeLab InvitationType = "lab"
)

type LaboratoryInvitation struct {
	BaseModel
	ExpiresAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"expires_at"`
	Type      InvitationType `gorm:"type:varchar(50);not null;index:idx_labinv_tt,priority:1" json:"type"`
	ThirdID   string         `gorm:"type:varchar(50);not null;index:idx_labinv_tt,priority:2" json:"third_id"`
	UserID    string         `gorm:"type:varchar(120);not null" json:"user_id"`
}

func (*LaboratoryInvitation) TableName() string {
	return "laboratory_invitation"
}
