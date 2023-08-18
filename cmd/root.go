package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "cobra-cli",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().String("openaiAPIKey", "", "OpenAI API Key")
	rootCmd.PersistentFlags().String("openaiModel", "davinci", "OpenAI Model")
	rootCmd.PersistentFlags().Bool("supervisedMode", true, "Supervised Mode")
	rootCmd.PersistentFlags().Bool("DebugMode", false, "Debug Mode")
	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("openaiModel", "davinci")
	viper.SetDefault("supervisedMode", false)
	viper.SetDefault("DebugMode", false)
	viper.SetDefault("versionNumber", "0.0.1")

	initConfig()

	if viper.Get("openaiAPIKey") == nil {
		var apiKey = []*survey.Question{
			{
				Name: "openaiAPIKey",
				Prompt: &survey.Input{
					Message: "No API Key detected, please enter your OpenAI API Key:",
				},
				Validate: survey.Required,
			},
		}

		ansmap := make(map[string]interface{})
		err := survey.Ask(apiKey, &ansmap, survey.WithShowCursor(true))

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if ansmap["openaiAPIKey"] != "" {
			fmt.Println("Setting API Key to: " + ansmap["openaiAPIKey"].(string))
			viper.Set("openaiAPIKey", ansmap["openaiAPIKey"])
			if err := viper.WriteConfig(); err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}

}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".clara")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, creating...")
			viper.WriteConfigAs(".clara")
		} else {
			fmt.Println("Config file found but another error was produced")
			// Config file was found but another error was produced
		}
	}

}
