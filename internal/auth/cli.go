package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultAuthURL = "https://live.gptcode.cloud"
)

type Credentials struct {
	AccessToken  string `yaml:"access_token"`
	RefreshToken string `yaml:"refresh_token"`
	Email        string `yaml:"email"`
	UserID       string `yaml:"user_id"`
	ExpiresAt    int64  `yaml:"expires_at"`
}

func Login() error {
	// 1. Start local server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start local server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// 2. Open browser
	baseURL := os.Getenv("GPTCODE_AUTH_URL")
	if baseURL == "" {
		baseURL = DefaultAuthURL
	}
	loginURL := fmt.Sprintf("%s/auth/login?redirect_to=%s", baseURL, callbackURL)

	fmt.Printf("Opening browser to log in...\n")
	fmt.Printf("If browser doesn't open, visit:\n%s\n", loginURL)

	if err := openBrowser(loginURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}

	// 3. Wait for callback
	codeCh := make(chan *Credentials, 1)
	errCh := make(chan error, 1)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			// Extract params
			query := r.URL.Query()
			accessToken := query.Get("access_token")
			refreshToken := query.Get("refresh_token")
			email := query.Get("email")

			if accessToken == "" {
				http.Error(w, "Missing access_token", http.StatusBadRequest)
				errCh <- fmt.Errorf("callback missing access_token")
				return
			}

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<h1>Login Successful!</h1><p>You can close this tab and return to the terminal.</p><script>window.close()</script>`))

			codeCh <- &Credentials{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				Email:        email,
				ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(), // Approx expiry
			}
		}),
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for cleanup
	defer server.Shutdown(context.Background())

	select {
	case creds := <-codeCh:
		// 4. Save credentials
		if err := SaveCredentials(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}
		fmt.Printf("\nSuccessfully logged in as %s\n", creds.Email)
		return nil
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("login timeout")
	}
}

func SaveCredentials(creds *Credentials) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	credsDir := filepath.Join(home, ".gptcode")
	if err := os.MkdirAll(credsDir, 0o755); err != nil {
		return err
	}

	credsPath := filepath.Join(credsDir, "credentials.yaml")
	data, err := yaml.Marshal(creds)
	if err != nil {
		return err
	}

	return os.WriteFile(credsPath, data, 0o600)
}

func LoadCredentials() (*Credentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	credsPath := filepath.Join(home, ".gptcode", "credentials.yaml")
	data, err := os.ReadFile(credsPath)
	if err != nil {
		return nil, err
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
