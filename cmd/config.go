package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/inancgumus/screen"
	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Clara configuration",
	Long:  `Update the Clara configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		updateConfig()
	},
}

func updateConfig() {
	fmt.Println("Clara Configuration,", aurora.Magenta("Aurora"))

	questions := []*survey.Question{
		{
			Name: "OpenaiAPIKey",
			Prompt: &survey.Input{
				Message: aurora.Sprintf(aurora.Magenta("OpenAI API Key (%s):"), viper.GetString("openaiAPIKey")),
			},
			Validate: survey.Required,
		},
		{
			Name: "OpenAIModel",
			Prompt: &survey.Select{
				Message: aurora.Sprintf(aurora.Magenta("OpenAI Model (%s):"), viper.GetString("openaiModel")),
				Options: []string{"gpt-4", "gpt-3.5-turbo"},
				Default: "gpt-4",
			},
			Validate: survey.Required,
		},
		{
			Name: "supervisedMode",
			Prompt: &survey.Confirm{
				Message: aurora.Sprintf(aurora.Magenta("Supervised Mode (%t):"), viper.GetBool("supervisedMode")),
				Default: true,
			},
		},
		{
			Name: "debugMode",
			Prompt: &survey.Confirm{
				Message: aurora.Sprintf(aurora.Magenta("Debug Mode (%t):"), viper.GetBool("DebugMode")),
				Default: false,
			},
		},
	}

	answers := struct {
		OpenaiAPIKey   string `survey:"openaiAPIKey"` // or you can tag fields to match a specific name
		OpenAIModel    string `survey:"openaiModel"`
		SupervisedMode bool   `survey:"supervisedMode"`
		DebugMode      bool   `survey:"debugMode"`
	}{}

	err := survey.Ask(questions, &answers, survey.WithShowCursor(true))

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	screen.Clear()
	screen.MoveTopLeft()

	fmt.Println(aurora.Magenta("OpenAI API Key:"), answers.OpenaiAPIKey)
	fmt.Println(aurora.Magenta("OpenAI Model:"), answers.OpenAIModel)
	fmt.Println(aurora.Magenta("Supervised Mode:"), answers.SupervisedMode)
	fmt.Println(aurora.Magenta("Debug Mode:"), answers.DebugMode)

	//get users confermintion
	var confirm = []*survey.Question{
		{
			Name: "confirm",
			Prompt: &survey.Confirm{
				Message: aurora.Sprintf(aurora.Magenta("Are you sure you want to update the configuration?")),
				Default: false,
			},
		},
	}

	confirmation := false
	err = survey.Ask(confirm, &confirmation, survey.WithShowCursor(true))

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if confirmation == false {
		fmt.Println("Aborting...")
		return
	}

	//update config
	viper.Set("openaiAPIKey", answers.OpenaiAPIKey)
	viper.Set("openaiModel", answers.OpenAIModel)
	viper.Set("supervisedMode", answers.SupervisedMode)
	viper.Set("debugMode", answers.DebugMode)

	err = viper.WriteConfig()
	if err != nil {
		fmt.Println("Error writing config:", err)
	}

}
