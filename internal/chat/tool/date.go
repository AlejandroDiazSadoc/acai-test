package tool

import (
	"context"
	"time"
)

// DateTool provides actual date
type DateTool struct{}

// NewDateTool is a constructor for the date tool.
func NewDateTool() *DateTool {
	return &DateTool{}
}

func (tool *DateTool) Name() string {
	return "get_today_date"
}

func (tool *DateTool) Description() string {
	return "Get today's date and time in RFC3339 format."
}

func (tool *DateTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}

// GetFunctionSchema defines the tool's expected signature for openAI
func (tool *DateTool) GetFunctionSchema() map[string]any {
	return map[string]any{}
}
