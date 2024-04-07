package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/wangergou2023/clara/assistant"
	"github.com/wangergou2023/clara/config"
)

var cfg = config.New()

func main() {
	cfg.AppLogger.Info("Clara is starting up... Please wait a moment.")

	config := openai.DefaultConfig(cfg.OpenAiAPIKey())
	//need"/v1"
	config.BaseURL = "https://llxspace.website/v1"
	openaiClient := openai.NewClientWithConfig(config)

	clara := assistant.Start(cfg, openaiClient)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		clara.Message(text)
	}
}
