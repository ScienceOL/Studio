package schedule

import "github.com/spf13/cobra"

func New() *cobra.Command {
	return &cobra.Command{
		Use:          "schedule",
		Long:         `api server workflow schedule`,
		SilenceUsage: true,
		// PersistentPreRunE:  initGlobalResource,
		// PersistentPostRunE: cleanGlobalResource,
		// PreRunE:            initMigrate,
		// RunE: func(cmd *cobra.Command, _ []string) error {
		// 	return migrate.Table(cmd.Root().Context())
		// },
		// PostRunE: func(cmd *cobra.Command, _ []string) error {
		// 	db.ClosePostgres(cmd.Context())
		// 	return nil
		// },
	}
}


