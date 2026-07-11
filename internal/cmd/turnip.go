package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
)

func newTurnipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "turnip",
		Short: "Turnip market / farm APIs",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "inventory",
		Short: "GET /api/turnip/inventory",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/turnip/inventory") },
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "farm",
		Short: "GET /api/turnip/farm",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/turnip/farm") },
	})

	var buyQty, buyAmount, buyMaxPrice string
	buy := &cobra.Command{
		Use:   "buy",
		Short: "POST /api/turnip/buy",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if buyQty != "" {
				body["quantity"] = buyQty
			}
			if buyAmount != "" {
				body["amount"] = buyAmount
			}
			if buyMaxPrice != "" {
				body["max_price"] = buyMaxPrice
			}
			return runJSON(http.MethodPost, "/api/turnip/buy", body)
		},
	}
	buy.Flags().StringVar(&buyQty, "qty", "", "quantity")
	buy.Flags().StringVar(&buyAmount, "amount", "", "spend amount")
	buy.Flags().StringVar(&buyMaxPrice, "max-price", "", "max unit price")
	cmd.AddCommand(buy)

	var sellQty, sellMinPrice string
	var sellBatchID int
	sell := &cobra.Command{
		Use:   "sell",
		Short: "POST /api/turnip/sell",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if sellQty != "" {
				body["quantity"] = sellQty
			}
			if sellBatchID > 0 {
				body["batch_id"] = sellBatchID
			}
			if sellMinPrice != "" {
				body["min_price"] = sellMinPrice
			}
			return runJSON(http.MethodPost, "/api/turnip/sell", body)
		},
	}
	sell.Flags().StringVar(&sellQty, "qty", "", "quantity")
	sell.Flags().IntVar(&sellBatchID, "batch-id", 0, "batch id")
	sell.Flags().StringVar(&sellMinPrice, "min-price", "", "min unit price")
	cmd.AddCommand(sell)
	return cmd
}
