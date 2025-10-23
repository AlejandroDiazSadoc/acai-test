package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/acai-travel/tech-challenge/internal/chat/clients"
	"github.com/acai-travel/tech-challenge/internal/chat/models"
)

// WeatherTool provides real-time weather information for any city.
type WeatherTool struct {
	Client *clients.WeatherClient
}

// NewWeatherTool is a constructor for the weather tool.
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{
		clients.NewWeatherClient(os.Getenv("WEATHER_API")),
	}
}

func (tool *WeatherTool) Name() string {
	return "get_weather"
}

func (tool *WeatherTool) Description() string {
	return "Retrieves the current weather and forecast for a specific location."
}

func (tool *WeatherTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	// Params check
	location, exist := args["location"].(string)
	if !exist || location == "" {
		return "", fmt.Errorf("missing required argument 'location'")
	}

	days := 1
	if d, ok := args["days"].(float64); ok && d > 0 {
		days = int(d)
	}

	weatherUrl := fmt.Sprintf("%s/forecast.json?key=%s&q=%s&days=%d",
		tool.Client.BaseURL, tool.Client.APIKey, url.QueryEscape(location), days)

	req, err := http.NewRequestWithContext(ctx, "GET", weatherUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tool.Client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("weather API returned status code %d", resp.StatusCode)
	}

	var apiResponse models.WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("failed to decode weather API response: %w", err)
	}

	// Convert weather to json
	weatherBytes, err := json.Marshal(apiResponse)
	if err != nil {
		return "", fmt.Errorf("error converting response: %w", err)
	}
	weatherResultString := string(weatherBytes)

	return weatherResultString, nil
}

// GetFunctionSchema defines the tool's expected signature for openAI
func (tool *WeatherTool) GetFunctionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]string{
				"type":        "string",
				"description": "The location to get the weather from.",
			},
			"days": map[string]string{
				"type":        "integer",
				"description": "Optional number of forecast days to retrieve. Defaults to 1.",
			},
		},
		"required": []string{"location"},
	}
}
