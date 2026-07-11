package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

const refreshSkew = 60 * time.Second

// EnsureAccessToken returns a valid access token, refreshing if near expiry.
func EnsureAccessToken(ctx context.Context, cfg config.Config) (string, error) {
	creds, err := Load()
	if err != nil {
		return "", err
	}
	if time.Until(creds.ExpiresAt) > refreshSkew && creds.AccessToken != "" {
		return creds.AccessToken, nil
	}
	if creds.RefreshToken == "" {
		_ = Clear()
		return "", fmt.Errorf("%w: session expired; run lilium auth login", ErrNotLoggedIn)
	}
	host := firstNonEmpty(cfg.Host, creds.Host)
	clientID := firstNonEmpty(cfg.ClientID, creds.ClientID)
	if host == "" || clientID == "" {
		return "", fmt.Errorf("missing host/client id for refresh")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", creds.RefreshToken)
	form.Set("client_id", clientID)
	refreshed, err := exchangeForm(ctx, http.DefaultClient, host+"/oauth/token", form)
	if err != nil {
		var oe *OAuthError
		if errors.As(err, &oe) && (oe.Code == "invalid_grant" || oe.Code == "invalid_token") {
			_ = Clear()
			return "", fmt.Errorf("%w: refresh failed (%s); run lilium auth login", ErrNotLoggedIn, oe.Code)
		}
		return "", err
	}
	refreshed.Host = host
	refreshed.ClientID = clientID
	if len(refreshed.Scopes) == 0 {
		refreshed.Scopes = creds.Scopes
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = creds.RefreshToken
	}
	if err := Save(refreshed); err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
