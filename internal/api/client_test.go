package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/auth"
	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

func TestClientAttachesBearerAndRetriesOn401(t *testing.T) {
	t.Setenv("LILIUM_CONFIG_DIR", t.TempDir())
	calls := 0
	tokens := []string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "refreshed",
				"refresh_token": "rt2",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})
			return
		}
		calls++
		tokens = append(tokens, r.Header.Get("Authorization"))
		if calls == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":"expired"}`)
			return
		}
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	if err := auth.Save(&auth.Credentials{
		AccessToken:  "stale",
		RefreshToken: "rt",
		ExpiresAt:    time.Now().Add(time.Hour),
		Host:         srv.URL,
		ClientID:     "cid",
	}); err != nil {
		t.Fatal(err)
	}

	c := New(config.Config{Host: srv.URL, ClientID: "cid"})
	c.HTTP = srv.Client()
	resp, err := c.Do(context.Background(), http.MethodGet, "/api/stock/portfolio", nil)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := ReadBody(resp)
	if resp.StatusCode != 200 || !strings.Contains(string(body), "ok") {
		t.Fatalf("status=%d body=%s", resp.StatusCode, body)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
	if tokens[0] != "Bearer stale" || tokens[1] != "Bearer refreshed" {
		t.Fatalf("tokens=%v", tokens)
	}
}
