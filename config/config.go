package config

type Cfg struct {
	openAiAPIKey string
	openAiModel  string

	supervisedModel bool
	debugMode       bool

	pluginsPath string

	malvusCfg MalvusCfg
}

type MalvusCfg struct {
	apiKey         string
	apiEndpoint    string
	collectionName string
}

func New() Cfg {

	malvusCfg := MalvusCfg{
		apiKey:         "fece80a7dc4ef3a041f09cb9db27119d6d12a8c65ca182ad7cc0ed595e981f2beed2e20c0a328062eb7c52838f6b0e8fed4a4c55",
		apiEndpoint:    "localhost:19530", //"https://in03-4082eb33e7b209a.api.gcp-us-west1.zillizcloud.com/v1/",
		collectionName: "CGPTMemory",
	}

	return Cfg{
		openAiAPIKey:    "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",
		openAiModel:     "gpt-4",
		supervisedModel: false,
		debugMode:       false,
		pluginsPath:     "./plugins",
		malvusCfg:       malvusCfg,
	}
}

func (c Cfg) OpenAiAPIKey() string {
	return c.openAiAPIKey
}

func (c Cfg) OpenAiModel() string {
	return c.openAiModel
}

func (c Cfg) SupervisedModel() bool {
	return c.supervisedModel
}

func (c Cfg) DebugMode() bool {
	return c.debugMode
}

func (c Cfg) SetDebugMode(debugMode bool) Cfg {
	c.debugMode = debugMode
	return c
}

func (c Cfg) PluginsPath() string {
	return c.pluginsPath
}

func (c Cfg) SetSupervisedModel(supervisedModel bool) Cfg {
	c.supervisedModel = supervisedModel
	return c
}

func (c Cfg) SetOpenAiAPIKey(openAiAPIKey string) Cfg {
	c.openAiAPIKey = openAiAPIKey
	return c
}

func (c Cfg) SetOpenAiModel(openAiModel string) Cfg {
	c.openAiModel = openAiModel
	return c
}

func (c Cfg) SetPluginsPath(pluginsPath string) Cfg {
	c.pluginsPath = pluginsPath
	return c
}

func (c Cfg) SetMalvusApiKey(apiKey string) Cfg {
	c.malvusCfg.apiKey = apiKey
	return c
}

func (c Cfg) MalvusApiKey() string {
	return c.malvusCfg.apiKey
}

func (c Cfg) MalvusApiEndpoint() string {
	return c.malvusCfg.apiEndpoint
}

func (c Cfg) SetMalvusApiEndpoint(apiEndpoint string) Cfg {
	c.malvusCfg.apiEndpoint = apiEndpoint
	return c
}

func (c Cfg) MalvusCollectionName() string {
	return c.malvusCfg.collectionName
}

func (c Cfg) SetMalvusCollectionName(collectionName string) Cfg {
	c.malvusCfg.collectionName = collectionName
	return c

}
