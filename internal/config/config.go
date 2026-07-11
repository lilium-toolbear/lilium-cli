package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DefaultHost is the production Lilium / ToolBear origin.
// Override with --host or LILIUM_HOST for staging.
const DefaultHost = "https://lilium.chat"

// DefaultClientID is the official public OIDC client UUID.
// Replace after seeding; override with LILIUM_CLIENT_ID for staging.
const DefaultClientID = ""

// DefaultCallbackPorts are tried in order for loopback PKCE.
var DefaultCallbackPorts = []int{3847, 3848}

// DefaultScopes are requested on login when --scopes is omitted.
var DefaultScopes = []string{
	"openid",
	"profile",
	"wallet:read",
	"stock:read",
	"stock:write",
	"market:read",
	"market:write",
	"turnip:read",
	"turnip:write",
}

// Config holds runtime CLI settings.
type Config struct {
	Host         string
	ClientID     string
	CallbackPort int // 0 = try DefaultCallbackPorts
}

// Load builds Config from flags/env with defaults.
func Load(host, clientID string, callbackPort int) (Config, error) {
	cfg := Config{
		Host:         firstNonEmpty(host, os.Getenv("LILIUM_HOST"), DefaultHost),
		ClientID:     firstNonEmpty(clientID, os.Getenv("LILIUM_CLIENT_ID"), DefaultClientID),
		CallbackPort: callbackPort,
	}
	cfg.Host = strings.TrimRight(strings.TrimSpace(cfg.Host), "/")
	if cfg.Host == "" {
		return Config{}, fmt.Errorf("host is required (--host or LILIUM_HOST)")
	}
	if !strings.HasPrefix(cfg.Host, "http://") && !strings.HasPrefix(cfg.Host, "https://") {
		return Config{}, fmt.Errorf("host must include scheme: %s", cfg.Host)
	}
	if p := os.Getenv("LILIUM_CALLBACK_PORT"); callbackPort == 0 && p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return Config{}, fmt.Errorf("invalid LILIUM_CALLBACK_PORT: %w", err)
		}
		cfg.CallbackPort = n
	}
	return cfg, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
