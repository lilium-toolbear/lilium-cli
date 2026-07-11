package auth

import "os"

// LoginMode selects the OIDC login flow.
type LoginMode string

const (
	ModeLoopback LoginMode = "loopback"
	ModeDevice   LoginMode = "device"
)

// DetectLoginMode chooses loopback vs device.
// force non-empty wins; SSH_CONNECTION / SSH_TTY imply device; else loopback.
func DetectLoginMode(env map[string]string, force LoginMode) LoginMode {
	if force != "" {
		return force
	}
	get := func(k string) string {
		if env != nil {
			if v, ok := env[k]; ok {
				return v
			}
		}
		return os.Getenv(k)
	}
	if get("SSH_CONNECTION") != "" || get("SSH_TTY") != "" {
		return ModeDevice
	}
	return ModeLoopback
}
