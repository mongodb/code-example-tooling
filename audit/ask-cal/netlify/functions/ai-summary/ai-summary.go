package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
	"github.com/tmc/langchaingo/prompts"
	"log"
	"os"
	"strings"
)

type RequestBody struct {
	Code    string `json:"code"`
	PageURL string `json:"pageUrl"`
}

func getAiSummaryFromHuggingFace(code string, pageUrl string) (string, error) {
	ctx := context.Background()
	template := prompts.NewPromptTemplate(
		`Find the code example below on the given webpage. In a few sentences, summarize the section of
			the webpage that contains the given code example. Provide information that gives developers useful context
			surrounding the code example - don't just describe the code example.
			Code: {{.code}}
			Page URL: {{.pageUrl}}`,
		[]string{"question", "context"},
	)
	prompt, err := template.Format(map[string]any{
		"pageUrl": pageUrl,
		"code":    code,
	})
	if err != nil {
		return "", fmt.Errorf("failed to format prompt: %w", err)
	}
	opts := llms.CallOptions{
		Model:       "mistralai/Mistral-7B-Instruct-v0.3",
		MaxTokens:   150,
		Temperature: 0.1,
	}

	llm, err := huggingface.New(
		huggingface.WithToken(os.Getenv("HUGGINGFACEHUB_API_TOKEN")),
		huggingface.WithModel("mistralai/Mistral-7B-Instruct-v0.3"),
	)
	if err != nil {
		return "", fmt.Errorf("failed to initialize a Hugging Face LLM (is your API Token set?): %v", err)
	}
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt, llms.WithOptions(opts))
	if err != nil {
		return "", fmt.Errorf("failed to generate a response from the prompt (is Hugging Face reachable?): %w", err)
	}
	response := extractLogicalResponse(completion)
	return response, nil
}

func extractLogicalResponse(input string) string {
	// Define the marker "Page URL:"
	marker := "Page URL:"

	// Locate the marker
	idx := strings.Index(input, marker)
	if idx == -1 {
		// If marker is not found, return an empty string
		return ""
	}

	// Extract the text after the marker
	textAfterMarker := input[idx+len(marker):]

	// Find the first line break after the URL
	lines := strings.Split(strings.TrimSpace(textAfterMarker), "\n")
	if len(lines) > 1 {
		// Return everything after the first line (logical response)
		return strings.Join(lines[1:], "\n")
	}

	return strings.TrimSpace(textAfterMarker) // Fallback: only the text after the marker
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var requestPayload RequestBody
	err := json.Unmarshal([]byte(request.Body), &requestPayload)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode:      422,
			Headers:         map[string]string{"Content-Type": "text/plain"},
			Body:            "Invalid input",
			IsBase64Encoded: false,
		}, err
	}
	summary, err := getAiSummaryFromHuggingFace(requestPayload.Code, requestPayload.PageURL)
	if err != nil {
		log.Printf("Error generating AI summary: %v", err)
		return &events.APIGatewayProxyResponse{
			StatusCode:      500,
			Headers:         map[string]string{"Content-Type": "text/plain"},
			Body:            "Failed to generate summary. Check if Hugging Face is reachable and your API token is set.",
			IsBase64Encoded: false,
		}, err
	}
	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "text/plain"},
		Body:            summary,
		IsBase64Encoded: false,
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
