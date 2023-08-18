package chatbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func init() {
	loadMemoryFromDisk()
}

var Functions = []openai.FunctionDefinition{
	{
		Name:        "storeMemory",
		Description: "Store a memory such as if the user gave infomation that might need to be referenced later",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"name": {
					Type:        jsonschema.String,
					Description: "the name of the memory to store",
				},
				"memory": {
					Type:        jsonschema.String,
					Description: "the memory to store",
				},
			},
			Required: []string{"name", "memory"},
		},
	},
	{
		Name:        "retrieveMemory",
		Description: "Retrieve a memory that was stored",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"name": {
					Type:        jsonschema.String,
					Description: "the name of the memory to retrieve",
				},
			},
			Required: []string{"name"},
		},
	},
}

type storeMemoryArgs struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

type retrieveMemoryArgs struct {
	Name string `json:"name"`
}

const memoryFilePath = "memories.json"

var memoryStorage = make(map[string]string)
var memoryMutex sync.RWMutex

func storeMemory(args string) (string, error) {
	debugMessageToLog("Storing memory..." + args)
	var margs storeMemoryArgs
	err := json.Unmarshal([]byte(args), &margs)
	if err != nil {
		return "", err
	}

	//TODO: Memories should not overwrite existing memories, but instead append to them
	//TODO: Memories in a set should be able to be updated i.e hobbies: ["programming", "reading"] should be able to be updated to hobbies: ["programming", "reading", "gaming"]

	// Lock the map for writing
	memoryMutex.Lock()
	memoryStorage[margs.Name] = margs.Memory
	memoryMutex.Unlock()

	// Now, save the updated memory to disk
	if err := saveMemoryToDisk(); err != nil {
		return "", err
	}

	return "Memory has been stored successfully.", nil
}

func retrieveMemory(args string) (string, error) {
	var margs retrieveMemoryArgs
	err := json.Unmarshal([]byte(args), &margs)
	if err != nil {
		return "", err
	}

	// Ensure memories are loaded from disk
	if err := loadMemoryFromDisk(); err != nil {
		return "", err
	}

	// Lock the map for reading
	memoryMutex.RLock()
	memory, exists := memoryStorage[margs.Name]
	memoryMutex.RUnlock()

	if !exists {
		return "Memory not found with name: " + margs.Name, fmt.Errorf("Memory with name %s not found", margs.Name)
	}

	return memory, nil
}

func saveMemoryToDisk() error {
	memoryMutex.RLock()
	data, err := json.Marshal(memoryStorage)
	memoryMutex.RUnlock()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(memoryFilePath, data, 0644)
}

func loadMemoryFromDisk() error {
	if _, err := os.Stat(memoryFilePath); os.IsNotExist(err) {
		return nil // file doesn't exist yet, so nothing to load
	}

	data, err := ioutil.ReadFile(memoryFilePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &memoryStorage)
}
