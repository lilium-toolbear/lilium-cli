package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials are persisted under ~/.config/lilium/credentials.json.
type Credentials struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes"`
	Host         string    `json:"host"`
	ClientID     string    `json:"client_id"`
}

// ErrNotLoggedIn is returned when credentials are missing.
var ErrNotLoggedIn = errors.New("not logged in")

// Path returns the credentials file path.
func Path() string {
	if dir := os.Getenv("LILIUM_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "credentials.json")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".config", "lilium", "credentials.json")
}

// Load reads credentials from disk.
func Load() (*Credentials, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotLoggedIn
		}
		return nil, err
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	if c.AccessToken == "" && c.RefreshToken == "" {
		return nil, ErrNotLoggedIn
	}
	return &c, nil
}

// Save writes credentials with directory mode 0700 and file mode 0600.
func Save(c *Credentials) error {
	if c == nil {
		return fmt.Errorf("credentials is nil")
	}
	path := Path()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Clear removes the credentials file if present.
func Clear() error {
	err := os.Remove(Path())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
