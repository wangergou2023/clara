// chatui.go
package chatui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-wordwrap"
)

type disableInputMsg struct{}
type enableInputMsg struct{}

type message struct {
	sender string
	text   string
	style  lipgloss.Style
}

type model struct {
	viewport      viewport.Model
	width         int
	height        int
	messages      []message
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	err           error
	newMessages   chan message
	inputDisabled bool
	UserMessages  chan string // Exported channel for user messages

}

type ChatUI struct {
	model     model
	prog      *tea.Program
	messages  chan message
	shouldRun bool
}

// The rest of your functions/methods (such as initialModel, formatMessages, etc.)
// should be unchanged and added here, but without the main() function.

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	//set width to the width of the viewport

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Clara!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:     ta,
		messages:     []message{},
		viewport:     vp,
		senderStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:          nil,
		newMessages:  make(chan message, 100), // buffer of 100 messages; adjust as needed
		UserMessages: make(chan string),       // Initialize the channel
	}
}

func NewChatUI() (*ChatUI, error) {
	m := initialModel()
	p := tea.NewProgram(m)

	chat := &ChatUI{
		model:    m,
		prog:     p,
		messages: make(chan message),
	}

	go chat.listenForMessages()

	return chat, nil
}

func (c *ChatUI) Run() error {
	_, err := c.prog.Run()
	close(c.messages)
	return err
}

func (c *ChatUI) listenForMessages() {
	for msg := range c.messages {
		c.model.messages = append(c.model.messages, msg)
		c.model.viewport.SetContent(formatMessages(c.model.messages, c.model.viewport.Width))
		c.model.viewport.GotoBottom() // Add this line to ensure the viewport is at the bottom
	}
}

func (c *ChatUI) AddMessage(sender, text string) {
	msgStyle := lipgloss.NewStyle() // default style
	if sender == "User" {
		msgStyle = c.model.senderStyle
	} else if sender == "Clara" {
		msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	} else if sender == "SYSTEM" { // Add this case
		msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // 9 for red in lipgloss' default palette
	}

	newMessage := message{
		sender: sender,
		text:   text,
		style:  msgStyle,
	}
	c.model.newMessages <- newMessage
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
