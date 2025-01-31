package cmd

import (
	"context"
	"os"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/config"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/output"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/controllers"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(meCmd)
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "show current user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		params := config.LoadParamsFromFileAndEnv(cmd.Flags())

		rportAPI := buildRport(params)
		rendr := &output.MeRenderer{
			Writer: os.Stdout,
			Format: getOutputFormat(),
		}

		meController := &controllers.MeController{
			Rport:      rportAPI,
			MeRenderer: rendr,
		}

		ctx, cancel := buildContext(context.Background())
		defer cancel()

		return meController.Me(ctx)
	},
}
