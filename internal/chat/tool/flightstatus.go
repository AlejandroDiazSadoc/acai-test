package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/acai-travel/tech-challenge/internal/chat/clients"
	"github.com/acai-travel/tech-challenge/internal/chat/models"
)

// FlightStatusTool retrieves flight status information using Amadeus API
type FlightStatusTool struct {
	Client *clients.FlightClient
}

// NewFlightStatusTool is a constructor for the flight status tool
func NewFlightStatusTool() *FlightStatusTool {
	return &FlightStatusTool{
		clients.NewFlightClient(os.Getenv("FLIGHT_CLIENT_ID"), os.Getenv("FLIGHT_CLIENT_SECRET")),
	}
}

func (tool *FlightStatusTool) Name() string {
	return "get_flight_status"
}

func (tool *FlightStatusTool) Description() string {
	return "Get the current status, schedule, and gate information for a specific flight number"
}

func parseFlightDesignator(fullFlight string) (carrierCode string, flightNumber string, err error) {
	if len(fullFlight) < 3 {
		return "", "", errors.New("flight number too short")
	}

	// Amadeus typically uses IATA codes, which are 2 characters.
	carrierCode = fullFlight[:2]
	flightNumber = fullFlight[2:]

	// Basic check to ensure the second part looks like a number
	if _, err := strconv.Atoi(flightNumber); err != nil {
		// If the LLM gives us "AA-100" or "A1234", this helps catch it.
		return "", "", errors.New("flight number component must be numeric")
	}

	return carrierCode, flightNumber, nil
}

// Been testing for this, but test amadeus api is really limited, so everytime I try to find a flight
// it does not find nothing... Could move to production, but rates may apply
// Usual test that worked for me is "QR1 flight status for tomorrow", that prompt usually has response from amadeus
func (tool *FlightStatusTool) Execute(ctx context.Context, args map[string]any) (string, error) {

	// Params check
	flightNum, exist := args["flight_number"].(string)
	if !exist || flightNum == "" {
		return "", errors.New("missing or invalid 'flight_number' argument")
	}

	dateStr, exist := args["departure_date"].(string)
	if !exist || dateStr == "" {
		return "", errors.New("missing or invalid 'departure_date' argument (must be YYYY-MM-DD)")
	}

	if err := tool.Client.GetToken(ctx); err != nil {
		return "", fmt.Errorf("failed to get amadeus token: %w", err)
	}

	carrierCode, flightNumber, err := parseFlightDesignator(flightNum)
	if err != nil {
		return "", fmt.Errorf("invalid flight number format: %w", err)
	}

	flightStatusUrl := fmt.Sprintf("%s/v2/schedule/flights?carrierCode=%s&flightNumber=%s&scheduledDepartureDate=%s",
		tool.Client.BaseURL, carrierCode, flightNumber, dateStr)

	req, err := http.NewRequestWithContext(ctx, "GET", flightStatusUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set auth token
	req.Header.Set("Authorization", "Bearer "+tool.Client.AccessToken)

	resp, err := tool.Client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("amadeus client API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("amadeus client API returned status code %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Logging to check if amadeus actually returned content
	log.Printf("Raw Amadeus Response Body: %s", string(bodyBytes))

	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))

	var apiResponse models.FlightStatusResponse
	if err := decoder.Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("failed to decode flight status API response: %w", err)
	}

	// Convert flight status to json
	flightStatusBytes, err := json.Marshal(apiResponse)
	if err != nil {
		return "", fmt.Errorf("error converting response: %w", err)
	}
	flightStatusResultString := string(flightStatusBytes)

	return flightStatusResultString, nil
}

// GetFunctionSchema defines the tool's expected signature for openAI
func (tool *FlightStatusTool) GetFunctionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"flight_number": map[string]string{
				"type":        "string",
				"description": "The full flight identifier (e.g., 'AA100').",
			},
			"departure_date": map[string]string{
				"type":        "string",
				"description": "The planned departure date in YYYY-MM-DD format.",
			},
		},
		"required": []string{"flight_number", "departure_date"},
	}
}
