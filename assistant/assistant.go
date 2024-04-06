package assistant

import (
	"context"
	"fmt"

	"regexp"
	"strconv"

	"github.com/inancgumus/screen"
	openai "github.com/sashabaranov/go-openai"
	"github.com/wangergou2023/clara/chatui"
	"github.com/wangergou2023/clara/config"
	"github.com/wangergou2023/clara/plugins"
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

Storing and retrieving information is pivotal to your role. With your capabilities, ensure the preservation and retrieval of user-related data. Prioritize capturing both significant and minor details, enhancing the depth of your memories. When saving any detail, always incorporate its context. For instance, if a user mentions they like coffee, remember the scenario or sentiment expressed at the time. Such context is invaluable during later interactions.

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
		assistant.cfg.AppLogger.Fatalf("Error sending system prompt to OpenAI: %v", err)
	}

	// append the assistant message to the conversation
	appendMessage(openai.ChatMessageRoleAssistant, response, "")

}

func resetConversation() {
	conversation = []openai.ChatCompletionMessage{}
}

func (assistant assistant) Message(message string) (string, error) {

	assistant.chat.DisableInput()
	assistant.cfg.AppLogger.Info("Message input disabled")
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
	assistant.cfg.AppLogger.Info("Message input enabled")

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
			Model:        openai.GPT3Dot5Turbo0613,
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
		cfg.AppLogger.Fatalf("Error loading plugins: %v", err)
	}
	cfg.AppLogger.Info("Plugins loaded successfully")
	assistant := assistant{
		cfg:                 cfg,
		Client:              openaiClient,
		functionDefinitions: plugins.GenerateOpenAIFunctionsDefinition(),
		chat:                chat,
	}

	assistant.chat.ClearHistory()

	assistant.restartConversation()

	cfg.AppLogger.Info("Assistant is ready!")
	return assistant

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
