package cmd

import "github.com/spf13/cobra"

func newWalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Wallet read APIs",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "balance",
		Short: "GET /api/wallet/balance",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/wallet/balance") },
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "GET /api/wallet/stats",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/wallet/stats") },
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "transactions",
		Short: "GET /api/wallet/transactions",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/wallet/transactions") },
	})
	return cmd
}
