package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/lilium-toolbear/lilium-cli/internal/api"
	"github.com/lilium-toolbear/lilium-cli/internal/auth"
)

func newStockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stock",
		Short: "Stock market APIs",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "portfolio",
		Short: "GET /api/stock/portfolio",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/stock/portfolio") },
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "overview",
		Short: "GET /api/stock/overview",
		RunE:  func(cmd *cobra.Command, args []string) error { return runGET("/api/stock/overview") },
	})

	var symbol, shares, amount string
	buy := &cobra.Command{
		Use:   "buy",
		Short: "POST /api/stock/buy",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := stockTradeBody(symbol, shares, amount)
			if err != nil {
				return err
			}
			return runJSONWithIdempotency(http.MethodPost, "/api/stock/buy", body)
		},
	}
	buy.Flags().StringVar(&symbol, "symbol", "", "ticker symbol")
	buy.Flags().StringVar(&shares, "shares", "", "share quantity")
	buy.Flags().StringVar(&amount, "amount", "", "notional amount")
	_ = buy.MarkFlagRequired("symbol")
	cmd.AddCommand(buy)

	var sellSymbol, sellShares, sellAmount string
	sell := &cobra.Command{
		Use:   "sell",
		Short: "POST /api/stock/sell",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := stockTradeBody(sellSymbol, sellShares, sellAmount)
			if err != nil {
				return err
			}
			return runJSONWithIdempotency(http.MethodPost, "/api/stock/sell", body)
		},
	}
	sell.Flags().StringVar(&sellSymbol, "symbol", "", "ticker symbol")
	sell.Flags().StringVar(&sellShares, "shares", "", "share quantity")
	sell.Flags().StringVar(&sellAmount, "amount", "", "notional amount")
	_ = sell.MarkFlagRequired("symbol")
	cmd.AddCommand(sell)
	return cmd
}

func stockTradeBody(symbol, shares, amount string) (map[string]any, error) {
	if shares == "" && amount == "" {
		return nil, fmt.Errorf("provide --shares or --amount")
	}
	if shares != "" && amount != "" {
		return nil, fmt.Errorf("provide only one of --shares or --amount")
	}
	body := map[string]any{"symbol": symbol}
	if shares != "" {
		body["shares"] = shares
	}
	if amount != "" {
		body["amount"] = amount
	}
	return body, nil
}

func newIdempotencyKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func runJSONWithIdempotency(method, path string, payload any) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token, err := auth.EnsureAccessToken(context.Background(), cfg)
	if err != nil {
		return err
	}
	host := cfg.Host
	if host == "" {
		creds, err := auth.Load()
		if err != nil {
			return err
		}
		host = creds.Host
	}
	req, err := http.NewRequestWithContext(context.Background(), method, host+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Idempotency-Key", newIdempotencyKey())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	data, err := api.ReadBody(resp)
	if err != nil {
		return err
	}
	if flagVerbose {
		fmt.Fprintf(os.Stderr, "%s %s -> %s\n", method, path, resp.Status)
	}
	if err := api.CheckStatus(resp, data); err != nil {
		_, _ = os.Stdout.Write(data)
		return err
	}
	return api.PrintJSON(os.Stdout, data)
}
