package main

import (
	"context"
	"encoding/json"
	"fmt"

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
	ID     int64
	memory string
	Vector []float32
}

type memoryResult struct {
	Memory string
	Score  float32
}

type inputDefinition struct {
	RequestType  string
	Memory       string
	Num_relevant int
}

func (c *Memory) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	c.cfg = cfg
	c.openaiClient = openaiClient

	ctx := context.Background()
	//config := milvus.Config{
	//	Address:       cfg.MalvusApiEndpoint(), // Cluster endpoint.
	//	Identifier:    "myconnection",          // Identifier for this connection.
	//	EnableTLSAuth: true,                    // Enable TLS Auth for transport security.
	//	APIKey:        cfg.MalvusApiKey(),      // API key.
	//}

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
		Description: "store and retrieve memories from long term memory. use requestType set to add a memory to the database, use requestType get to retrieve the most relevant memories.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"requestType": {
					Type:        jsonschema.String,
					Description: "the type of request to make either set or get, set will add a memory to the database, get will return the most relevant memories",
				},
				"memory": {
					Type:        jsonschema.String,
					Description: "the memory to add or get, example: I like to eat pizza",
				},
				"num_relevant": {
					Type:        jsonschema.Integer,
					Description: "the number of relevant memories to return, example: 5",
				},
			},
			Required: []string{"requestType", "memory"},
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

	if args.Memory == "" {
		return fmt.Sprintf(`%v`, "memory is required but was empty"), nil
	}

	switch args.RequestType {
	case "set":
		ok, err := c.setMemory(args.Memory)
		if err != nil {
			return fmt.Sprintf(`%v`, err), err
		}
		if ok {
			return fmt.Sprintf(`%v`, "Memory set successfully"), nil
		}

	case "get":
		memoryResponse, err := c.getMemory(args.Memory, args.Num_relevant)
		if err != nil {
			return fmt.Sprintf(`%v`, err), err
		}
		return fmt.Sprintf(`%v}`, memoryResponse), nil
	default:
		return "unknown request type check out Example for how to use the memory plug", nil

	}

	return "", nil
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

func (c Memory) setMemory(newMemory string) (bool, error) {
	embeddings := c.getEmbeddingsFromOpenAI(newMemory)

	longTermMemory := memory{
		ID:     1,
		memory: newMemory,
		Vector: embeddings.Embedding,
	}

	memories := []memory{
		longTermMemory,
	}

	memoryData := make([]string, 0, len(memories))
	vectors := make([][]float32, 0, len(memories))

	for _, memory := range memories {
		memoryData = append(memoryData, memory.memory)
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

func (c Memory) getMemory(memoryToGet string, num_relevant int) (string, error) {
	embeddings := c.getEmbeddingsFromOpenAI(memoryToGet)

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
		return fmt.Sprint("unable to search milvus"), err
	}

	memoryResults := make([]memoryResult, 0, len(searchResult)*topK)

	for _, sr := range searchResult {
		memoryFields := c.getStringSliceFromColumn(sr.Fields.GetColumn("memory"))

		for i := 0; i < len(sr.Scores); i++ {
			memoryResults = append(memoryResults, memoryResult{
				Memory: memoryFields[i],
				Score:  sr.Scores[i],
			})
		}
	}

	// marshal memoryResults to json
	jsonMemoryResults, err := json.Marshal(memoryResults)
	if err != nil {
		return "", err
	}

	return string(jsonMemoryResults), nil

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
			Description:    "Test book search",
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
