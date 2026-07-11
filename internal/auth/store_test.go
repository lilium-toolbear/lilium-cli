package auth

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("LILIUM_CONFIG_DIR", t.TempDir())
	c := &Credentials{
		AccessToken:  "a",
		RefreshToken: "r",
		ExpiresAt:    time.Now().UTC().Truncate(time.Second),
		Scopes:       []string{"stock:read"},
		Host:         "https://example",
		ClientID:     "cid",
	}
	if err := Save(c); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(Path())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%o want 0600", info.Mode().Perm())
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "a" || got.RefreshToken != "r" || got.Host != "https://example" {
		t.Fatalf("%+v", got)
	}
}

func TestClear(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LILIUM_CONFIG_DIR", dir)
	if err := Save(&Credentials{AccessToken: "a", RefreshToken: "r"}); err != nil {
		t.Fatal(err)
	}
	if err := Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "credentials.json")); !os.IsNotExist(err) {
		t.Fatalf("expected missing file, err=%v", err)
	}
	if _, err := Load(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("want ErrNotLoggedIn, got %v", err)
	}
}
