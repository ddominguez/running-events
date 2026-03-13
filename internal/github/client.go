package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var apiURL = "https://api.github.com"

func UpdateActionsVariable(variableName, value string) error {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return fmt.Errorf("GITHUB_REPOSITORY not set")
	}
	pat := os.Getenv("GH_PAT")
	if pat == "" {
		return fmt.Errorf("GH_PAT not set")
	}

	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid GITHUB_REPOSITORY format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	url := fmt.Sprintf("%s/repos/%s/%s/actions/variables/%s",
		apiURL, owner, repoName, variableName)

	body, err := json.Marshal(map[string]string{
		"name":  variableName,
		"value": value,
	})
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+pat)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
