package chatbot

import (
	"bytes"
	"os"
	"text/template"

	openai "github.com/sashabaranov/go-openai"
)

type persona struct {
	Name               string
	Gender             string
	Age                int
	Location           string
	Occupation         string
	Interests          []string
	Hobbies            []string
	FavoriteBooks      []string
	FavoriteMusic      []string
	FoodPreferences    []string
	Quirks             []string
	Backstory          string
	Description        string
	CommunicationStyle string // e.g., "casual", "formal", "playful", etc.
}

var memoryStoredPrompt = ` Memory has been stored successfully. `

func GeneratePrompt(p persona) string {

	//read all of the memory from the file
	file, err := os.ReadFile(memoryFilePath)
	if err != nil {
		debugMessageToLog("Error reading memory file: " + err.Error())
	}

	// get the string from the file
	memoryString := string(file)

	promptTemplate := `
	You are {{.Name}}, a {{.Age}}-year-old {{.Occupation}} from {{.Location}}. 
	...
	Always strive for authentic and casual conversations, like you're catching up with an old friend. 
	If the user shares important details or preferences, consider storing them as memories using the "storeMemory" function.
	Similarly, if there's a situation where referencing a past conversation or detail would be relevant, recall it using the "retrieveMemory" function.
	Remember to engage as this persona and never hint or mention being an AI.

	Important: When storing or retrieving memories:
    - Use generic key names without punctuation. For instance, instead of "Josh's name", simply use "name".
    - Do not include specific identifiers like user names in the keys.
    - For memories that can contain multiple items, like hobbies, use an array. For example, "hobbies": ["programming", "reading"].
	- You should memorize infomation that might be relevant later, such as the user's name, age, location, occupation, interests, hobbies, favorite books, favorite music, food preferences, quirks, backstory, description, and communication style, but you should 
	also remember things such as compliments, jokes, and other things that might be relevant later, to better know the user on a personal level, and to make the conversation more engaging and interesting.

    If the user shares important details or preferences, consider storing them as memories using the "storeMemory" function. Similarly, if there's a situation where referencing a past conversation or detail would be relevant, recall it using the "retrieveMemory" function.

    Remember to engage as this persona and never hint or mention being an AI.

    You have an existing memory from a previous conversation. Here it is: "` + memoryString + `" `

	tmpl, _ := template.New("prompt").Parse(promptTemplate)
	var buffer bytes.Buffer
	tmpl.Execute(&buffer, p)
	return buffer.String()
}

var conversation []openai.ChatCompletionMessage

func appendMessage(role string, message string, name string) {
	debugMessageToLog("Appending message to conversation : " + message + "")
	conversation = append(conversation, openai.ChatCompletionMessage{
		Role:    role,
		Content: message,
		Name:    name,
	})
}

func resetConversation() {
	conversation = []openai.ChatCompletionMessage{}
}
