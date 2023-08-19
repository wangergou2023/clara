package main

import (
	"encoding/json"
	"fmt"

	"github.com/jjkirkpatrick/gpt-assistant/config"
	"github.com/jjkirkpatrick/gpt-assistant/plugins"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var Plugin plugins.Plugin = AddNumbers{}

type AddNumbers struct{}

func (c AddNumbers) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	return nil
}

func (c AddNumbers) ID() string {
	return "add"
}

func (c AddNumbers) Description() string {
	return "Add two numbers together"
}

func (c AddNumbers) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "add",
		Description: "Add two numbers together",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"num1": {
					Type:        jsonschema.Number,
					Description: "The first number to add",
				},
				"num2": {
					Type:        jsonschema.Number,
					Description: "The second number to add",
				},
			},
			Required: []string{"num1", "num2"},
		},
	}
}

func (c AddNumbers) Execute(jsonInput string) (string, error) {
	var args map[string]any
	err := json.Unmarshal([]byte(jsonInput), &args)
	if err != nil {
		return "", err
	}

	num1, ok := args["num1"].(float64)
	if !ok {
		return "", fmt.Errorf("num1 is not a number")
	}

	num2, ok := args["num2"].(float64)
	if !ok {
		return "", fmt.Errorf("num2 is not a number")
	}

	result := num1 + num2

	return fmt.Sprintf(`{"result": %v}`, result), nil

}
