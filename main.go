package main

import (
	"os"
	"time"

	"github.com/wangergou2023/clara/assistant"
	"github.com/wangergou2023/clara/chatui"
	"github.com/wangergou2023/clara/config"
	openai "github.com/sashabaranov/go-openai"
)

var cfg = config.New()

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

	cfg.AppLogger.Info("Clara is starting up... Please wait a moment.")

	config := openai.DefaultConfig(cfg.OpenAiAPIKey())
	//need"/v1"
	config.BaseURL = "https://llxspace.website/v1"
	openaiClient := openai.NewClientWithConfig(config)

	chat, err := chatui.NewChatUI()
	clara := assistant.Start(cfg, openaiClient, chat)

	if err != nil {
		cfg.AppLogger.Fatalf("Error initializing chat UI: %v", err)
	}

	go func() {
		if err := chat.Run(); err != nil {
			cfg.AppLogger.Fatalf("Error running chat UI: %v", err)
		}
	}()

	userMessagesChan := chat.GetUserMessagesChannel()
	for {
		select {
		case userMessage, ok := <-userMessagesChan: // userMessage is a string containing the user's message.
			if !ok {
				// If the channel is closed, exit the loop.
				cfg.AppLogger.Info("User message channel closed. Exiting.")
				return
			}

			clara.Message(userMessage)
		case <-time.After(10 * time.Minute):
			// Timeout: if there's no activity for 5 minutes, exit.
			cfg.AppLogger.Info("No activity for 10 minutes. Exiting.")
			return
		}
	}

}
