package main

import (
	"os"

	"github.com/jjkirkpatrick/clara/assistant"
	"github.com/jjkirkpatrick/clara/config"
	"github.com/jjkirkpatrick/clara/plugins"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

var cfg = config.New()
var openaiClient *openai.Client

func main() {
	if cfg.OpenAiAPIKey() == "" {
		key := os.Getenv("OPENAI_API_KEY")

		if key != "" {
			cfg = cfg.SetOpenAiAPIKey(key)
		} else {
			key := assistant.GetUserMessage("Please enter your OpenAI API key: ")
			cfg = cfg.SetOpenAiAPIKey(key)
		}
	}

	openaiClient = openai.NewClient(cfg.OpenAiAPIKey())

	if err := plugins.LoadPlugins(cfg, openaiClient); err != nil {
		log.Fatalf("Failed to load plugins: %v", err)
	}

	startAssistant()

}

func startAssistant() {
	functionDefinitions := plugins.GenerateOpenAIFunctionsDefinition()

	clara := assistant.Start(cfg, openaiClient, functionDefinitions)

	for {

		message := assistant.GetUserMessage("")
		clara.Message(message)
	}

}
