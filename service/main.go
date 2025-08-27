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
		Short:        "unilab",
		Long:         "unilab 智能实验室后端服务",
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
