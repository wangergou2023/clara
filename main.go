package main

import (
	"github.com/jjkirkpatrick/gpt-assistant/assistant"
	"github.com/jjkirkpatrick/gpt-assistant/config"
	"github.com/jjkirkpatrick/gpt-assistant/plugins"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

var cfg = config.New()

var openaiClient *openai.Client

type inputDefinition struct {
	RequestType  string
	Memory       string
	Num_relevant int
}

func main() {
	openaiClient = openai.NewClient(cfg.OpenAiAPIKey())

	if err := plugins.LoadPlugins(cfg, openaiClient); err != nil {
		log.Fatalf("Failed to load plugins: %v", err)
	}

	functionDefinitions := plugins.GenerateOpenAIFunctionsDefinition()

	javis := assistant.Start(cfg, openaiClient, functionDefinitions)

	for {

		message := assistant.GetUserMessage()
		javis.Message(message)
	}

}

//func main() {
//	openaiClient = openai.NewClient(cfg.OpenAiAPIKey())
//
//	if err := plugins.LoadPlugins(cfg, openaiClient); err != nil {
//		log.Fatalf("Failed to load plugins: %v", err)
//	}
//
//	setRequest := inputDefinition{
//		RequestType: "set",
//		Memory:      "I like to eat pizza",
//	}
//
//	requestJosn, err := json.Marshal(setRequest)
//
//	if err != nil {
//		log.Fatalf("Failed to marshal request: %v", err)
//	}
//
//	jsonResponse, err := plugins.CallPlugin("memory", string(requestJosn))
//	if err != nil {
//		log.Fatalf("Failed to call plugin: %v", err)
//	}
//	log.Printf("Response: %v", jsonResponse)
//
//	getRequest := inputDefinition{
//		RequestType:  "get",
//		Num_relevant: 5,
//		Memory:       "pizza",
//	}
//
//	requestJosn, err = json.Marshal(getRequest)
//
//	if err != nil {
//		log.Fatalf("Failed to marshal request: %v", err)
//	}
//
//	jsonResponse, err = plugins.CallPlugin("memory", string(requestJosn))
//	if err != nil {
//		log.Fatalf("Failed to call plugin: %v", err)
//	}
//	log.Printf("Response: %v", jsonResponse)
//
//}
//
