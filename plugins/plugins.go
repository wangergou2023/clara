package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/sashabaranov/go-openai"
	"github.com/wangergou2023/clara/chatui"
	"github.com/wangergou2023/clara/config"
)

var loadedPlugins = make(map[string]Plugin)

type Plugin interface {
	Init(cfg config.Cfg, openaiClient *openai.Client, chat *chatui.ChatUI) error
	ID() string
	Description() string
	FunctionDefinition() openai.FunctionDefinition
	Execute(string) (string, error)
}

type PluginResponse struct {
	Error  string `json:"error,omitempty"`  // Contains error message if any error occurs.
	Result string `json:"result,omitempty"` // Contains result if successful.
}

func LoadPlugins(cfg config.Cfg, openaiClient *openai.Client, chat *chatui.ChatUI) error {
	loadedPlugins = make(map[string]Plugin)

	// Load plugins from compiled folder
	files, err := os.ReadDir(cfg.PluginsPath() + "/compiled")
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".so" {
			cfg.AppLogger.Info("Loading plugin: ", file.Name())
			err := loadSinglePlugin(cfg.PluginsPath()+"/compiled/"+file.Name(), cfg, openaiClient, chat)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func loadSinglePlugin(path string, cfg config.Cfg, openaiClient *openai.Client, chat *chatui.ChatUI) error {

	plugin, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := plugin.Lookup("Plugin")
	if err != nil {
		return err
	}

	p, ok := symbol.(*Plugin)
	if !ok {
		return fmt.Errorf("unexpected type from module symbol: %s", path)
	}
	err = (*p).Init(cfg, openaiClient, chat)
	if err != nil {
		return err
	}
	loadedPlugins[(*p).ID()] = *p
	return nil
}

// CallPlugin finds a plugin by its ID and executes it with the provided arguments.
func CallPlugin(id string, jsonInput string) (string, error) {
	response := PluginResponse{}

	plugin, exists := GetPluginByID(id)
	if !exists {
		response.Error = fmt.Sprintf("plugin with ID %s not found", id)
		jsonResponse, err := json.Marshal(response)
		return string(jsonResponse), err
	}

	result, err := plugin.Execute(jsonInput)
	if err != nil {
		response.Error = err.Error()
	} else {
		response.Result = result
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("error marshaling response to JSON: %v", err)
	}

	return string(jsonResponse), nil
}

func IsPluginLoaded(id string) bool {
	_, exists := loadedPlugins[id]
	return exists
}

func GetPluginByID(id string) (Plugin, bool) {
	p, exists := loadedPlugins[id]
	return p, exists
}

func GetAllPlugins() map[string]Plugin {
	return loadedPlugins
}

func GenerateOpenAIFunctionsDefinition() []openai.FunctionDefinition {
	var definitions []openai.FunctionDefinition

	for _, plugin := range loadedPlugins {
		def := plugin.FunctionDefinition()
		definitions = append(definitions, def)
	}

	return definitions
}
