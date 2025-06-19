package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

//go:embed prompts
var promptsFS embed.FS

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a message in the OpenAI conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response structure from OpenAI API
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a choice in the OpenAI response
type Choice struct {
	Message Message `json:"message"`
}

// Usage represents the token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIClient represents an OpenAI API client
type OpenAIClient struct {
	APIKey string
	Client *http.Client
}

// NewOpenAIClient creates a new OpenAI API client
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

// GetMovieSuggestions sends a system prompt and movie list to OpenAI for suggestions
func (o *OpenAIClient) GetMovieSuggestions(systemPrompt, movieList string) (string, error) {
	request := OpenAIRequest{
		Model: "gpt-4o-mini", // You can change this to other models like "gpt-4" or "gpt-3.5-turbo"
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: movieList,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var response OpenAIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// GetPrompt reads a prompt file from the embedded filesystem
func GetPrompt(filename string) (string, error) {
	promptBytes, err := promptsFS.ReadFile("prompts/" + filename)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %v", filename, err)
	}
	return string(promptBytes), nil
}

// GetSuggestionsForPlexMovies is a convenience function that takes a list of movies and gets AI suggestions
func GetSuggestionsForPlexMovies(movieList []string) (string, error) {
	// Get OpenAI API key from environment variables
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return "", fmt.Errorf("missing OPENAI_API_KEY environment variable. Please add it to your .env file")
	}

	// Read the system prompt from the embedded file
	systemPrompt, err := GetPrompt("movie_recommendations.txt")
	if err != nil {
		return "", fmt.Errorf("error reading prompt: %v", err)
	}

	client := NewOpenAIClient(openAIKey)

	suggestions, err := client.GetMovieSuggestions(systemPrompt, strings.Join(movieList, ","))
	if err != nil {
		fmt.Printf("Error getting suggestions: %v\n", err)
		return "", fmt.Errorf("error getting suggestions: %v", err)
	}

	return suggestions, nil
}
