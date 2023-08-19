package main

import (
	"fmt"
	"math/rand"

	"github.com/jjkirkpatrick/gpt-assistant/config"
	"github.com/jjkirkpatrick/gpt-assistant/plugins"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var Plugin plugins.Plugin = RandomQuote{}

type RandomQuote struct{}

func (c RandomQuote) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	return nil
}

// list of quotes
var quotes = []string{
	"\"The best preparation for tomorrow is doing your best today.\" - H. Jackson Brown, Jr.",
	"\"Believe you can and you're halfway there.\" - Theodore Roosevelt",
	"\"It does not matter how slowly you go as long as you do not stop.\" - Confucius",
	"\"Our greatest weakness lies in giving up. The most certain way to succeed is always to try just one more time.\" - Thomas A. Edison",
}

func (c RandomQuote) ID() string {
	return "random-quote"
}

func (c RandomQuote) Description() string {
	return "Returns random quote"
}

func (c RandomQuote) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "random-quote",
		Description: "get a random quote from a list of quotes provided by the plugin author",
		Parameters: jsonschema.Definition{
			Type:       jsonschema.Object,
			Properties: map[string]jsonschema.Definition{},
			Required:   []string{},
		},
	}
}

func (c RandomQuote) Execute(jsonInput string) (string, error) {
	return fmt.Sprintf(`{"result": %v}`, quotes[rand.Intn(len(quotes))]), nil
}
