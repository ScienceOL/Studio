package workflow

import (
	cwf "github.com/scienceol/studio/service/pkg/core/workflow"
)

// Doc-only DTOs for Swagger. Keep API response shapes close to handlers
// and avoid using Go generics in annotations.

// TemplateListPage 用于 Swagger 展示节点模板分页结果（避免使用泛型类型）
type TemplateListPage struct {
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Data     []cwf.TemplateListResp `json:"data"`
}

// WorkflowTemplateListPage 用于 Swagger 展示工作流模板分页结果
type WorkflowTemplateListPage struct {
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Data     []cwf.TemplateListRes `json:"data"`
}

// TaskPageMore 用于 Swagger 展示滚动分页（HasMore）结果
type TaskPageMore struct {
	HasMore  bool           `json:"has_more"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Data     []cwf.TaskResp `json:"data"`
}
