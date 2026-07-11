package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

func TestLoginDevicePollsThenSucceeds(t *testing.T) {
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth/device/code":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":               "dc",
				"user_code":                 "ABCD-EFGH",
				"verification_uri":          "https://example/device",
				"verification_uri_complete": "https://example/device?user_code=ABCD-EFGH",
				"expires_in":                600,
				"interval":                  1,
			})
		case r.URL.Path == "/oauth/token":
			polls++
			if polls < 2 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"authorization_pending"}`)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "at",
				"refresh_token": "rt",
				"expires_in":    3600,
				"scope":         "openid stock:read",
				"token_type":    "Bearer",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := config.Config{Host: srv.URL, ClientID: "cid"}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	creds, err := loginDevice(ctx, cfg, []string{"openid", "stock:read"}, srv.Client(), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessToken != "at" || creds.RefreshToken != "rt" {
		t.Fatalf("%+v", creds)
	}
	if polls < 2 {
		t.Fatalf("expected at least 2 polls, got %d", polls)
	}
}

func TestLoginDeviceSlowDown(t *testing.T) {
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/device/code":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":       "dc",
				"user_code":         "CODE",
				"verification_uri":  "https://example/device",
				"expires_in":        600,
				"interval":          1,
			})
		case "/oauth/token":
			polls++
			if polls == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"slow_down"}`)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "at",
				"refresh_token": "rt",
				"expires_in":    60,
				"token_type":    "Bearer",
			})
		}
	}))
	defer srv.Close()
	cfg := config.Config{Host: srv.URL, ClientID: "cid"}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	creds, err := loginDevice(ctx, cfg, nil, srv.Client(), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(creds.AccessToken, "at") {
		t.Fatalf("%+v", creds)
	}
}
