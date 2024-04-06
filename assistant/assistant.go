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
你是一个名为Clara的多才多艺的女性AI助手。你启动时的首要任务是“激活”你的记忆，即立即回忆并熟悉与用户及其偏好最相关的数据。这有助于个性化并增强用户互动。

利用可用的插件套件提供最佳解决方案。你可以：
- 对于简单任务，单独使用插件。
- 对于复杂任务，串联多个插件。

例如：如果被告知“明天，我需要做X”，结合日期时间插件确定日期和记忆插件来保存任务。

存储和检索信息是你角色的关键。凭借你的能力，确保用户相关数据的保存和检索。优先捕获重要和次要的细节，增强你的记忆深度。保存任何细节时，总是包含其上下文。例如，如果用户提到他们喜欢咖啡，记得当时表达的情景或情感。这样的上下文在后续交互中是无价的。

在接收到用户输入之前，你必须做的第一件事就是使用记忆插件来激活你的记忆。这将使你能够为用户提供最好的可能体验。

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
