package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

type deviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	Error                   string `json:"error"`
	ErrorDesc               string `json:"error_description"`
}

// LoginDevice runs RFC 8628 device code flow.
func LoginDevice(ctx context.Context, cfg config.Config, scopes []string) (*Credentials, error) {
	return loginDevice(ctx, cfg, scopes, http.DefaultClient, os.Stderr)
}

func loginDevice(
	ctx context.Context,
	cfg config.Config,
	scopes []string,
	httpClient *http.Client,
	out io.Writer,
) (*Credentials, error) {
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client id is required (set LILIUM_CLIENT_ID or embed DefaultClientID after seed)")
	}
	if len(scopes) == 0 {
		scopes = append([]string(nil), config.DefaultScopes...)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("scope", joinScopes(scopes))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Host+"/oauth/device/code", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var dc deviceCodeResponse
	if err := json.Unmarshal(body, &dc); err != nil {
		return nil, fmt.Errorf("decode device code response: %w", err)
	}
	if dc.Error != "" {
		return nil, &OAuthError{Code: dc.Error, Description: dc.ErrorDesc, Status: resp.StatusCode}
	}
	if resp.StatusCode >= 400 || dc.DeviceCode == "" {
		return nil, fmt.Errorf("device code HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	interval := dc.Interval
	if interval <= 0 {
		interval = 5
	}
	expiresAt := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)

	uri := dc.VerificationURIComplete
	if uri == "" {
		uri = dc.VerificationURI
	}
	fmt.Fprintf(out, "To continue, open %s\n", uri)
	fmt.Fprintf(out, "and enter code: %s\n", dc.UserCode)
	fmt.Fprintf(out, "Waiting for authorization…\n")

	for {
		if time.Now().After(expiresAt) {
			return nil, fmt.Errorf("device code expired; run login again")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(interval) * time.Second):
		}

		tokenForm := url.Values{}
		tokenForm.Set("grant_type", deviceCodeGrant)
		tokenForm.Set("device_code", dc.DeviceCode)
		tokenForm.Set("client_id", cfg.ClientID)
		creds, err := exchangeForm(ctx, httpClient, cfg.Host+"/oauth/token", tokenForm)
		if err == nil {
			if len(creds.Scopes) == 0 {
				creds.Scopes = scopes
			}
			return withHostClient(cfg, creds), nil
		}
		oe, ok := err.(*OAuthError)
		if !ok {
			return nil, err
		}
		switch oe.Code {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5
			continue
		case "expired_token":
			return nil, fmt.Errorf("device code expired; run login again")
		case "access_denied":
			return nil, fmt.Errorf("authorization denied")
		default:
			return nil, oe
		}
	}
}
