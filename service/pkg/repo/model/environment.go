package model

type EnvironmentStatus string

const (
	INIT    EnvironmentStatus = "INIT"
	DELETED EnvironmentStatus = "DELETED"
)

// 实验室环境表
type Laboratory struct {
	BaseModel
	Name         string            `gorm:"type:varchar(120);not null;index:idx_laboratory_user_status_name,priority:3" json:"name"`
	UserID       string            `gorm:"type:varchar(120);not null;index:idx_laboratory_user_status_name,priority:1" json:"user_id"`
	Status       EnvironmentStatus `gorm:"type:varchar(20);not null;index:idx_laboratory_user_status_name,priority:2" json:"status"`
	AccessKey    string            `gorm:"type:varchar(120);not null;uniqueIndex:idx_laboratory_lab_id_ak_sk,priority:1" json:"access_key"`
	AccessSecret string            `gorm:"type:varchar(120);not null;uniqueIndex:idx_laboratory_lab_id_ak_sk,priority:2" json:"access_secret"`
	Description  *string           `gorm:"type:text" json:"description"`
}

// TableName specifies the table name for GORM
func (*Laboratory) TableName() string {
	return "laboratory"
}
