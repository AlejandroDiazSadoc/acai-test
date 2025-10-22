package clients

import (
	"net/http"
	"time"
)

// Client working with weather api
type WeatherClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewWeatherClient(apiKey string) *WeatherClient {
	return &WeatherClient{
		APIKey:     apiKey,
		BaseURL:    "http://api.weatherapi.com/v1",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}
