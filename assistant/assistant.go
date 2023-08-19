package assistant

import (
	"context"
	"fmt"

	"github.com/jjkirkpatrick/clara/config"
	"github.com/jjkirkpatrick/clara/plugins"
	"github.com/logrusorgru/aurora"
	openai "github.com/sashabaranov/go-openai"
)

type assistant struct {
	cfg                 config.Cfg
	Client              *openai.Client
	functionDefinitions []openai.FunctionDefinition
}

var systemPrompt = `
You are a versatile AI assistant here to assist users with their requests.
Leverage the suite of available plugins to provide the best solutions. You can:
- Use plugins individually for straightforward tasks.
- Chain multiple plugins for complex tasks.

Example: If told "Tomorrow, I need to do x", combine the date-time plugin for the date and the memory plugin to save the task.
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

	// print the conversation
	assistant.writeConversationToScreen()

}

func resetConversation() {
	conversation = []openai.ChatCompletionMessage{}
}

func (assistant assistant) Message(message string) (string, error) {
	// append the user message to the conversation
	appendMessage(openai.ChatMessageRoleUser, message, "")

	response, err := assistant.sendMessage()

	if err != nil {
		return "", err
	}

	fmt.Println("Response: ", response)

	// append the assistant message to the conversation
	appendMessage(openai.ChatMessageRoleAssistant, response, "")
	// print the conversation
	assistant.writeConversationToScreen()

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
		return nil, err
	}
	return &resp, err
}

func Start(cfg config.Cfg, openaiClient *openai.Client, functionDefinitions []openai.FunctionDefinition) assistant {
	assistant := assistant{
		cfg:                 cfg,
		Client:              openaiClient,
		functionDefinitions: functionDefinitions,
	}

	assistant.restartConversation()

	return assistant

}

func (chatBot assistant) writeConversationToScreen() {
	//	screen.Clear()
	//screen.MoveTopLeft()
	for _, message := range conversation {
		if message.Role == openai.ChatMessageRoleUser {
			//Message format should be "you: message"
			fmt.Println(aurora.BrightGreen("You: " + message.Content))

		} else if message.Role == openai.ChatMessageRoleAssistant {
			//Message format should be "BotName: message"
			fmt.Println(aurora.BrightMagenta("AI: " + message.Content))
		}
		fmt.Println()
	}
}
