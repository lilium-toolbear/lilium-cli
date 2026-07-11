package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

var (
	flagHost         string
	flagClientID     string
	flagCallbackPort int
	flagVerbose      bool
)

var rootCmd = &cobra.Command{
	Use:           "lilium",
	Short:         "Lilium CLI — call ToolBear APIs as yourself",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagHost, "host", "", "ToolBear origin (env LILIUM_HOST)")
	rootCmd.PersistentFlags().StringVar(&flagClientID, "client-id", "", "OIDC client id (env LILIUM_CLIENT_ID)")
	rootCmd.PersistentFlags().IntVar(&flagCallbackPort, "callback-port", 0, "loopback callback port")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "verbose HTTP output")

	rootCmd.AddCommand(newAuthCmd())
	rootCmd.AddCommand(newAPICmd())
	rootCmd.AddCommand(newWalletCmd())
	rootCmd.AddCommand(newStockCmd())
	rootCmd.AddCommand(newMarketCmd())
	rootCmd.AddCommand(newTurnipCmd())
}

func loadConfig() (config.Config, error) {
	return config.Load(flagHost, flagClientID, flagCallbackPort)
}
