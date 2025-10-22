package tool

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v2"
)

type Tool interface {
	// Name returns the unique name of the tool (e.g., "get_weather")
	Name() string

	// Description returns description of the tool
	Description() string

	// GetFunctionSchema returns schema of tool for opeanAI
	GetFunctionSchema() map[string]any

	// Execute runs the core logic of the tool.
	// The map[string]any contains the arguments parsed by the model
	// It returns the result as a string for the model to incorporate into its final response
	Execute(ctx context.Context, args map[string]any) (string, error)
}

type Map map[string]Tool

func (tools Map) GetToolByName(name string) (Tool, error) {
	tool, exist := tools[name]
	if !exist {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  any
}

// OaiToolDefinitions converts the internal tool.Map into the OpenAI API format.
func (tools Map) OaiToolDefinitions() []openai.ChatCompletionToolUnionParam {
	var oaiDefs []openai.ChatCompletionToolUnionParam
	for _, tool := range tools {
		oaiDefs = append(oaiDefs, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        tool.Name(),
			Description: openai.String(tool.Description()),
			Parameters:  openai.FunctionParameters(tool.GetFunctionSchema()),
		}))
	}
	return oaiDefs
}
