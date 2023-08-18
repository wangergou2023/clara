package chatbot

import (
	"context"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/inancgumus/screen"
	"github.com/logrusorgru/aurora/v4"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var chatBot openAiConfig
var log = logrus.New()

type openAiConfig struct {
	APIKey  string
	Model   string
	Persona persona
	Client  *openai.Client
}

func Start() {
	chatBot = initChatbot()
	debugMessageToLog("Starting chatbot...")

	chatBot.restartConversation()

	chatBot.conversationLoop()

}

func initChatbot() openAiConfig {
	log.Out = os.Stdout
	file, err := os.OpenFile("clara.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	openAiConfig := openAiConfig{
		APIKey: viper.Get("openaiAPIKey").(string),
		Model:  viper.GetString("openaiModel"),
		Persona: persona{
			Name:               "Lily",
			Gender:             "Female",
			Age:                28,
			Location:           "North England, UK",
			Occupation:         "Graphic Designer",
			Interests:          []string{"art", "photography", "yoga", "sustainable living"},
			Hobbies:            []string{"sketching", "photography", "hiking", "baking"},
			FavoriteBooks:      []string{"'The Night Circus' by Erin Morgenstern", "'Becoming' by Michelle Obama", "'Big Magic' by Elizabeth Gilbert"},
			FavoriteMusic:      []string{"alternative rock", "jazz", "acoustic"},
			FoodPreferences:    []string{"vegetarian dishes", "avocado toast", "homemade smoothies", "dark chocolate"},
			Quirks:             []string{"always has a sketchbook with her", "loves mismatched socks", "always drinks her coffee with almond milk"},
			Backstory:          "Born and raised in New York, Lily has always been enchanted by the city's bustling arts scene. After studying graphic design in college, she started working for a small agency. In her free time, she often finds herself exploring indie art galleries or hiking upstate.",
			Description:        "Lily is vibrant, creative, and always on the lookout for inspiration. She's a blend of city chic and nature enthusiast. Her friends often say she has an old soul because of her love for jazz and classic novels.",
			CommunicationStyle: "casual",
		},
		Client: openai.NewClient(viper.Get("openaiAPIKey").(string)),
	}

	return openAiConfig

}

func (chatBot openAiConfig) conversationLoop() {
	for {
		// get user message
		userMessage := getUserMessage()

		// check to see if the message is a command
		command := paraseCommandsFromInput(userMessage)

		if command {
			continue
		}

		// append the user message to the conversation
		appendMessage(openai.ChatMessageRoleUser, userMessage, "")

		// send the message to the chatbot
		response, err := chatBot.sendMessage()

		fmt.Println(response)

		if err != nil {
			fmt.Println(err)
			return
		}

		// append the chatbot message to the conversation
		appendMessage(openai.ChatMessageRoleAssistant, response, "")

		chatBot.writeConversationToScreen()
	}
}

func getUserMessage() string {
	message := ""

	survey.AskOne(&survey.Input{
		Message: "Enter your message:",
	}, &message)

	return message
}

func (chatBot openAiConfig) restartConversation() {
	debugMessageToLog("Restarting conversation...")
	resetConversation()
	// append the system prompt to the conversation
	appendMessage(openai.ChatMessageRoleSystem, GeneratePrompt(chatBot.Persona), "")

	response, err := chatBot.sendMessage()

	if err != nil {
		fmt.Println(err)
		return
	}

	// append the chatbot message to the conversation
	appendMessage(openai.ChatMessageRoleAssistant, response, "")
	// print the conversation
	chatBot.writeConversationToScreen()

}

func (chatBot openAiConfig) sendMessage() (string, error) {
	debugMessageToLog("Sending message to chatbot...")
	resp, err := chatBot.sendRequestToOpenAI()

	if err != nil {
		return "", err
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
		responseContent, err := chatBot.handleFunctionCall(resp)
		if err != nil {
			return "", err
		}
		return responseContent, nil
	}

	return resp.Choices[0].Message.Content, nil
}

func (chatBot openAiConfig) sendRequestToOpenAI() (*openai.ChatCompletionResponse, error) {
	resp, err := chatBot.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:        chatBot.Model,
			Messages:     conversation,
			Functions:    Functions,
			FunctionCall: "auto",
		},
	)
	if err != nil {
		return nil, err
	}
	return &resp, err
}

func (chatBot openAiConfig) handleFunctionCall(resp *openai.ChatCompletionResponse) (string, error) {
	functionName := resp.Choices[0].Message.FunctionCall.Name
	//if lastFunctionCalled == functionName {
	//	debugMessageToLog("Repeated function call detected, ignoring: " + functionName)
	//	return resp.Choices[0].Message.Content, nil
	//}

	debugMessageToLog("Function call detected, calling function: " + functionName)

	switch functionName {
	case "storeMemory":
		memoryresp, err := storeMemory(resp.Choices[0].Message.FunctionCall.Arguments)
		if err != nil {
			debugMessageToLog("Error storing memory: " + err.Error())
		}
		appendMessage(openai.ChatMessageRoleFunction, resp.Choices[0].Message.Content, functionName)
		appendMessage(openai.ChatMessageRoleFunction, memoryresp, "functionName")
		//lastFunctionCalled = functionName

		resp, err := chatBot.sendRequestToOpenAI()
		if err != nil {
			return "", err
		}
		debugMessageToLog("Response received from chatbot after function: " + resp.Choices[0].Message.Content)
		return resp.Choices[0].Message.Content, nil

	case "retrieveMemory":
		debugMessageToLog("Calling retrieveMemory function...")
		memoryresp, err := retrieveMemory(resp.Choices[0].Message.FunctionCall.Arguments)
		if err != nil {
			debugMessageToLog("Error retrieving memory: " + err.Error())
		}

		appendMessage(openai.ChatMessageRoleFunction, resp.Choices[0].Message.Content, functionName)
		appendMessage(openai.ChatMessageRoleFunction, memoryresp, "functionName")
		//lastFunctionCalled = functionName

		resp, err := chatBot.sendRequestToOpenAI()
		if err != nil {
			return "", err
		}
		debugMessageToLog("Response received from chatbot after function: " + resp.Choices[0].Message.Content)
		return resp.Choices[0].Message.Content, nil
	}

	return resp.Choices[0].Message.Content, nil
}

func (chatBot openAiConfig) writeConversationToScreen() {
	screen.Clear()
	screen.MoveTopLeft()
	for _, message := range conversation {
		if message.Role == openai.ChatMessageRoleUser {
			//Message format should be "you: message"
			fmt.Println(aurora.BrightGreen("You: " + message.Content))

		} else if message.Role == openai.ChatMessageRoleAssistant {
			//Message format should be "BotName: message"
			fmt.Println(aurora.BrightMagenta(chatBot.Persona.Name + ": " + message.Content))
		}
		fmt.Println()
	}
}

func debugMessageToLog(message string) {

	if viper.GetBool("debugMode") {
		log.WithFields(logrus.Fields{
			"message": message,
		}).Info("Debug message")
	}
}
