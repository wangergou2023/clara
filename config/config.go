package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Cfg struct {
	openAiAPIKey string

	pluginsPath string
	logName     string

	malvusCfg MalvusCfg

	AppLogger *logrus.Logger
}

type MalvusCfg struct {
	apiEndpoint    string
	collectionName string
}

func New() Cfg {

	malvusCfg := MalvusCfg{
		apiEndpoint:    "http://llxspace.store:19530", //"https://in03-4082eb33e7b209a.api.gcp-us-west1.zillizcloud.com/v1/",
		collectionName: "CGPTMemory",
	}

	cfg := Cfg{
		openAiAPIKey: "sk-2uEpTNt8ESEvkQUG30De33002fEc411e841a29FaD4Db5cE5",
		pluginsPath:  "./plugins",
		malvusCfg:    malvusCfg,
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

func (c Cfg) PluginsPath() string {
	return c.pluginsPath
}

func (c Cfg) SetOpenAiAPIKey(openAiAPIKey string) Cfg {
	c.openAiAPIKey = openAiAPIKey
	return c
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
