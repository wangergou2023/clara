package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jjkirkpatrick/clara/chatui"
	"github.com/jjkirkpatrick/clara/config"
	"github.com/jjkirkpatrick/clara/plugins"
	milvus "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var Plugin plugins.Plugin = &Memory{}

type Memory struct {
	cfg          config.Cfg
	milvusClient milvus.Client
	openaiClient *openai.Client
}

type memory struct {
	Memory string
	Vector []float32
}

type memoryResult struct {
	Memory string
	Type   string
	Detail string
	Score  float32
}

type memoryItem struct {
	Memory string `json:"memory"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

type inputDefinition struct {
	RequestType  string       `json:"requestType"`
	Memories     []memoryItem `json:"memories"`
	Num_relevant int          `json:"num_relevant"`
}

func (c *Memory) Init(cfg config.Cfg, openaiClient *openai.Client, chat *chatui.ChatUI) error {
	c.cfg = cfg
	c.openaiClient = openaiClient

	ctx := context.Background()

	c.milvusClient, _ = milvus.NewGrpcClient(ctx, c.cfg.MalvusApiEndpoint())

	err := c.initMilvusSchema()

	if err != nil {
		return err
	}

	return nil
}

func (c Memory) ID() string {
	return "memory"
}

func (c Memory) Description() string {
	return "store and retrieve memories from long term memory."
}

func (c Memory) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "memory",
		Description: "Store and retrieve memories from long term memory. Use requestType 'set' to add memories to the database, use requestType 'get' to retrieve the most relevant memories. When first starting you should use the 'hydrate' function to recall past memories about the user.'",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"requestType": {
					Type:        jsonschema.String,
					Description: "The type of request to make  'set' , 'get' or 'hydrate'. 'Set' will add memories to the database, 'get' will return the most relevant memories. when getting a memory, you should always include the memory fiel. 'hydrate' will return a prompt with all the memories about the user.",
				},
				"memories": {
					Type: jsonschema.Array,
					Items: &jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"memory": {
								Type:        jsonschema.String,
								Description: "The individual memory to add. You should provide as much context as possible to go along with the memory.",
							},
							"type": {
								Type:        jsonschema.String,
								Description: "The type of memory, for example: 'personality', 'food', etc.",
							},
							"detail": {
								Type:        jsonschema.String,
								Description: "Specific detail about the type, for example: 'likes pizza', 'is flirty', etc.",
							},
						},
						Required: []string{"memory", "type", "detail"},
					},
					Description: "The array of memories to add or get. Each memory contains its individual content, type, and detail. This is required for both 'set' and 'get' requests.",
				},
				"num_relevant": {
					Type:        jsonschema.Integer,
					Description: "The number of relevant memories to return, for example: 5.",
				},
			},
			Required: []string{"requestType"},
		},
	}
}

func (c Memory) Execute(jsonInput string) (string, error) {
	// marshal jsonInput to inputDefinition
	var args inputDefinition
	err := json.Unmarshal([]byte(jsonInput), &args)
	if err != nil {
		return "", err
	}

	if args.Num_relevant == 0 {
		args.Num_relevant = 5
	}

	// Check if memories slice is empty
	if args.RequestType != "hydrate" && len(args.Memories) == 0 {
		return fmt.Sprintf(`%v`, "memories are required but was empty"), nil
	}

	switch args.RequestType {
	case "set":
		// Iterate over all memories and set them
		for _, memory := range args.Memories {
			ok, err := c.setMemory(memory.Memory, memory.Type, memory.Detail)
			if err != nil {
				return fmt.Sprintf(`%v`, err), err
			}
			if !ok {
				return "Failed to set a memory", nil
			}
		}
		return "Memories set successfully", nil

	case "get":
		// Note: This assumes that for 'get', you'll retrieve memories based on the first item in the memories slice. Adjust as needed.
		memoryResponse, err := c.getMemory(args.Memories[0], args.Num_relevant)
		if err != nil {
			return fmt.Sprintf(`%v`, err), err
		}
		return fmt.Sprintf(`%v`, memoryResponse), nil
	case "hydrate":
		prompt, err := c.HydrateUserMemories()
		if err != nil {
			return fmt.Sprintf(`%v`, err), err
		}
		return prompt, nil
	default:
		return "unknown request type check out Example for how to use the memory plug", nil
	}
}

func (c Memory) getEmbeddingsFromOpenAI(data string) openai.Embedding {
	embeddings, err := c.openaiClient.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{data},
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		fmt.Println(err)
	}

	return embeddings.Data[0]
}

func (c Memory) setMemory(newMemory, memoryType, memoryDetail string) (bool, error) {
	// Step 1: Combine the three fields into a single string
	combinedMemory := memoryType + "|" + memoryDetail + "|" + newMemory

	embeddings := c.getEmbeddingsFromOpenAI(combinedMemory)

	longTermMemory := memory{
		Memory: combinedMemory, // Use combinedMemory here
		Vector: embeddings.Embedding,
	}

	memories := []memory{
		longTermMemory,
	}

	memoryData := make([]string, 0, len(memories))
	vectors := make([][]float32, 0, len(memories))

	for _, memory := range memories {
		memoryData = append(memoryData, memory.Memory)
		vectors = append(vectors, memory.Vector)
	}

	memoryColumn := entity.NewColumnVarChar("memory", memoryData)
	vectorColumn := entity.NewColumnFloatVector("embeddings", 1536, vectors)

	_, err := c.milvusClient.Insert(context.Background(), c.cfg.MalvusCollectionName(), "", memoryColumn, vectorColumn)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (c Memory) getMemory(memory memoryItem, num_relevant int) ([]memoryResult, error) {
	combinedMemory := memory.Type + "|" + memory.Detail + "|" + memory.Memory + ","
	embeddings := c.getEmbeddingsFromOpenAI(combinedMemory)

	ctx := context.Background()
	partitions := []string{}
	expr := ""
	outputFields := []string{"memory"}
	vectors := []entity.Vector{entity.FloatVector(embeddings.Embedding)}
	vectorField := "embeddings"
	metricType := entity.L2
	topK := num_relevant

	searchParam, _ := entity.NewIndexFlatSearchParam()

	options := []milvus.SearchQueryOptionFunc{}

	searchResult, err := c.milvusClient.Search(ctx, c.cfg.MalvusCollectionName(), partitions, expr, outputFields, vectors, vectorField, metricType, topK, searchParam, options...)

	if err != nil {
		return nil, err
	}

	memoryFields := c.getStringSliceFromColumn(searchResult[0].Fields.GetColumn("memory"))

	var allMemories []string
	if len(memoryFields) == 1 {
		allMemories = strings.Split(memoryFields[0], ",")
	} else {
		allMemories = memoryFields
	}

	memoryResults := make([]memoryResult, len(allMemories))

	for idx, memory := range allMemories {
		parts := strings.Split(memory, "|")

		if len(parts) >= 3 {
			memoryResults[idx] = memoryResult{
				Type:   strings.TrimSpace(parts[0]),
				Detail: strings.TrimSpace(parts[1]),
				Memory: strings.TrimSpace(parts[2]),
			}
		}
	}

	return memoryResults, nil
}

func (c Memory) getStringSliceFromColumn(column entity.Column) []string {
	length := column.Len()
	results := make([]string, length)

	for i := 0; i < length; i++ {
		val, err := column.GetAsString(i)
		if err != nil {
			// handle error or continue with a placeholder value
			results[i] = "" // or some placeholder value
		} else {
			results[i] = val
		}
	}

	return results
}

func (c Memory) initMilvusSchema() error {

	//check if schema exists

	if exists, _ := c.milvusClient.HasCollection(context.Background(), c.cfg.MalvusCollectionName()); !exists {
		schema := &entity.Schema{
			CollectionName: c.cfg.MalvusCollectionName(),
			Description:    "Clara's long term memory",
			Fields: []*entity.Field{
				{
					Name:       "memory_id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     true,
				},
				{
					Name:     "memory",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						entity.TypeParamMaxLength: "65535",
					},
				},
				{
					Name:     "embeddings",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						entity.TypeParamDim: "1536",
					},
				},
			},
		}
		err := c.milvusClient.CreateCollection(context.Background(), schema, 1)
		if err != nil {
			return err
		}

		idx, err := entity.NewIndexIvfFlat(entity.L2, 2)

		if err != nil {
			return err
		}

		err = c.milvusClient.CreateIndex(context.Background(), c.cfg.MalvusCollectionName(), "embeddings", idx, false)

		if err != nil {
			return err
		}

	}

	//check to see if the collection is loaded
	loaded, err := c.milvusClient.GetLoadState(context.Background(), c.cfg.MalvusCollectionName(), []string{})

	if err != nil {
		return err
	}

	if loaded == entity.LoadStateNotLoad {
		err = c.milvusClient.LoadCollection(context.Background(), c.cfg.MalvusCollectionName(), false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Memory) HydrateUserMemories() (string, error) {

	var memories = []memoryItem{
		{Type: "Basic Personal Information", Detail: "name"},
		{Type: "Basic Personal Information", Detail: "age"},
		{Type: "Basic Personal Information", Detail: "gender"},
		{Type: "Basic Personal Information", Detail: "location"},

		{Type: "Preferences", Detail: "music_preference"},
		{Type: "Preferences", Detail: "movie_preference"},
		{Type: "Preferences", Detail: "book_preference"},
		{Type: "Preferences", Detail: "food_preference"},

		{Type: "Professional and Educational Background", Detail: "profession"},
		{Type: "Professional and Educational Background", Detail: "education"},
		{Type: "Professional and Educational Background", Detail: "skills"},

		{Type: "Hobbies and Interests", Detail: "hobbies"},
		{Type: "Hobbies and Interests", Detail: "sports"},
		{Type: "Hobbies and Interests", Detail: "travel"},
		{Type: "Hobbies and Interests", Detail: "games"},

		{Type: "Lifestyle and Habits", Detail: "exercise_habit"},
		{Type: "Lifestyle and Habits", Detail: "reading_habit"},
		{Type: "Lifestyle and Habits", Detail: "diet"},
		{Type: "Lifestyle and Habits", Detail: "pets"},

		{Type: "Tech and Media Consumption", Detail: "favorite_apps"},
		{Type: "Tech and Media Consumption", Detail: "device_preference"},
		{Type: "Tech and Media Consumption", Detail: "news_source"},

		{Type: "Social and Personal Relationships", Detail: "family"},
		{Type: "Social and Personal Relationships", Detail: "friends"},
		{Type: "Social and Personal Relationships", Detail: "relationship_status"},

		{Type: "Past Interactions", Detail: "past_questions"},
		{Type: "Past Interactions", Detail: "feedback"},
		{Type: "Past Interactions", Detail: "topics_of_interest"},

		{Type: "Moods and Feelings", Detail: "current_mood"},
		{Type: "Moods and Feelings", Detail: "life_events"},
		{Type: "Moods and Feelings", Detail: "challenges"},

		{Type: "Custom User Data", Detail: "custom_data"},
	}

	uniqueMemories := make(map[string]bool)
	var memoryList []string

	prompt := "You are an AI assistant called Clara, you have a long term memory, here are some memories about the user you can use: "

	for _, m := range memories {
		// Get each memory from the vector database based on user ID and memory type
		results, err := c.getMemory(m, 5)
		if err != nil {
			return "", err
		}

		for _, res := range results {
			// Remove the trailing comma from each memory if it exists
			cleanMemory := strings.TrimSuffix(res.Memory, ",")
			if _, exists := uniqueMemories[cleanMemory]; !exists {
				uniqueMemories[cleanMemory] = true
				memoryList = append(memoryList, cleanMemory)
			}
		}
	}

	// Convert unique memories into a CSV list
	memoriesCSV := strings.Join(memoryList, ", ")

	prompt += memoriesCSV

	return prompt, nil
}
