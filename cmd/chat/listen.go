/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Delta456/box-cli-maker/v2"
	"github.com/kkdai/youtube/v2"
	"github.com/pang0103/go-chatgpt-cli/cmd/config"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	url string
)

// listenCmd represents the listen command
var ListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen to a YouTube video",
	Long:  `The listen command will download the audio from a YouTube video and feed to the OpenAI API. IT will give a summary of the video. Note that the model OpenAI API does not support visual content yet. The summary is based on the audio transcript.`,
	Run: func(cmd *cobra.Command, args []string) {
		file := downloadVideo(url)
		audioUrl, err := uploadAudio(file)
		if err != nil {
			log.Fatal(err)
		}

		transcript, err := getTranscript(audioUrl)
		if err != nil {
			log.Fatal(err)
		}
		videoSummary := summarizeByTranscript(transcript)
		fmt.Println()
		fmt.Println("Video summary: ")
		fmt.Println(videoSummary)

		startNewConversation(IntroductionBox{
			title:   "ChatGPT",
			message: "Start a conversation based on the video content",
		})
	},
}

func getTranscript(audioURL string) (string, error) {
	endpoint := "https://api.assemblyai.com/v2/transcript"

	// Create the request body as a JSON object.
	requestBody, err := json.Marshal(map[string]string{
		"audio_url": audioURL,
	})
	if err != nil {
		return "", fmt.Errorf("error creating request body: %v", err)
	}

	// Create the HTTP request.
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", config.Conf.AssemblyaiKey)

	// Send the HTTP request.
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the JSON response.
	var respBody struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	// Poll for the transcript.
	pollEndpoint := fmt.Sprintf("%s/%s", endpoint, respBody.ID)
	for {
		req, err := http.NewRequest("GET", pollEndpoint, nil)
		if err != nil {
			return "", fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("authorization", config.Conf.AssemblyaiKey)

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("error sending request: %v", err)
		}

		// Check if the transcript is ready.
		var pollResp struct {
			Status     string `json:"status"`
			Transcript string `json:"text"`
			Error      string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&pollResp)
		if err != nil {
			return "", fmt.Errorf("error decoding response: %v", err)
		}
		fmt.Printf("\rStatus: %s", pollResp.Status)

		resp.Body.Close()

		if pollResp.Status == "completed" {
			//fmt.Println("Transcript:", pollResp.Transcript)
			return pollResp.Transcript, nil
		} else if pollResp.Status == "error" {
			return "", fmt.Errorf("error getting transcript: %v", pollResp.Error)
		}

		// Wait before polling again.
		time.Sleep(1 * time.Second)
	}
}

func uploadAudio(filePath string) (string, error) {
	// Open the file.
	fmt.Println("Submitting audio file to AssemblyAI...")
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a new HTTP request with the file.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("error copying file: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}

	// Create the HTTP request.
	req, err := http.NewRequest("POST", "https://api.assemblyai.com/v2/upload", body)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("authorization", config.Conf.AssemblyaiKey)

	// Send the HTTP request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the JSON response.
	var respBody struct {
		UploadURL string `json:"upload_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return respBody.UploadURL, nil
}

func downloadVideo(youtubeUrl string) string {
	// Replace the videoID with the ID of the YouTube video you want to download.
	videoID := youtubeUrl

	// Create a new YouTube client.
	client := youtube.Client{}

	// Get the video information.
	video, err := client.GetVideo(videoID)
	if err != nil {
		log.Fatalf("Error getting video: %v", err)
	}

	Box := box.New(box.Config{Px: 1, Py: 1, Type: "Double", Color: "Green", TitlePos: "Top"})
	Box.Print("Video Infos", fmt.Sprintf("Title: %s\nAuthor: %s\nDuration: %v", video.Title, video.Author, video.Duration))
	fmt.Println("Downloading video...")

	// Get the highest quality video format.
	videoFormat := video.Formats.WithAudioChannels()[0]

	// Get the video stream.
	videoStream, _, err := client.GetStream(video, &videoFormat)
	if err != nil {
		log.Fatalf("Error getting video stream: %v", err)
	}

	// Create the local file where the video will be saved.
	videoFile, err := os.Create(fmt.Sprintf("%s.mp4", strings.ReplaceAll(video.Title, "/", "_")))
	if err != nil {
		log.Fatalf("Error creating local file: %v", err)
	}
	defer videoFile.Close()

	// Download the video stream and save it to the local file.
	_, err = io.Copy(videoFile, videoStream)
	if err != nil {
		log.Fatalf("Error downloading and saving video: %v", err)
	}

	// Use FFmpeg to extract the audio from the video.
	audioFile := fmt.Sprintf("%s.mp3", strings.ReplaceAll(video.Title, "/", "_"))
	fmt.Println("Extracting audio...")
	cmd := exec.Command("ffmpeg", "-i", fmt.Sprintf("%s.mp4", strings.ReplaceAll(video.Title, "/", "_")), "-vn", "-acodec", "libmp3lame", audioFile)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error extracting audio: %v", err)
	}
	fmt.Println("Audio extracted!")
	return audioFile
}

func init() {
	ListenCmd.Flags().StringVarP(&url, "url", "u", "", "URL to listen to")
}
