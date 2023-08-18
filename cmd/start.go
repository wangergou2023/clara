package cmd

import (
	"github.com/inancgumus/screen"
	chatbot "github.com/jjkirkpatrick/clara/chatBot"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Clara",
	Long:  `Stat Clara`,
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func start() {
	screen.Clear()
	screen.MoveTopLeft()
	chatbot.Start()
}
