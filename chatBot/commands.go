package chatbot

import (
	"fmt"
	"os"

	"github.com/inancgumus/screen"
)

// list of commands

var commands = []string{
	"/restart",
	"/help",
	"/exit",
	"/clear",
}

func paraseCommandsFromInput(message string) bool {

	// check to see if the message is a command
	for _, command := range commands {
		if message == command {
			// handle the command
			handleCommand(message)
			return true
		}
	}
	return false
}

func handleCommand(command string) {
	switch command {
	case "/restart":
		chatBot.restartConversation()
	case "/help":
		fmt.Println("Help message")
	case "/exit":
		os.Exit(0)
	case "/clear":
		screen.Clear()
		screen.MoveTopLeft()
	}
}
