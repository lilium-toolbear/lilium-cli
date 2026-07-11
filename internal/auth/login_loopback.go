package auth

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/lilium-toolbear/lilium-cli/internal/config"
)

// LoginLoopback runs authorization-code + PKCE against a local callback.
func LoginLoopback(ctx context.Context, cfg config.Config, scopes []string) (*Credentials, error) {
	return loginLoopback(ctx, cfg, scopes, http.DefaultClient, openBrowser)
}

func loginLoopback(
	ctx context.Context,
	cfg config.Config,
	scopes []string,
	httpClient *http.Client,
	opener func(string) error,
) (*Credentials, error) {
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client id is required (set LILIUM_CLIENT_ID or embed DefaultClientID after seed)")
	}
	if len(scopes) == 0 {
		scopes = append([]string(nil), config.DefaultScopes...)
	}
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		return nil, err
	}
	state, err := GenerateState()
	if err != nil {
		return nil, err
	}

	ports := config.DefaultCallbackPorts
	if cfg.CallbackPort > 0 {
		ports = []int{cfg.CallbackPort}
	}

	var ln net.Listener
	var port int
	for _, p := range ports {
		l, listenErr := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if listenErr != nil {
			continue
		}
		ln = l
		port = p
		break
	}
	if ln == nil {
		return nil, fmt.Errorf("could not bind loopback callback on ports %v", ports)
	}
	defer ln.Close()

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("error") != "" {
			msg := q.Get("error_description")
			if msg == "" {
				msg = q.Get("error")
			}
			http.Error(w, msg, http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("authorization error: %s", msg):
			default:
			}
			return
		}
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("state mismatch"):
			default:
			}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("missing authorization code"):
			default:
			}
			return
		}
		_, _ = io.WriteString(w, "Login successful. You can close this window.")
		select {
		case codeCh <- code:
		default:
		}
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	authURL, err := url.Parse(cfg.Host + "/oauth/authorize")
	if err != nil {
		return nil, err
	}
	q := authURL.Query()
	q.Set("response_type", "code")
	q.Set("client_id", cfg.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", joinScopes(scopes))
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	authURL.RawQuery = q.Encode()

	fmt.Fprintf(os.Stderr, "Opening browser for login…\nIf it does not open, visit:\n%s\n", authURL.String())
	if opener != nil {
		_ = opener(authURL.String())
	}

	var code string
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case code = <-codeCh:
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", cfg.ClientID)
	form.Set("code_verifier", verifier)

	creds, err := exchangeForm(ctx, httpClient, cfg.Host+"/oauth/token", form)
	if err != nil {
		return nil, err
	}
	if len(creds.Scopes) == 0 {
		creds.Scopes = scopes
	}
	return withHostClient(cfg, creds), nil
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	return cmd.Start()
}
