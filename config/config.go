// 定义配置包
package config

// 导入必要的包

// 用于格式化输出
// 用于操作系统相关的操作，如文件操作

// 定义Milvus数据库配置的结构体
type MalvusCfg struct {
	apiEndpoint    string // Milvus服务器的API终端地址
	collectionName string // Milvus中用于存储数据的集合名称
}

// 定义主配置结构体
type Cfg struct {
	openAiAPIKey string // OpenAI API的密钥

	openWeatherMapAPIKey string // OpenWeatherMap API的密钥

	pluginsPath string // 插件存放的路径
	logName     string // 日志文件的名称

	malvusCfg MalvusCfg // Milvus数据库的配置
}

// New函数用于创建并初始化Cfg配置实例
func New() Cfg {
	// 初始化Milvus配置
	malvusCfg := MalvusCfg{
		apiEndpoint:    "http://llxspace.store:19530", // Milvus API终端地址
		collectionName: "CGPTMemory",                  // Milvus集合名称
	}

	// 初始化主配置
	cfg := Cfg{
		openAiAPIKey:         "sk-2uEpTNt8ESEvkQUG30De33002fEc411e841a29FaD4Db5cE5", // OpenAI API的密钥
		openWeatherMapAPIKey: "787947f021c60a672678b3a9a20b2d4b",                    // OpenWeatherMap API的密钥
		pluginsPath:          "./plugins",                                           // 插件路径
		malvusCfg:            malvusCfg,                                             // 设置Milvus配置
	}

	return cfg // 返回配置实例
}

// OpenAiAPIKey方法返回OpenAI API的密钥
func (c Cfg) OpenAiAPIKey() string {
	return c.openAiAPIKey
}

// PluginsPath方法返回插件存放的路径
func (c Cfg) PluginsPath() string {
	return c.pluginsPath
}

// SetOpenAiAPIKey方法设置OpenAI API的密钥
func (c Cfg) SetOpenAiAPIKey(openAiAPIKey string) Cfg {
	c.openAiAPIKey = openAiAPIKey
	return c
}

func (c Cfg) OpenWeatherMapAPIKey() string {
	return c.openWeatherMapAPIKey
}

// MalvusApiEndpoint方法返回Milvus API终端的地址
func (c Cfg) MalvusApiEndpoint() string {
	return c.malvusCfg.apiEndpoint
}

// SetMalvusApiEndpoint方法设置Milvus API终端的地址
func (c Cfg) SetMalvusApiEndpoint(apiEndpoint string) Cfg {
	c.malvusCfg.apiEndpoint = apiEndpoint
	return c
}

// MalvusCollectionName方法返回Milvus中用于存储数据的集合名称
func (c Cfg) MalvusCollectionName() string {
	return c.malvusCfg.collectionName
}

// SetMalvusCollectionName方法设置Milvus中用于存储数据的集合名称
func (c Cfg) SetMalvusCollectionName(collectionName string) Cfg {
	c.malvusCfg.collectionName = collectionName
	return c
}
