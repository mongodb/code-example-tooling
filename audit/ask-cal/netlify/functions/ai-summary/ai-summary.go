package main

import (
	"context"
	"encoding/json"
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

func getAiSummaryFromHuggingFace(code string, pageUrl string) string {
	ctx := context.Background()
	template := prompts.NewPromptTemplate(
		`Find the code example below on the given webpage. In a few sentences, summarize the section of
			the webpage that contains the given code example. Provide information that gives developers useful context
			surrounding the code example - don't just describe the code example.
			Page URL: {{.pageUrl}}
			Code: {{.code}}`,
		[]string{"question", "context"},
	)
	prompt, err := template.Format(map[string]any{
		"pageUrl": pageUrl,
		"code":    code,
	})
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
		log.Fatalf("failed to initialize a Hugging Face LLM: %v", err)
	}
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt, llms.WithOptions(opts))
	if err != nil {
		log.Fatalf("failed to generate a response from the prompt: %v", err)
	}
	response := strings.Split(completion, "\n\n")
	// For this particular LLM implementation, the first string is the prompt it was given, and subsequent strings are related paragraphs.
	// Omit the prompt and provide the subsequent strings as one string for the response.
	var responseLines strings.Builder
	for index, line := range response {
		if index == 0 {
			continue
		} else {
			responseLines.WriteString(line + "\n\n")
		}
	}
	return responseLines.String()
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
	summary := getAiSummaryFromHuggingFace(requestPayload.Code, requestPayload.PageURL)
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
