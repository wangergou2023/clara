package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jjkirkpatrick/clara/assistant"
	"github.com/jjkirkpatrick/clara/chatui"
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

	fmt.Println("Clara is starting up... Please wait a moment.")

	openaiClient = openai.NewClient(cfg.OpenAiAPIKey())

	chat, err := chatui.NewChatUI()
	clara := assistant.Start(cfg, openaiClient, chat)

	if err != nil {
		log.Fatalf("Error initializing chat UI: %v", err)
	}

	go func() {
		if err := chat.Run(); err != nil {
			log.Fatalf("Error running chat UI: %v", err)
		}
	}()

	userMessagesChan := chat.GetUserMessagesChannel()
	for {
		select {
		case userMessage, ok := <-userMessagesChan: // userMessage is a string containing the user's message.
			if !ok {
				// If the channel is closed, exit the loop.
				return
			}

			clara.Message(userMessage)
		case <-time.After(10 * time.Minute):
			// Timeout: if there's no activity for 5 minutes, exit.
			fmt.Println("No activity for 10 minutes. Exiting.")
			return
		}
	}

}
