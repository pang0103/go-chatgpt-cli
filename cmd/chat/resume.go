/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package chat

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/Delta456/box-cli-maker/v2"
	"github.com/fatih/color"
	"github.com/pang0103/go-chatgpt-cli/cmd/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var (
	conversationId string
)

// resumeCmd represents the resume command
var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "resume a conversation",
	Long:  `Resume a conversation`,
	Run: func(cmd *cobra.Command, args []string) {
		resumeConversation()
	},
}

func resumeConversation() {
	// Load the conversations from the file
	messages = getConversation(conversationId).Messages

	Box := box.New(box.Config{Px: 1, Py: 1, Type: "Double", Color: "Green", TitlePos: "Top"})
	Box.Println("Success", "Resumed conversation "+conversationId)

	for {
		color.New(color.FgGreen).Printf("You: ")

		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "exit" {
			fmt.Println("Exiting conversation...")
			return
		}

		if message == "save" {
			saveConversation(messages)
			return
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message,
		})

		client := openai.NewClient(config.Conf.ApiKey)
		stream, err := client.CreateChatCompletionStream(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    config.Conf.Model,
				Messages: messages,
				Stream:   true,
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}

		defer stream.Close()

		res := ""
		fmt.Println()
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println()
				break
			}
			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				return
			}
			content := response.Choices[0].Delta.Content
			fmt.Printf(content)
			res += content
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: res,
		})
		fmt.Println()
	}

}

func init() {
	resumeCmd.Flags().StringVarP(&conversationId, "id", "i", "", "conversation id")
	if err := resumeCmd.MarkFlagRequired("id"); err != nil {
		panic(err)
	}
}
