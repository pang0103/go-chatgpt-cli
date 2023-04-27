/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Delta456/box-cli-maker/v2"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/pang0103/go-chatgpt-cli/cmd/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// chatCmd represents the chat command
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start a new conversation",
	Long:  `Start a new conversation`,
	Run: func(cmd *cobra.Command, args []string) {
		startNewConversation()
	},
}

var messages []openai.ChatCompletionMessage

func startNewConversation() {
	Box := box.New(box.Config{Px: 1, Py: 1, Type: "Double", Color: "Green", TitlePos: "Top"})
	Box.Println("Model", config.Conf.Model)

	quit := false

	for !quit {

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
	fmt.Printf("Conversation finished")
}

func saveConversation(messages []openai.ChatCompletionMessage) {
	fmt.Println("Saving conversation...")

	// Create a new conversation object with a UUID, the given messages, and a topic
	conversation := Conversation{
		Id:       uuid.New().String(),
		Messages: messages,
		Topic:    generateTopic(messages),
	}

	// Read the existing conversations from the JSON file
	conversations := loadConversation()

	// Append the new conversation to the existing conversations
	conversations = append(conversations, conversation)

	// Encode the conversations to a JSON-encoded byte slice
	jsonData, err := json.MarshalIndent(conversations, "", "  ")
	if err != nil {
		panic(err)
	}

	// Write the byte slice to the JSON file
	err = ioutil.WriteFile("cache.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Conversation saved.")
}

func loadConversation() []Conversation {
	// Read the existing conversations from the JSON file
	fileData, err := os.ReadFile("cache.json")
	if err != nil {
		fmt.Println("No conversation found.")
		return []Conversation{}
	}

	// Decode the JSON-encoded data into a slice of Conversation objects
	var conversations []Conversation
	err = json.Unmarshal(fileData, &conversations)
	if err != nil {
		panic(err)
	}

	return conversations
}

func getConversation(id string) Conversation {
	conversations := loadConversation()
	for _, conversation := range conversations {
		if conversation.Id == id {
			return conversation
		}
	}
	panic("Conversation not found")
}

func generateTopic(messages []openai.ChatCompletionMessage) string {
	prompt := "Write an extremely concise subtitle for this conversation with no more than a few words. All words should be capitalized. Exclude punctuation."

	client := openai.NewClient(config.Conf.ApiKey)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: config.Conf.Model,
			Messages: append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			}),
		},
	)

	if err != nil {
		panic(err)
	}

	return resp.Choices[0].Message.Content
}

func init() {
	ChatCmd.AddCommand(historyCmd)
	ChatCmd.AddCommand(resumeCmd)
}
