package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUpdateActionsVariable_MissingEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr string
	}{
		{
			name:    "missing GITHUB_REPOSITORY",
			envVars: map[string]string{"GH_PAT": "test_pat"},
			wantErr: "GITHUB_REPOSITORY not set",
		},
		{
			name:    "missing GH_PAT",
			envVars: map[string]string{"GITHUB_REPOSITORY": "owner/repo"},
			wantErr: "GH_PAT not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("GITHUB_REPOSITORY")
			os.Unsetenv("GH_PAT")
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer os.Unsetenv("GITHUB_REPOSITORY")
			defer os.Unsetenv("GH_PAT")

			err := UpdateActionsVariable("TEST_VAR", "test_value")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestUpdateActionsVariable_InvalidRepoFormat(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "invalid")
	os.Setenv("GH_PAT", "test_pat")
	defer os.Unsetenv("GITHUB_REPOSITORY")
	defer os.Unsetenv("GH_PAT")

	err := UpdateActionsVariable("TEST_VAR", "test_value")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid GITHUB_REPOSITORY format") {
		t.Errorf("expected invalid format error, got %q", err.Error())
	}
}

func TestUpdateActionsVariable_Success(t *testing.T) {
	var receivedReq http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/repos/owner/repo/actions/variables/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test_pat" {
			t.Errorf("expected Authorization header, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		json.NewDecoder(r.Body).Decode(&receivedReq)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	originalAPIURL := apiURL
	apiURL = server.URL
	defer func() { apiURL = originalAPIURL }()

	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GH_PAT", "test_pat")
	defer os.Unsetenv("GITHUB_REPOSITORY")
	defer os.Unsetenv("GH_PAT")

	err := UpdateActionsVariable("STRAVA_REFRESH_TOKEN", "new_refresh_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateActionsVariable_Non2xxResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"error": "invalid variable name"}`))
	}))
	defer server.Close()

	originalAPIURL := apiURL
	apiURL = server.URL
	defer func() { apiURL = originalAPIURL }()

	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GH_PAT", "test_pat")
	defer os.Unsetenv("GITHUB_REPOSITORY")
	defer os.Unsetenv("GH_PAT")

	err := UpdateActionsVariable("TEST_VAR", "test_value")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "422") {
		t.Errorf("expected status 422 in error, got %q", err.Error())
	}
}
