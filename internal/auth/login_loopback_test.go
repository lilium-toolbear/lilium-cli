package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

func TestExchangeFormSuccess(t *testing.T) {
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at",
			"refresh_token": "rt",
			"expires_in":    3600,
			"scope":         "openid stock:read",
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "abc")
	form.Set("redirect_uri", "http://127.0.0.1:3847/callback")
	form.Set("client_id", "cid")
	form.Set("code_verifier", "verifier")
	creds, err := exchangeForm(context.Background(), srv.Client(), srv.URL, form)
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessToken != "at" || gotForm.Get("code_verifier") != "verifier" {
		t.Fatalf("creds=%+v form=%v", creds, gotForm)
	}
}

func TestExchangeFormOAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"invalid_grant","error_description":"bad"}`)
	}))
	defer srv.Close()
	_, err := exchangeForm(context.Background(), srv.Client(), srv.URL, url.Values{"grant_type": {"authorization_code"}})
	oe, ok := err.(*OAuthError)
	if !ok || oe.Code != "invalid_grant" || !strings.Contains(oe.Error(), "bad") {
		t.Fatalf("got %v", err)
	}
}

func TestRefreshSuccessAndInvalidClears(t *testing.T) {
	t.Setenv("LILIUM_CONFIG_DIR", t.TempDir())
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = r.ParseForm()
		if r.PostForm.Get("refresh_token") == "bad" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-at",
			"refresh_token": "new-rt",
			"expires_in":    3600,
			"scope":         "openid",
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	cfg := config.Config{Host: srv.URL, ClientID: "cid"}
	if err := Save(&Credentials{
		AccessToken:  "old",
		RefreshToken: "good",
		ExpiresAt:    time.Now().Add(-time.Minute),
		Host:         srv.URL,
		ClientID:     "cid",
	}); err != nil {
		t.Fatal(err)
	}
	tok, err := EnsureAccessToken(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if tok != "new-at" {
		t.Fatalf("token=%s", tok)
	}
	got, _ := Load()
	if got.RefreshToken != "new-rt" {
		t.Fatalf("%+v", got)
	}

	if err := Save(&Credentials{
		AccessToken:  "old",
		RefreshToken: "bad",
		ExpiresAt:    time.Now().Add(-time.Minute),
		Host:         srv.URL,
		ClientID:     "cid",
	}); err != nil {
		t.Fatal(err)
	}
	_, err = EnsureAccessToken(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	_, loadErr := Load()
	if !errors.Is(loadErr, ErrNotLoggedIn) {
		t.Fatalf("expected cleared store, got %v", loadErr)
	}
	_ = fmt.Sprintf("%d", calls)
}
