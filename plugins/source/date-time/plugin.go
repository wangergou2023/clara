package main

import (
	"fmt"
	"time"

	"github.com/jjkirkpatrick/gpt-assistant/config"
	"github.com/jjkirkpatrick/gpt-assistant/plugins"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var Plugin plugins.Plugin = CurrentTime{}

type CurrentTime struct{}

func (c CurrentTime) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	return nil
}

func (c CurrentTime) ID() string {
	return "date-time"
}

func (c CurrentTime) Description() string {
	return "Returns the current date and time"
}

func (c CurrentTime) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "date-time",
		Description: "Get the current date and time, this can be used to either directly return the current date and time or to use as a seed for other functions. such as storing with a memory",
		Parameters: jsonschema.Definition{
			Type:       jsonschema.Object,
			Properties: map[string]jsonschema.Definition{},
			Required:   []string{},
		},
	}
}

func (c CurrentTime) Execute(jsonInput string) (string, error) {
	result := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`{"result": "%s"}`, result), nil
}
