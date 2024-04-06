package assistant

// 导入需要的包
import (
	"errors" // 用于创建错误

	"github.com/manifoldco/promptui" // 用于创建和处理命令行界面的提示
)

// GetUserMessage 函数接受一个可选消息作为参数，返回用户输入的字符串
func GetUserMessage(optionalMessage string) string {
	// 如果没有提供可选消息，则使用默认消息 "Enter message"
	if optionalMessage == "" {
		optionalMessage = "Enter message"
	}

	// 定义一个验证函数，确保用户输入不为空
	validate := func(input string) error {
		// 如果输入字符串长度为0，返回一个错误
		if len(input) == 0 {
			return errors.New("please enter a message")
		}
		// 如果输入有效，返回nil表示无错误
		return nil
	}

	// 创建一个新的Prompt对象，设置标签为可选消息，以及上面定义的验证函数
	prompt := promptui.Prompt{
		Label:    optionalMessage, // 设置提示标签
		Validate: validate,        // 设置验证函数
	}

	// 运行Prompt，等待用户输入，并接收返回的结果和错误
	result, err := prompt.Run()

	// 如果运行Prompt时发生错误，返回空字符串
	if err != nil {
		return ""
	}

	// 如果成功获取输入，返回输入的结果
	return result
}
