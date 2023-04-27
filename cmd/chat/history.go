/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package chat

import (
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

type Conversation struct {
	Id       string                         `json:"id"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Topic    string                         `json:"topic"`
}

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "list conversations history",
	Long:  `list conversations history`,
	Run: func(cmd *cobra.Command, args []string) {
		printHistory()
	},
}

func printHistory() {
	fmt.Println("Loading conversation...")

	// Load the conversations from the file
	conversations := loadConversation()

	table := simpletable.New()

	// Header
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "#"},
			{Align: simpletable.AlignCenter, Text: "ID"},
			{Align: simpletable.AlignCenter, Text: "Topic"},
			//{Align: simpletable.AlignCenter, Text: "Messages"},
		},
	}

	// Content
	for i, conversation := range conversations {
		r := []*simpletable.Cell{
			{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", i+1)},
			{Align: simpletable.AlignRight, Text: conversation.Id},
			{Align: simpletable.AlignRight, Text: conversation.Topic},
			//{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", len(conversation.Messages))},
		}
		table.Body.Cells = append(table.Body.Cells, r)
	}

	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())

}

func init() {

}
