package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GeneratePKCE returns code_verifier and S256 code_challenge.
func GeneratePKCE() (verifier, challenge string, err error) {
	verifier, err = randomURLSafe(32)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// GenerateState returns a CSRF state value.
func GenerateState() (string, error) {
	s, err := randomURLSafe(16)
	if err != nil {
		return "", fmt.Errorf("state: %w", err)
	}
	return s, nil
}
