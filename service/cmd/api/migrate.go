package api

import (
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/model/migrate"
	"github.com/spf13/cobra"
)

func NewMigrate() *cobra.Command {
	return &cobra.Command{
		Use:                "migrate",
		Long:               `api server db migrate`,
		SilenceUsage:       true,
		PersistentPreRunE:  initGlobalResource,
		PersistentPostRunE: cleanGlobalResource,
		PreRunE:            initMigrate,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return migrate.Table(cmd.Root().Context())
		},
		PostRunE: func(cmd *cobra.Command, _ []string) error {
			db.ClosePostgres(cmd.Context())
			return nil
		},
	}
}
