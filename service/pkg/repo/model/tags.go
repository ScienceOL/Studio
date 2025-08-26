package model

type TagType string

const (
	WorkflowTagType         = "workflow_tag_type"
	WrokflowTemplateTagType = "workflow_template_tag_type"
)

type Tags struct {
	BaseModel
	Type TagType `gorm:"varchar(80)"`
}
