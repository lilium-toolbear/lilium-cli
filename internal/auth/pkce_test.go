package auth

import (
	"strings"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	v, c, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	if len(v) < 43 {
		t.Fatalf("verifier too short: %d", len(v))
	}
	if len(c) != 43 {
		t.Fatalf("challenge length=%d want 43", len(c))
	}
	if strings.ContainsAny(c, "+/=") {
		t.Fatalf("challenge not url-safe: %s", c)
	}
}
