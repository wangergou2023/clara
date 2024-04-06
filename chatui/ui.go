// chatui.go
package chatui

// 导入所需的包
import (
	"fmt"     // 用于格式化字符串
	"strings" // 提供用于操作字符串的实用函数

	"github.com/charmbracelet/bubbles/textarea" // 提供文本区域组件
	"github.com/charmbracelet/bubbles/viewport" // 提供可滚动视图组件
	tea "github.com/charmbracelet/bubbletea"    // 引入bubbletea，用于创建CLI应用
	"github.com/charmbracelet/lipgloss"         // 用于美化CLI界面的样式库
	"github.com/mitchellh/go-wordwrap"          // 用于文本自动换行
)

// 定义启用和禁用输入的消息结构体
type disableInputMsg struct{}
type enableInputMsg struct{}

// message 结构体用于存储消息信息
type message struct {
	sender string         // 消息发送者
	text   string         // 消息文本
	style  lipgloss.Style // 消息样式
}

// model 结构体用于存储界面状态
type model struct {
	viewport      viewport.Model // 视图组件，用于显示消息
	width         int            // 界面宽度
	height        int            // 界面高度
	messages      []message      // 消息列表
	textarea      textarea.Model // 文本输入组件
	senderStyle   lipgloss.Style // 发送者样式
	err           error          // 存储可能出现的错误
	newMessages   chan message   // 用于接收新消息的通道
	inputDisabled bool           // 输入是否被禁用
	UserMessages  chan string    // 用于外部获取用户输入消息的通道
}

// ChatUI 结构体用于管理整个聊天UI
type ChatUI struct {
	model     model        // UI模型
	prog      *tea.Program // Bubble Tea程序实例
	messages  chan message // 消息通道
	shouldRun bool         // 程序是否应该继续运行
}

func initialModel() model {
	ta := textarea.New()                 // 创建一个新的文本区域
	ta.Placeholder = "Send a message..." // 设置占位符
	ta.Focus()                           // 聚焦文本区域

	ta.Prompt = "┃ "   // 设置提示符
	ta.CharLimit = 280 // 设置字符限制

	ta.SetWidth(30) // 设置文本区域宽度
	ta.SetHeight(3) // 设置文本区域高度

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle() // 设置聚焦样式

	ta.ShowLineNumbers = false // 不显示行号

	vp := viewport.New(30, 5) // 创建一个新的视图组件
	vp.SetContent(`Welcome to Clara!
Type a message and press Enter to send.`) // 设置初始内容

	ta.KeyMap.InsertNewline.SetEnabled(false) // 禁止插入新行

	return model{
		textarea:     ta,
		messages:     []message{},
		viewport:     vp,
		senderStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:          nil,
		newMessages:  make(chan message, 100), // 创建消息通道，缓冲区大小100
		UserMessages: make(chan string),       // 创建用户消息通道
	}
}

func NewChatUI() (*ChatUI, error) {
	m := initialModel()    // 初始化模型
	p := tea.NewProgram(m) // 创建Bubble Tea程序

	chat := &ChatUI{
		model:    m,
		prog:     p,
		messages: make(chan message),
	}

	go chat.listenForMessages() // 启动消息监听协程

	return chat, nil
}

func (c *ChatUI) Run() error {
	_, err := c.prog.Run() // 运行Bubble Tea程序
	close(c.messages)      // 程序结束时关闭消息通道
	return err
}

func (c *ChatUI) listenForMessages() {
	for msg := range c.messages { // 从消息通道接收消息
		c.model.messages = append(c.model.messages, msg)                                      // 将消息添加到模型中
		c.model.viewport.SetContent(formatMessages(c.model.messages, c.model.viewport.Width)) // 更新视图内容
		c.model.viewport.GotoBottom()                                                         // 滚动到底部
	}
}

func (c *ChatUI) AddMessage(sender, text string) {
	msgStyle := lipgloss.NewStyle() // 默认样式
	if sender == "User" {
		msgStyle = c.model.senderStyle
	} else if sender == "Clara" {
		msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	} else if sender == "SYSTEM" {
		msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}

	newMessage := message{
		sender: sender,
		text:   text,
		style:  msgStyle,
	}
	c.model.newMessages <- newMessage // 将新消息发送到通道
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.width
		m.viewport.Height = m.height - 6
		m.textarea.SetWidth(m.width)
		return m, nil
	case disableInputMsg:
		m.inputDisabled = true
		m.textarea.Placeholder = "Waiting for Clara..."
		return m, nil
	case enableInputMsg:
		m.inputDisabled = false
		m.textarea.Placeholder = "Send a message..."
		return m, nil
	case tea.KeyMsg:

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:

			if m.inputDisabled {
				return m, nil
			}

			// User's message
			newMessage := message{
				sender: "User",
				text:   m.textarea.Value(),
				style:  m.senderStyle,
			}
			m.messages = append(m.messages, newMessage)
			m.UserMessages <- m.textarea.Value()

			m.viewport.SetContent(formatMessages(m.messages, m.viewport.Width))

			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

		return m, tea.Batch(tiCmd, vpCmd)

	default:
		// Check for new messages
		select {
		case newMsg := <-m.newMessages:
			m.messages = append(m.messages, newMsg)
			m.viewport.SetContent(formatMessages(m.messages, m.viewport.Width))
			m.viewport.GotoBottom()
		default:
			// Do nothing if no new messages
		}
		return m, tea.Batch(tiCmd, vpCmd)
	}
}

func formatMessages(messages []message, viewportWidth int) string {
	halfWidth := viewportWidth / 2
	paddingBetween := 10 // this can be adjusted based on your preference for the gap between the two messages

	var formattedMessages []string
	for _, msg := range messages {
		// Render and wrap the text
		wrappedText := wrapText(msg.style.Render(msg.text), uint(halfWidth-paddingBetween))
		wrappedLines := strings.Split(wrappedText, "\n")

		// Determine the padding for the message depending on the sender
		if msg.sender == "User" { // If the sender is not "AI", format it to be left aligned
			formattedMessages = append(formattedMessages, fmt.Sprintf("%s: %s", msg.sender, wrappedText))
		} else { // If the sender is "AI", format it to be right aligned
			// Find the longest line for padding calculation
			longestLine := 0
			for _, line := range wrappedLines {
				if len(line) > longestLine {
					longestLine = len(line)
				}
			}

			// Calculate the required spaces to pad the message to the right side
			paddingSize := viewportWidth - longestLine - len(msg.sender) - 2 - paddingBetween
			if paddingSize < 0 {
				paddingSize = 0
			}
			leftPadding := strings.Repeat(" ", paddingSize)

			for idx, line := range wrappedLines {
				if idx == 0 {
					formattedMessages = append(formattedMessages, fmt.Sprintf("%s%s: %s", leftPadding, msg.sender, line))
				} else {
					formattedMessages = append(formattedMessages, fmt.Sprintf("%s%s", leftPadding, line))
				}
			}
		}
	}

	return strings.Join(formattedMessages, "\n")
}

func wrapText(text string, width uint) string {
	return wordwrap.WrapString(text, width)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func (c *ChatUI) DisableInput() {
	c.prog.Send(disableInputMsg{})
}

func (c *ChatUI) EnableInput() {
	c.prog.Send(enableInputMsg{})
}

func (c *ChatUI) GetUserMessagesChannel() chan string {
	return c.model.UserMessages
}

func (c *ChatUI) ClearHistory() {
	c.model.messages = []message{}
	c.model.viewport.SetContent(formatMessages(c.model.messages, c.model.viewport.Width))
}
