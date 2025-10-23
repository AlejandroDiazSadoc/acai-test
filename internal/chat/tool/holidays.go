package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

// HolidaysTool provides holidays.
type HolidaysTool struct{}

// NewHolidaysTool is a constructor for the holidays tool.
func NewHolidaysTool() *HolidaysTool {
	return &HolidaysTool{}
}

func (tool *HolidaysTool) Name() string {
	return "get_holidays"
}

func (tool *HolidaysTool) Description() string {
	return "Gets local bank and public holidays. Each line is a single holiday in the format 'YYYY-MM-DD: Holiday Name'."
}

func LoadCalendar(ctx context.Context, link string) ([]*ics.VEvent, error) {
	slog.InfoContext(ctx, "Loading calendar", "link", link)

	cal, err := ics.ParseCalendarFromUrl(link, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	return cal.Events(), nil
}

type Payload struct {
	BeforeDate time.Time `json:"before_date,omitempty"`
	AfterDate  time.Time `json:"after_date,omitempty"`
	MaxCount   int       `json:"max_count,omitempty"`
}

func parseArgs(args map[string]any) (Payload, error) {
	// First marshal the map to JSON
	raw, err := json.Marshal(args)
	if err != nil {
		return Payload{}, fmt.Errorf("failed to marshal args: %w", err)
	}

	// Then unmarshal into your struct
	var payload Payload
	err = json.Unmarshal(raw, &payload)
	if err != nil {
		return Payload{}, fmt.Errorf("failed to unmarshal into payload: %w", err)
	}

	return payload, nil
}

func (tool *HolidaysTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	link := "https://www.officeholidays.com/ics/spain/catalonia"
	if v := os.Getenv("HOLIDAY_CALENDAR_LINK"); v != "" {
		link = v
	}

	// Params check
	params, err := parseArgs(args)
	if err != nil {
		return "", fmt.Errorf("failed to parse tool call arguments: %w", err)
	}

	events, err := LoadCalendar(ctx, link)
	if err != nil {
		return "", fmt.Errorf("failed to load holiday events: %w", err)
	}

	var holidays []string
	for _, event := range events {
		date, err := event.GetAllDayStartAt()
		if err != nil {
			continue
		}

		if params.MaxCount > 0 && len(holidays) >= params.MaxCount {
			break
		}

		if !params.BeforeDate.IsZero() && date.After(params.BeforeDate) {
			continue
		}

		if !params.AfterDate.IsZero() && date.Before(params.AfterDate) {
			continue
		}

		holidays = append(holidays, date.Format(time.DateOnly)+": "+event.GetProperty(ics.ComponentPropertySummary).Value)
	}

	return strings.Join(holidays, "\n"), nil
}

// GetFunctionSchema defines the tool's expected signature for openAI
func (tool *HolidaysTool) GetFunctionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"before_date": map[string]string{
				"type":        "string",
				"description": "Optional date in RFC3339 format to get holidays before this date. If not provided, all holidays will be returned.",
			},
			"after_date": map[string]string{
				"type":        "string",
				"description": "Optional date in RFC3339 format to get holidays after this date. If not provided, all holidays will be returned.",
			},
			"max_count": map[string]string{
				"type":        "integer",
				"description": "Optional maximum number of holidays to return. If not provided, all holidays will be returned.",
			},
		},
	}
}
