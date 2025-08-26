package migrate

import (
	"context"

	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"github.com/scienceol/studio/service/pkg/utils"
)

func Table(_ context.Context) error {
	return utils.IfErrReturn(func() error {
		return db.DB().DBIns().AutoMigrate(
			&model.Laboratory{},             // 实验室
			&model.ResourceNodeTemplate{},   // 资源模板
			&model.ResourceHandleTemplate{}, // 资源 handle 模板
			&model.WorkflowNodeTemplate{},   // 实验室动作
			&model.MaterialNode{},           // 物料节点
			&model.MaterialEdge{},           // 物料边
			&model.Workflow{},
			&model.WorkflowNode{},
			&model.WorkflowEdge{},
			&model.WorkflowConsole{},
			&model.WorkflowHandleTemplate{},
			&model.WorkflowNodeJob{},
			&model.WorkflowTask{},
			&model.Tags{},
			&model.LaboratoryMember{},
			&model.LaboratoryInvitation{},
		) // 动作节点handle 模板
	})
}
