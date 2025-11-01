// @title           ScienceOL Studio API
// @version         1.0
// @description     Studio 实验室管理系统 API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  harveyque@outlook.com

// @license.name  GNU Affero General Public License v3.0
// @license.url   http://www.gnu.org/licenses/agpl-3.0.en.html

// @host      localhost:48197
// @BasePath  /api
// @schemes   http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"os"

	"github.com/scienceol/studio/service/cmd/api"
	"github.com/scienceol/studio/service/cmd/schedule"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/spf13/cobra"
)

func main() {
	rootCtx := utils.SetupSignalContext()
	root := &cobra.Command{
		SilenceUsage: true,
		Short:        "ScienceOL",
		Long:         "ScienceOL Studio 后端服务",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}
	root.SetContext(rootCtx)
	root.AddCommand(api.NewWeb())
	root.AddCommand(api.NewMigrate())
	root.AddCommand(schedule.New())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
