package cmd

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

func newMarketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market",
		Short: "Player market APIs",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "listings",
		Short: "GET /api/market/listings",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/market/listings") },
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "my-orders",
		Short: "GET /api/market/my-orders",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/market/my-orders") },
	})

	var (
		buyCategory, buyKey, buyQuality, buyPrice string
		buyQty                                    int
		buyPalMinLevel                            int
	)
	buy := &cobra.Command{
		Use:   "buy",
		Short: "POST /api/market/buy",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{
				"item_category": buyCategory,
				"item_key":      buyKey,
				"item_quality":  buyQuality,
				"price":         buyPrice,
				"quantity":      buyQty,
			}
			if buyPalMinLevel > 0 {
				body["pal_min_level"] = buyPalMinLevel
			}
			return runJSON(http.MethodPost, "/api/market/buy", body)
		},
	}
	buy.Flags().StringVar(&buyCategory, "category", "", "item category")
	buy.Flags().StringVar(&buyKey, "item-key", "", "item matching key")
	buy.Flags().StringVar(&buyQuality, "quality", "", "item quality")
	buy.Flags().StringVar(&buyPrice, "price", "", "max price per unit")
	buy.Flags().IntVar(&buyQty, "qty", 1, "quantity")
	buy.Flags().IntVar(&buyPalMinLevel, "pal-min-level", 0, "minimum pal level")
	_ = buy.MarkFlagRequired("category")
	_ = buy.MarkFlagRequired("item-key")
	_ = buy.MarkFlagRequired("price")
	cmd.AddCommand(buy)

	var (
		sellCategory, sellItemID, sellPrice string
		sellQty, sellPalID, sellEggID       int
	)
	sell := &cobra.Command{
		Use:   "sell",
		Short: "POST /api/market/sell",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{
				"item_category": sellCategory,
				"item_id":       sellItemID,
				"price":         sellPrice,
				"quantity":      sellQty,
			}
			if sellPalID > 0 {
				body["pal_id"] = sellPalID
			}
			if sellEggID > 0 {
				body["egg_id"] = sellEggID
			}
			return runJSON(http.MethodPost, "/api/market/sell", body)
		},
	}
	sell.Flags().StringVar(&sellCategory, "category", "", "item category")
	sell.Flags().StringVar(&sellItemID, "item-id", "", "item id")
	sell.Flags().StringVar(&sellPrice, "price", "", "price per unit")
	sell.Flags().IntVar(&sellQty, "qty", 1, "quantity")
	sell.Flags().IntVar(&sellPalID, "pal-id", 0, "pal id")
	sell.Flags().IntVar(&sellEggID, "egg-id", 0, "egg id")
	_ = sell.MarkFlagRequired("category")
	_ = sell.MarkFlagRequired("price")
	cmd.AddCommand(sell)

	cancel := &cobra.Command{
		Use:   "cancel <order_id>",
		Short: "POST /api/market/cancel/{order_id}",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("order_id must be an integer")
			}
			return runJSON(http.MethodPost, fmt.Sprintf("/api/market/cancel/%d", id), map[string]any{})
		},
	}
	cmd.AddCommand(cancel)
	return cmd
}
