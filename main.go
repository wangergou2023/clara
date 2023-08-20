package main

import (
	"os"

	"github.com/jjkirkpatrick/clara/assistant"
	"github.com/jjkirkpatrick/clara/config"
	openai "github.com/sashabaranov/go-openai"
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

	startAssistant()

}

func startAssistant() {

	clara := assistant.Start(cfg, openaiClient)

	for {

		message := assistant.GetUserMessage("")
		clara.Message(message)
	}

}
