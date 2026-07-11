package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

const deviceCodeGrant = "urn:ietf:params:oauth:grant-type:device_code"

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// OAuthError is an OAuth error response.
type OAuthError struct {
	Code        string
	Description string
	Status      int
}

func (e *OAuthError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Description)
	}
	return e.Code
}

func exchangeForm(ctx context.Context, httpClient *http.Client, tokenURL string, form url.Values) (*Credentials, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("decode token response: %w (body=%s)", err, truncate(string(body), 200))
	}
	if tr.Error != "" {
		return nil, &OAuthError{Code: tr.Error, Description: tr.ErrorDesc, Status: resp.StatusCode}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token endpoint HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("token response missing access_token")
	}
	expires := time.Now().UTC().Add(time.Duration(tr.ExpiresIn) * time.Second)
	if tr.ExpiresIn <= 0 {
		expires = time.Now().UTC().Add(time.Hour)
	}
	return &Credentials{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    expires,
		Scopes:       strings.Fields(tr.Scope),
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func joinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}

func withHostClient(cfg config.Config, creds *Credentials) *Credentials {
	creds.Host = cfg.Host
	creds.ClientID = cfg.ClientID
	return creds
}
