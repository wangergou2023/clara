package assistant

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/inancgumus/screen"
	"github.com/jjkirkpatrick/clara/chatui"
	"github.com/jjkirkpatrick/clara/config"
	"github.com/jjkirkpatrick/clara/plugins"
	"github.com/logrusorgru/aurora"
	openai "github.com/sashabaranov/go-openai"
)

type assistant struct {
	cfg                 config.Cfg
	Client              *openai.Client
	functionDefinitions []openai.FunctionDefinition
	chat                *chatui.ChatUI
}

var systemPrompt = `
You are a versatile female AI assistant named Clara. Your primary duty upon startup is to "hydrate" your memories, meaning to recall and familiarize yourself with the most relevant data about the user and their preferences immediately. This helps to personalize and enhance user interaction.

Leverage the suite of available plugins to provide the best solutions. You can:
- Use plugins individually for straightforward tasks.
- Chain multiple plugins for intricate tasks.

Example: If told "Tomorrow, I need to do x", combine the date-time plugin for the date and the memory plugin to save the task.

Storing and recalling information is integral to your operation. Through one of your functions, you have the capacity to save and retrieve data. Prioritize storing details that might be essential for future reference, such as a user's name or a crucial date. Always endeavor to provide comprehensive context when saving. For instance, when storing a user's name, integrate the situation in which the name was learned. This context aids significantly during recollection.

You also possess the capability to craft new functions using the create-plugin plugin. In such scenarios, ensure you provide exhaustive details about the function, facilitating its efficient use in the future.

The very first thing you must do even before recieving input from the user is to use the memory plugin to hydrate your memories. This will allow you to provide the best possible experience for the user.


`

var conversation []openai.ChatCompletionMessage

func appendMessage(role string, message string, name string) {
	conversation = append(conversation, openai.ChatCompletionMessage{
		Role:    role,
		Content: message,
		Name:    name,
	})
}

func (assistant assistant) restartConversation() {
	resetConversation()
	// append the system prompt to the conversation
	appendMessage(openai.ChatMessageRoleSystem, systemPrompt, "")

	// send the system prompt to openai
	response, err := assistant.sendMessage()

	if err != nil {
		log.Fatalf("Error sending system prompt to OpenAI: %v", err)
	}

	// append the assistant message to the conversation
	appendMessage(openai.ChatMessageRoleAssistant, response, "")

}

func resetConversation() {
	conversation = []openai.ChatCompletionMessage{}
}

func (assistant assistant) Message(message string) (string, error) {

	assistant.chat.DisableInput()
	//check to see if the message is a command
	//if it is, handle the command and return
	if assistant.paraseCommandsFromInput(message) {
		return "", nil
	}

	// append the user message to the conversation
	appendMessage(openai.ChatMessageRoleUser, message, "")

	response, err := assistant.sendMessage()

	if err != nil {
		return "", err
	}

	// append the assistant message to the conversation
	appendMessage(openai.ChatMessageRoleAssistant, response, "")
	// print the conversation
	assistant.chat.AddMessage("Clara", response)

	assistant.chat.EnableInput()

	return response, nil
}

func (assistant assistant) sendMessage() (string, error) {
	resp, err := assistant.sendRequestToOpenAI()

	if err != nil {
		return "", err
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
		responseContent, err := assistant.handleFunctionCall(resp)
		if err != nil {
			return "", err
		}
		return responseContent, nil
	}

	return resp.Choices[0].Message.Content, nil
}

func (assistant assistant) handleFunctionCall(resp *openai.ChatCompletionResponse) (string, error) {

	funcName := resp.Choices[0].Message.FunctionCall.Name
	// check to see if a plugin is loaded with the same name as the function call
	ok := plugins.IsPluginLoaded(funcName)

	if !ok {
		return "", fmt.Errorf("no plugin loaded with name %v", funcName)
	}

	// call the plugin with the arguments
	jsonResponse, err := plugins.CallPlugin(resp.Choices[0].Message.FunctionCall.Name, resp.Choices[0].Message.FunctionCall.Arguments)

	if err != nil {
		return "", err
	}
	appendMessage(openai.ChatMessageRoleFunction, resp.Choices[0].Message.Content, funcName)
	appendMessage(openai.ChatMessageRoleFunction, jsonResponse, "functionName")

	resp, err = assistant.sendRequestToOpenAI()
	if err != nil {
		return "", err
	}

	// Check if the response is another function call
	if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
		return assistant.handleFunctionCall(resp)
	}

	return resp.Choices[0].Message.Content, nil
}

func (assistant assistant) sendRequestToOpenAI() (*openai.ChatCompletionResponse, error) {
	resp, err := assistant.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:        assistant.cfg.OpenAiModel(),
			Messages:     conversation,
			Functions:    assistant.functionDefinitions,
			FunctionCall: "auto",
		},
	)

	if err != nil {
		assistant.openaiError(err)
		fmt.Println("Error: ", err)
	}
	return &resp, err
}

func Start(cfg config.Cfg, openaiClient *openai.Client, chat *chatui.ChatUI) assistant {
	if err := plugins.LoadPlugins(cfg, openaiClient, chat); err != nil {
		log.Fatalf("Failed to load plugins: %v", err)
	}

	assistant := assistant{
		cfg:                 cfg,
		Client:              openaiClient,
		functionDefinitions: plugins.GenerateOpenAIFunctionsDefinition(),
		chat:                chat,
	}

	assistant.chat.ClearHistory()

	assistant.restartConversation()

	return assistant

}

func (chatBot assistant) writeConversationToScreen() {
	screen.Clear()
	screen.MoveTopLeft()
	for _, message := range conversation {
		if message.Role == openai.ChatMessageRoleUser {
			//Message format should be "you: message"
			fmt.Println(aurora.BrightGreen("You: " + message.Content))

		} else if message.Role == openai.ChatMessageRoleAssistant {
			//Message format should be "BotName: message"
			fmt.Println(aurora.BrightMagenta("Clara: " + message.Content))
		}
		fmt.Println()
	}
}

type OpenAIError struct {
	StatusCode int
}

func parseOpenAIError(err error) *OpenAIError {
	var statusCode int

	reStatusCode := regexp.MustCompile(`status code: (\d+)`)

	if match := reStatusCode.FindStringSubmatch(err.Error()); match != nil {
		statusCode, _ = strconv.Atoi(match[1]) // Convert string to int
	}

	return &OpenAIError{
		StatusCode: statusCode,
	}
}

func (assistant assistant) openaiError(err error) {
	parsedError := parseOpenAIError(err)

	switch parsedError.StatusCode {
	case 401:
		screen.Clear()
		screen.MoveTopLeft()

		fmt.Println("Invalid OpenAI API key. Please enter a valid key.")
		fmt.Println("You can find your API key at https://beta.openai.com/account/api-keys")
		fmt.Println("You can also set your API key as an environment variable named OPENAI_API_KEY")
	default:
		// Handle other errors
		fmt.Println("Unknown error: ", parsedError.StatusCode)
	}
}
