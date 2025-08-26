package model

type TagType string

const (
	WorkflowTemplateTag TagType = "workflow_template_tag"
)

// 标签系统，记录各个表的标签
type Tags struct {
	BaseModel
	Type TagType `gorm:"type:varchar(80);not null;uniqueIndex:idx_tags_tn,priority:1" json:"type"`
	Name string  `gorm:"type:varchar(200);not null;uniqueIndex:idx_tags_tn,priority:2" json:"name"`
}

func (*Tags) TableName() string {
	return "tags"
}
