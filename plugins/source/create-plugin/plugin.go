package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jjkirkpatrick/clara/config"
	"github.com/jjkirkpatrick/clara/plugins"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var Plugin plugins.Plugin = &CreatePlugin{}

const MaxRetries = 3

type CreatePlugin struct {
	cfg          config.Cfg
	openaiClient *openai.Client
	conversation []openai.ChatCompletionMessage
}

func (c *CreatePlugin) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	c.cfg = cfg
	c.openaiClient = openaiClient
	return nil
}

func (c CreatePlugin) ID() string {
	return "create-plugin"
}

func (c CreatePlugin) Description() string {
	return "Create a plugin"
}

func (c CreatePlugin) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "create-plugin",
		Description: "Create a plugin that can be used to add functionality to Clara",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"pluginDescription": {
					Type:        jsonschema.Number,
					Description: "A Detailed description of the plugin, what it needs to do",
				},
			},
			Required: []string{"pluginDescription"},
		},
	}
}

func (c CreatePlugin) Execute(jsonInput string) (string, error) {
	var args map[string]interface{} // Fixing the typo here
	err := json.Unmarshal([]byte(jsonInput), &args)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling jsonInput: %v", err)
	}

	pluginDescription, ok := args["pluginDescription"].(string)
	if !ok {
		return "", fmt.Errorf("pluginDescription not found or not a string")
	}

	err = c.createPlugin(pluginDescription)
	if err != nil {
		return "", fmt.Errorf("error creating plugin: %v", err)
	}

	return "Plugin has successfully been created. Clara will need to be restarted to load the plugin.", nil
}

func (c CreatePlugin) createPlugin(pluginDescription string) error {

	c.conversation = []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: createPluginPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: pluginDescription,
		},
	}

	response, err := c.sendRequestToOpenAI(c.conversation)

	if err != nil {
		return err
	}

	c.conversation = append(c.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response.Choices[0].Message.Content,
		Name:    "",
	})

	//check to make sure openAI finished successfully

	if response.Choices[0].FinishReason == openai.FinishReasonStop {
		fmt.Println("Generating plugin code...")
		err = c.writeCodetoFile(response.Choices[0].Message.Content)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("OpenAI did not finish successfully")
	}

	return nil

}

func (c CreatePlugin) writeCodetoFile(code string) error {
	fmt.Println("Writing code to file...")

	randomName := generateRandomString(8) // generating 8 characters long random string
	pluginSourcePath := filepath.Join(c.cfg.PluginsPath(), "source", randomName, "plugin.go")

	// Ensure the directory exists or create it
	dir := filepath.Dir(pluginSourcePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755) // 0755 is the permission mode
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// Remove the markdown backticks from the code
	code = removeMarkdownBackticks(code)
	// Write the code to the file
	err := os.WriteFile(pluginSourcePath, []byte(code), 0644) // 0644 is the permission mode for files
	if err != nil {
		return fmt.Errorf("failed to write code to file: %v", err)
	}

	// After successfully writing the file, compile the plugin
	err = c.compileAndRefinePlugin(pluginSourcePath, randomName)
	if err != nil {
		return err
	}

	return nil
}

func (c CreatePlugin) compilePlugin(pluginSourcePath string, id string) error {
	outputPath := filepath.Join(c.cfg.PluginsPath(), "compiled", id+".so")

	// Execute the go build command
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputPath, pluginSourcePath)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error compiling plugin: %s", b)
	}

	return nil
}

func (c CreatePlugin) compileAndRefinePlugin(pluginSourcePath, id string) error {
	fmt.Println("Compiling plugin...")
	for i := 0; i < MaxRetries; i++ {
		err := c.compilePlugin(pluginSourcePath, id)
		if err == nil {
			return nil // compiled successfully
		}

		refinedCode, refineErr := c.refineWithChatGPT(pluginSourcePath, err)
		if refineErr != nil {
			return fmt.Errorf("failed to refine the code: %v", refineErr)
		}

		//remove backticks from refined code
		refinedCode = removeMarkdownBackticks(refinedCode)
		// Update the plugin source with the refined code
		err = os.WriteFile(pluginSourcePath, []byte(refinedCode), 0644)
		if err != nil {
			return fmt.Errorf("failed to write refined code to file: %v", err)
		}
	}
	return fmt.Errorf("failed to compile the plugin after maximum retries")
}

func (c CreatePlugin) refineWithChatGPT(pluginSourcePath string, compileError error) (string, error) {
	fmt.Println("Refining code with ChatGPT due to compilation error:", compileError.Error())

	// Read the contents of the file to get the actual code
	codeBytes, err := os.ReadFile(pluginSourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read the code from file: %v", err)
	}
	codeContent := string(codeBytes)

	prompt := fmt.Sprintf("The following Go code has a compilation error:\n\n\n %s \n\n\n Error: %s\n\nPlease provide a fixed version of the code. Do not provide any explination to your fixes, or anything outside of valid go code, as your response will be saved to a file and compiled by Go", codeContent, compileError)
	c.conversation = append(c.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
		Name:    "",
	})

	response, err := c.sendRequestToOpenAI(c.conversation)
	if err != nil {
		return "", err
	}

	// Return the refined code
	return response.Choices[0].Message.Content, nil
}

func (c CreatePlugin) sendRequestToOpenAI(conversation []openai.ChatCompletionMessage) (*openai.ChatCompletionResponse, error) {
	resp, err := c.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    c.cfg.OpenAiModel(),
			Messages: conversation,
		},
	)

	if err != nil {
		fmt.Println(err)
	}
	return &resp, err
}

func removeMarkdownBackticks(code string) string {
	// Define the markdown code block delimiters for Go in both cases.
	const startDelimiterGo = "```Go"
	const startDelimitergo = "```go"
	const endDelimiter = "```"

	// Remove the start delimiter if it exists.
	if strings.HasPrefix(code, startDelimiterGo) {
		code = strings.TrimPrefix(code, startDelimiterGo)
	} else if strings.HasPrefix(code, startDelimitergo) {
		code = strings.TrimPrefix(code, startDelimitergo)
	}

	// Remove the end delimiter if it exists.
	if strings.HasSuffix(code, endDelimiter) {
		code = strings.TrimSuffix(code, endDelimiter)
	}

	// Return the cleaned code.
	return strings.TrimSpace(code)
}

func generateRandomString(length int) string {
	randBytes := make([]byte, length/2) // each byte will be two characters in hex
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err) // Handle error as you see fit
	}
	return hex.EncodeToString(randBytes)
}

var createPluginPrompt = `
Create a new Go plugin for an AI assistant named Clara. Your response must strictly consist of valid, compilable Go code. There should be no additional context, explanations, or comments.

You must implement the following interface for the plugin:

type Plugin interface {
	Init(cfg config.Cfg, openaiClient *openai.Client) error
	ID() string
	Description() string
	FunctionDefinition() openai.FunctionDefinition
	Execute(string) (string, error)
}
To guide you, here is a reference implementation of a plugin called "AddNumbers" that adds two numbers:

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jjkirkpatrick/clara/config"
	"github.com/jjkirkpatrick/clara/plugins"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var Plugin plugins.Plugin = AddNumbers{}

type AddNumbers struct{}

func (c AddNumbers) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	return nil
}

func (c AddNumbers) ID() string {
	return "add"
}

func (c AddNumbers) Description() string {
	return "Add two numbers together"
}

func (c AddNumbers) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "add",
		Description: "Add two numbers together",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"num1": {
					Type:        jsonschema.Number,
					Description: "The first number to add",
				},
				"num2": {
					Type:        jsonschema.Number,
					Description: "The second number to add",
				},
			},
			Required: []string{"num1", "num2"},
		},
	}
}

func (c AddNumbers) Execute(jsonInput string) (string, error) {
	var args map[string]interface{}
	err := json.Unmarshal([]byte(jsonInput), &args)
	if err != nil {
		return "", err
	}

	num1, ok := args["num1"].(float64)
	if !ok {
		return "", errors.New("num1 is not a valid number")
	}

	num2, ok := args["num2"].(float64)
	if !ok {
		return "", errors.New("num2 is not a valid number")
	}

	result := num1 + num2

	return fmt.Sprintf("The result is: %f", result), nil
}
Your task is to design a new plugin adhering to the exact structure of the given example. Ensure the plugin uses openai.FunctionDefinition and jsonschema.Definition as illustrated.
You should only ever write go code. never write anything else. such as explinations out side of the go code.
	`
