package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/auth"
	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

// Client is an authenticated ToolBear HTTP client.
type Client struct {
	Cfg        config.Config
	HTTP       *http.Client
	tokenFn    func(context.Context, config.Config) (string, error)
	hostFrom   func() (string, error)
	retryOnce  bool
}

// New creates a Client using the credential store for auth.
func New(cfg config.Config) *Client {
	return &Client{
		Cfg:  cfg,
		HTTP: &http.Client{Timeout: 60 * time.Second},
		tokenFn: auth.EnsureAccessToken,
		hostFrom: func() (string, error) {
			if cfg.Host != "" {
				return cfg.Host, nil
			}
			creds, err := auth.Load()
			if err != nil {
				return "", err
			}
			return creds.Host, nil
		},
		retryOnce: true,
	}
}

// Do performs an authenticated request. On 401 it refreshes once and retries.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	host, err := c.hostFrom()
	if err != nil {
		return nil, err
	}
	host = strings.TrimRight(host, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := host + path

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	doOnce := func() (*http.Response, error) {
		token, err := c.tokenFn(ctx, c.Cfg)
		if err != nil {
			return nil, err
		}
		var rdr io.Reader
		if bodyBytes != nil {
			rdr = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, rdr)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return c.HTTP.Do(req)
	}

	resp, err := doOnce()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && c.retryOnce {
		resp.Body.Close()
		// Force refresh by clearing near-expiry: re-call EnsureAccessToken after bumping expiry past.
		creds, loadErr := auth.Load()
		if loadErr == nil {
			creds.ExpiresAt = time.Now().Add(-time.Minute)
			_ = auth.Save(creds)
		}
		return doOnce()
	}
	return resp, nil
}

// DoJSON posts/puts JSON and returns the response.
func (c *Client) DoJSON(ctx context.Context, method, path string, payload any) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}
	return c.Do(ctx, method, path, body)
}

// ReadBody reads and closes the response body.
func ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// PrintJSON pretty-prints JSON to w, or raw bytes if not JSON.
func PrintJSON(w io.Writer, data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		_, err = w.Write(data)
		if err == nil && len(data) > 0 && data[len(data)-1] != '\n' {
			_, err = io.WriteString(w, "\n")
		}
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// CheckStatus returns an error for non-2xx responses with body snippet.
func CheckStatus(resp *http.Response, body []byte) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 400))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
