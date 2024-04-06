// 定义配置包
package config

// 导入必要的包
import (
	"fmt" // 用于格式化输出
	"os"  // 用于操作系统相关的操作，如文件操作

	"github.com/sirupsen/logrus" // 引入logrus包，用于日志记录
)

// 定义Milvus数据库配置的结构体
type MalvusCfg struct {
	apiEndpoint    string // Milvus服务器的API终端地址
	collectionName string // Milvus中用于存储数据的集合名称
}

// 定义主配置结构体
type Cfg struct {
	openAiAPIKey string // OpenAI API的密钥

	pluginsPath string // 插件存放的路径
	logName     string // 日志文件的名称

	malvusCfg MalvusCfg // Milvus数据库的配置

	AppLogger *logrus.Logger // 应用程序使用的日志记录器
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
		openAiAPIKey: "sk-2uEpTNt8ESEvkQUG30De33002fEc411e841a29FaD4Db5cE5", // OpenAI API的密钥
		pluginsPath:  "./plugins",                                           // 插件路径
		malvusCfg:    malvusCfg,                                             // 设置Milvus配置
	}

	// 初始化日志记录器
	err := cfg.InitLogger()
	if err != nil {
		// 如果初始化日志记录器失败，则打印错误信息并退出程序
		fmt.Println("Error initializing logger: ", err)
		os.Exit(1)
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

// InitLogger方法初始化日志记录器
func (c *Cfg) InitLogger() error {
	if c.logName == "" {
		// 如果没有指定日志文件名称，则使用默认名称
		c.logName = "clara.log"
	}
	// 打开或创建日志文件
	file, err := os.OpenFile(c.logName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// 如果打开或创建文件失败，则返回错误
		return fmt.Errorf("failed to open log file %s for output: %s", c.logName, err)
	}
	c.AppLogger = logrus.New() // 创建新的logrus日志记录器实例
	c.AppLogger.Out = file     // 设置日志输出到文件
	return nil
}
