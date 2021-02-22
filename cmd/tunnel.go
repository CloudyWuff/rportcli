package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/cli"
	"github.com/cloudradar-monitoring/rportcli/internal/pkg/output"
	"github.com/cloudradar-monitoring/rportcli/internal/pkg/utils"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/controllers"
	"github.com/spf13/cobra"
)

var (
	tunnelCreateRequirementsP map[string]*string
)

func init() {
	tunnelsCmd.AddCommand(tunnelListCmd)
	tunnelsCmd.AddCommand(tunnelDeleteCmd)

	tunnelCreateRequirements := controllers.GetCreateTunnelRequirements()
	tunnelCreateRequirementsP = make(map[string]*string, len(tunnelCreateRequirements))
	for _, req := range tunnelCreateRequirements {
		paramVal := ""
		tunnelCreateCmd.Flags().StringVarP(&paramVal, req.Field, req.ShortName, req.Default, req.Description)
		tunnelCreateRequirementsP[req.Field] = &paramVal
	}
	tunnelsCmd.AddCommand(tunnelCreateCmd)

	rootCmd.AddCommand(tunnelsCmd)
}

var tunnelsCmd = &cobra.Command{
	Use:   "tunnel [command]",
	Short: "Tunnel API",
	Args:  cobra.ArbitraryArgs,
}

var tunnelListCmd = &cobra.Command{
	Use:   "list",
	Short: "Tunnel List API",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rportAPI, err := buildRport()
		if err != nil {
			return err
		}

		tr := &output.TunnelRenderer{
			ColCountCalculator: utils.CalcTerminalColumnsCount,
		}
		tunnelController := &controllers.TunnelController{
			Rport:          rportAPI,
			TunnelRenderer: tr,
			IPProvider:     utils.APIIPProvider{},
		}

		return tunnelController.Tunnels(context.Background(), os.Stdout)
	},
}

const minArgsCount = 2

var tunnelDeleteCmd = &cobra.Command{
	Use:   "delete <CLIENT_ID> <TUNNEL_ID>",
	Short: "Terminates the specified tunnel of the specified client",
	Args:  cobra.MinimumNArgs(minArgsCount),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < minArgsCount {
			return fmt.Errorf("either CLIENT_ID or TUNNEL_ID is not provided")
		}

		rportAPI, err := buildRport()
		if err != nil {
			return err
		}

		tr := &output.TunnelRenderer{
			ColCountCalculator: utils.CalcTerminalColumnsCount,
		}
		tunnelController := &controllers.TunnelController{
			Rport:          rportAPI,
			TunnelRenderer: tr,
			IPProvider:     utils.APIIPProvider{},
		}

		return tunnelController.Delete(context.Background(), os.Stdout, args[0], args[1])
	},
}

var tunnelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new tunnel",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tunnelCreateRequirements := make(map[string]string, len(tunnelCreateRequirementsP))
		for k, valP := range tunnelCreateRequirementsP {
			tunnelCreateRequirements[k] = *valP
		}
		params := cli.FromValues(tunnelCreateRequirements)

		err := cli.CheckRequirementsError(params, controllers.GetCreateTunnelRequirements())
		if err != nil {
			return err
		}

		rportAPI, err := buildRport()
		if err != nil {
			return err
		}

		tr := &output.TunnelRenderer{
			ColCountCalculator: utils.CalcTerminalColumnsCount,
		}
		tunnelController := &controllers.TunnelController{
			Rport:          rportAPI,
			TunnelRenderer: tr,
			IPProvider:     utils.APIIPProvider{},
		}

		return tunnelController.Create(context.Background(), os.Stdout, params)
	},
}
