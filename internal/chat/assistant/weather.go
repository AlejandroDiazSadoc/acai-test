package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type WeatherClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

type WeatherAPIResponse struct {
	Location struct {
		Name    string `json:"name"`
		Region  string `json:"region"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		WindKph   float64 `json:"wind_kph"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Date string `json:"date"`
			Day  struct {
				MaxtempC  float64 `json:"maxtemp_c"`
				MintempC  float64 `json:"mintemp_c"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func NewWeatherClient(apiKey string) *WeatherClient {
	return &WeatherClient{
		APIKey:     apiKey,
		BaseURL:    "http://api.weatherapi.com/v1",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func getWeatherAndForecast(ctx context.Context, apiKey string, location string) (*WeatherAPIResponse, error) {
	client := NewWeatherClient(apiKey)

	weather_url := fmt.Sprintf("%s/forecast.json?key=%s&q=%s&days=%q",
		client.BaseURL, client.APIKey, url.QueryEscape(location), 1)

	req, err := http.NewRequestWithContext(ctx, "GET", weather_url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Handle API errors like 400 (bad query) or 401 (invalid key)
		return nil, fmt.Errorf("weather API returned status code %d", resp.StatusCode)
	}

	var apiResponse WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode weather API response: %w", err)
	}

	return &apiResponse, nil

}
