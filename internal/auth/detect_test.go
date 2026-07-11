package auth

import "testing"

func TestDetectLoginMode(t *testing.T) {
	cases := []struct {
		name  string
		env   map[string]string
		force LoginMode
		want  LoginMode
	}{
		{"desktop", map[string]string{}, "", ModeLoopback},
		{"ssh", map[string]string{"SSH_CONNECTION": "1"}, "", ModeDevice},
		{"ssh tty", map[string]string{"SSH_TTY": "/dev/pts/0"}, "", ModeDevice},
		{"force web on ssh", map[string]string{"SSH_CONNECTION": "1"}, ModeLoopback, ModeLoopback},
		{"force device", map[string]string{}, ModeDevice, ModeDevice},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectLoginMode(tc.env, tc.force)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
