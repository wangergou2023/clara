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
		apiEndpoint:    "localhost:19530",
		collectionName: "CGPTMemory",
	}

	return Cfg{
		openAiAPIKey:    "",
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
