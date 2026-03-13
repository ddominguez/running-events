package strava

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRefreshToken_MissingEnvVars(t *testing.T) {
	os.Unsetenv("STRAVA_CLIENT_ID")
	os.Unsetenv("STRAVA_CLIENT_SECRET")
	os.Unsetenv("STRAVA_REFRESH_TOKEN")

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr string
	}{
		{
			name:    "missing client ID",
			envVars: map[string]string{"STRAVA_CLIENT_SECRET": "secret", "STRAVA_REFRESH_TOKEN": "refresh"},
			wantErr: "STRAVA_CLIENT_ID not set",
		},
		{
			name:    "missing client secret",
			envVars: map[string]string{"STRAVA_CLIENT_ID": "id", "STRAVA_REFRESH_TOKEN": "refresh"},
			wantErr: "STRAVA_CLIENT_SECRET not set",
		},
		{
			name:    "missing refresh token",
			envVars: map[string]string{"STRAVA_CLIENT_ID": "id", "STRAVA_CLIENT_SECRET": "secret"},
			wantErr: "STRAVA_REFRESH_TOKEN not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("STRAVA_CLIENT_ID")
			os.Unsetenv("STRAVA_CLIENT_SECRET")
			os.Unsetenv("STRAVA_REFRESH_TOKEN")

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				os.Unsetenv("STRAVA_CLIENT_ID")
				os.Unsetenv("STRAVA_CLIENT_SECRET")
				os.Unsetenv("STRAVA_REFRESH_TOKEN")
			}()

			_, _, _, err := RefreshToken()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRefreshToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/oauth/token" {
			t.Errorf("expected /oauth/token, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
			ExpiresAt:    1234567890,
		})
	}))
	defer server.Close()

	originalTokenURL := tokenURL
	tokenURL = server.URL + "/oauth/token"
	defer func() { tokenURL = originalTokenURL }()

	os.Setenv("STRAVA_CLIENT_ID", "test_client_id")
	os.Setenv("STRAVA_CLIENT_SECRET", "test_client_secret")
	os.Setenv("STRAVA_REFRESH_TOKEN", "test_refresh_token")
	defer os.Unsetenv("STRAVA_CLIENT_ID")
	defer os.Unsetenv("STRAVA_CLIENT_SECRET")
	defer os.Unsetenv("STRAVA_REFRESH_TOKEN")

	accessToken, refreshToken, expiresAt, err := RefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken != "new_access_token" {
		t.Errorf("expected access_token %q, got %q", "new_access_token", accessToken)
	}
	if refreshToken != "new_refresh_token" {
		t.Errorf("expected refresh_token %q, got %q", "new_refresh_token", refreshToken)
	}
	if expiresAt != 1234567890 {
		t.Errorf("expected expires_at %d, got %d", 1234567890, expiresAt)
	}
}
