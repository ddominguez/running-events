package strava

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const BaseURL = "https://www.strava.com/api/v3"

var tokenURL = "https://www.strava.com/oauth/token"

type Client struct {
	baseURL     string
	accessToken string
}

func NewClient(accessToken string) *Client {
	return &Client{
		baseURL:     BaseURL,
		accessToken: accessToken,
	}
}

func (c *Client) Get(endpoint string) ([]byte, error) {
	url := c.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return body, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func RefreshToken() (accessToken, refreshToken string, expiresAt int64, err error) {
	clientID := os.Getenv("STRAVA_CLIENT_ID")
	if clientID == "" {
		return "", "", 0, fmt.Errorf("STRAVA_CLIENT_ID not set")
	}
	clientSecret := os.Getenv("STRAVA_CLIENT_SECRET")
	if clientSecret == "" {
		return "", "", 0, fmt.Errorf("STRAVA_CLIENT_SECRET not set")
	}
	refreshToken = os.Getenv("STRAVA_REFRESH_TOKEN")
	if refreshToken == "" {
		return "", "", 0, fmt.Errorf("STRAVA_REFRESH_TOKEN not set")
	}

	data := fmt.Sprintf(
		"grant_type=refresh_token&client_id=%s&client_secret=%s&refresh_token=%s",
		clientID, clientSecret, refreshToken,
	)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return "", "", 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, fmt.Errorf("reading response: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", "", 0, fmt.Errorf("decoding response: %w", err)
	}

	return tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresAt, nil
}
