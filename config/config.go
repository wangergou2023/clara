package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Cfg struct {
	openAiAPIKey string
	openAiModel  string

	supervisedModel bool
	debugMode       bool

	pluginsPath string
	logName     string

	malvusCfg MalvusCfg

	AppLogger *logrus.Logger
}

type MalvusCfg struct {
	apiKey         string
	apiEndpoint    string
	collectionName string
}

func New() Cfg {

	malvusCfg := MalvusCfg{
		apiEndpoint:    "http://llxspace.store:19530", //"https://in03-4082eb33e7b209a.api.gcp-us-west1.zillizcloud.com/v1/",
		collectionName: "CGPTMemory",
	}

	cfg := Cfg{
		openAiAPIKey:    "sk-2uEpTNt8ESEvkQUG30De33002fEc411e841a29FaD4Db5cE5",
		openAiModel:     "gpt-4",
		supervisedModel: false,
		debugMode:       false,
		pluginsPath:     "./plugins",
		malvusCfg:       malvusCfg,
	}

	err := cfg.InitLogger()
	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		os.Exit(1)
	}

	return cfg
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

func (c *Cfg) InitLogger() error {
	if c.logName == "" {
		c.logName = "clara.log"
	}
	file, err := os.OpenFile(c.logName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file %s for output: %s", c.logName, err)
	}
	c.AppLogger = logrus.New()
	c.AppLogger.Out = file
	return nil
}
