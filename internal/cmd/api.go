package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lilium-toolbear/lilium-cli/internal/api"
)

func newAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api <METHOD> <PATH>",
		Short: "Call a ToolBear API path with auth",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path := args[1]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			var body io.Reader
			if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				if len(bytes.TrimSpace(data)) > 0 {
					body = bytes.NewReader(data)
				}
			}
			client := api.New(cfg)
			resp, err := client.Do(context.Background(), method, path, body)
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
				if len(data) > 0 && data[len(data)-1] != '\n' {
					fmt.Println()
				}
				return err
			}
			return api.PrintJSON(os.Stdout, data)
		},
	}
	return cmd
}

func runGET(path string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client := api.New(cfg)
	resp, err := client.Do(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	data, err := api.ReadBody(resp)
	if err != nil {
		return err
	}
	if flagVerbose {
		fmt.Fprintf(os.Stderr, "GET %s -> %s\n", path, resp.Status)
	}
	if err := api.CheckStatus(resp, data); err != nil {
		_, _ = os.Stdout.Write(data)
		return err
	}
	return api.PrintJSON(os.Stdout, data)
}

func runJSON(method, path string, payload any) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client := api.New(cfg)
	resp, err := client.DoJSON(context.Background(), method, path, payload)
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
