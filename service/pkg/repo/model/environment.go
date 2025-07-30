package model

type EnvironmentStatus string

const (
	INIT    EnvironmentStatus = "INIT"
	DELETED EnvironmentStatus = "DELETED"
)

// 实验室环境表
type Laboratory struct {
	BaseModel
	Name        string            `gorm:"type:varchar(120);not null;index:idx_laboratory_user_status_name,priority:3" json:"name"`
	UserID      string            `gorm:"type:varchar(120);not null;index:idx_laboratory_l_user_status_name,priority:1" json:"user_id"`
	Status      EnvironmentStatus `gorm:"type:varchar(20);not null;index:idx_laboratory_user_status_name,priority:2" json:"status"`
	Description *string           `gorm:"type:text" json:"description"`
}

// TableName specifies the table name for GORM
func (*Laboratory) TableName() string {
	return "laboratory"
}
