package clients

import (
	"net/http"
	"time"
)

// WeatherClient works with weatherapi
type WeatherClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewWeatherClient creates a new WeatherClient instance.
func NewWeatherClient(apiKey string) *WeatherClient {
	return &WeatherClient{
		APIKey:     apiKey,
		BaseURL:    "http://api.weatherapi.com/v1",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}
