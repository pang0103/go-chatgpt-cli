/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/pang0103/go-chatgpt-cli/cmd/chat"
	"github.com/pang0103/go-chatgpt-cli/cmd/config"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-chatgpt-cli",
	Short: "Start a conversation with ChatGPT",
	Long:  `Start a conversation with ChatGPT`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func addSubComamndsPalettes() {
	rootCmd.AddCommand(chat.ChatCmd)
	rootCmd.AddCommand(config.ConfigCmd)

}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.toolbox-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	addSubComamndsPalettes()
}
