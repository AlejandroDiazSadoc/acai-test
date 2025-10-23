package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client working with amadeus
type FlightClient struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	BaseUrl      string
	TokenUrl     string
	TokenExpiry  time.Time
	HTTPClient   *http.Client
}

func NewFlightClient(id, secret string) *FlightClient {
	return &FlightClient{
		ClientID:     id,
		ClientSecret: secret,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
		BaseUrl:      "https://test.api.amadeus.com",
		TokenUrl:     "v1/security/oauth2/token",
	}
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *FlightClient) GetToken(ctx context.Context) error {
	// Check if the current token is still valid
	if time.Now().Before(c.TokenExpiry) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)

	body := strings.NewReader(data.Encode())

	fullURL := c.BaseUrl + "/" + c.TokenUrl
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, body)
	if err != nil {
		return fmt.Errorf("could not create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(errorBody))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" || tokenResp.ExpiresIn == 0 {
		return errors.New("token response missing access_token or expires_in")
	}

	c.AccessToken = tokenResp.AccessToken
	c.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}
