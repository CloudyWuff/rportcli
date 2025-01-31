package cmd

import (
	"fmt"

	options "github.com/breathbath/go_utils/v2/pkg/config"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/api"
	"github.com/cloudradar-monitoring/rportcli/internal/pkg/applog"
	"github.com/cloudradar-monitoring/rportcli/internal/pkg/config"
	"github.com/cloudradar-monitoring/rportcli/internal/pkg/output"
	"github.com/cloudradar-monitoring/rportcli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	Verbose      = false
	OutputFormat = output.FormatHuman
	Timeout      = ""
	IsJSONPretty = false
	rootCmd      = &cobra.Command{
		Use:           "rportcli",
		Short:         "Rport cli",
		Version:       version(),
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if OutputFormat != "" && OutputFormat != output.FormatHuman && OutputFormat != output.FormatYAML && OutputFormat != output.FormatJSON {
				return fmt.Errorf(
					"unknown format '%s', supported formats are %s, %s",
					OutputFormat,
					output.FormatJSON,
					output.FormatYAML,
				)
			}
			return nil
		},
	}
)

func getOutputFormat() string {
	if OutputFormat == output.FormatJSON {
		if IsJSONPretty {
			return output.FormatJSONPretty
		}
		return output.FormatJSON
	}

	return OutputFormat
}

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(
		&IsJSONPretty,
		"json-pretty",
		"j",
		false,
		"in combination with json format this flag will pretty print the json data",
	)
	rootCmd.PersistentFlags().StringVarP(
		&OutputFormat,
		"output",
		"o",
		output.FormatHuman,
		fmt.Sprintf("Output format: %s, %s or %s", output.FormatJSON, output.FormatYAML, output.FormatHuman),
	)
	rootCmd.PersistentFlags().StringVarP(
		&Timeout,
		"timeout",
		"t",
		"",
		"Timeout value as seconds, e.g. 10s, minutes e.g. 1m or hours e.g. 2h, if not provided no timeout will be set",
	)
}

func initLog() {
	applog.Init(Verbose)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

func buildRport(params *options.ParameterBag) *api.Rport {
	auth := &utils.FallbackAuth{
		PrimaryAuth: &utils.StorageBasicAuth{
			AuthProvider: func() (login, pass string, err error) {
				login = params.ReadString(config.Login, "")
				pass = params.ReadString(config.Password, "")
				return
			},
		},
		FallbackAuth: &utils.BearerAuth{
			TokenProvider: func() (string, error) {
				return params.ReadString(config.Token, ""), nil
			},
		},
	}
	serverURL := params.ReadString(config.ServerURL, config.DefaultServerURL)
	if serverURL == "" {
		serverURL = config.DefaultServerURL
	}
	rportAPI := api.New(serverURL, auth)

	return rportAPI
}
