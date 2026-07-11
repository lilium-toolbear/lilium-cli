package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/lilium-toolbear/lilium-cli/internal/auth"
	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with ToolBear OIDC",
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.Clear()
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show login status",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := auth.Load()
			if err != nil {
				fmt.Println("not logged in")
				return nil
			}
			fmt.Printf("logged in\nhost: %s\nclient_id: %s\nexpires_at: %s\nscopes: %s\n",
				creds.Host, creds.ClientID, creds.ExpiresAt.Format(time.RFC3339), strings.Join(creds.Scopes, " "))
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "token",
		Short: "Print access token (for agents)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			tok, err := auth.EnsureAccessToken(context.Background(), cfg)
			if err != nil {
				return err
			}
			fmt.Println(tok)
			return nil
		},
	})
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var web, device bool
	var scopes []string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in via browser loopback or device code",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			if cfg.ClientID == "" {
				return fmt.Errorf("client id is required (set LILIUM_CLIENT_ID or embed DefaultClientID after seed)")
			}
			if web && device {
				return fmt.Errorf("use only one of --web or --device")
			}
			var force auth.LoginMode
			switch {
			case web:
				force = auth.ModeLoopback
			case device:
				force = auth.ModeDevice
			}
			mode := auth.DetectLoginMode(nil, force)
			if len(scopes) == 0 {
				scopes = append([]string(nil), config.DefaultScopes...)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()

			var creds *auth.Credentials
			switch mode {
			case auth.ModeDevice:
				creds, err = auth.LoginDevice(ctx, cfg, scopes)
			default:
				creds, err = auth.LoginLoopback(ctx, cfg, scopes)
			}
			if err != nil {
				return err
			}
			if err := auth.Save(creds); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Logged in as OIDC client %s (%s)\n", creds.ClientID, mode)
			return nil
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "force loopback browser login")
	cmd.Flags().BoolVar(&device, "device", false, "force device code login")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "OIDC scopes (space/comma separated via repeated flags)")
	return cmd
}
