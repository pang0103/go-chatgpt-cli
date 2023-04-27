/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	ApiKey string `json:"api_key"`
	Model  string `json:"model"`
}

var (
	configFilePath = filepath.Join(os.Getenv("HOME"), ".go-chatgpt-cli.json")
	Conf           = Config{
		ApiKey: "",
		Model:  openai.GPT3Dot5Turbo,
	}
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure your API key and model etc",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run config")
		err := setConfig()
		if err != nil {
			panic(err)
		}
	},
}

// setConfig saves the current Conf values to the configuration file.
func setConfig() error {
	configData, err := json.MarshalIndent(Conf, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configFilePath, configData, 0644)
}

// loadConfig loads the configuration file, if it exists, and sets the Conf variable.
func loadConfig() error {
	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configData, &Conf)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	ConfigCmd.Flags().StringVarP(&Conf.ApiKey, "api_key", "k", "", "API key")
	ConfigCmd.Flags().StringVarP(&Conf.Model, "model", "m", "", "model")

	err := loadConfig()
	if err != nil {
		panic(err)
	}

	if Conf.Model == "" {
		Conf.Model = "gpt-3.5-turbo"
	}

	if Conf.ApiKey == "" {
		Conf.ApiKey = os.Getenv("OPENAI_API_KEY")
	}
}
