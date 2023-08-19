package assistant

import (
	"errors"

	"github.com/manifoldco/promptui"
)

func GetUserMessage(optionalMessage string) string {
	if optionalMessage == "" {
		optionalMessage = "Enter message"
	}

	validate := func(input string) error {
		// check string is not empty
		if len(input) == 0 {
			return errors.New("please enter a message")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    optionalMessage,
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return ""
	}

	return result
}
