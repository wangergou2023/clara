package util

import "fmt"

func PromptUserForInput(prompt string) string {
	fmt.Println(prompt)
	var input string
	fmt.Scanln(&input)

	

	return input
}
