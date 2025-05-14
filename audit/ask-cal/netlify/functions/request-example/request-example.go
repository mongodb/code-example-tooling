package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"os"
)

type RequestBody struct {
	Description string `json:"Description"`
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var body RequestBody
	err := json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode:      422,
			Headers:         map[string]string{"Content-Type": "text/plain"},
			Body:            "Invalid input",
			IsBase64Encoded: false,
		}, err
	}

	url := os.Getenv("CODE_EXAMPLE_REQUEST_WEB_HOOK")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(request.Body)))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}
	// Set request headers
	req.Header.Set("Content-Type", "application/json")

	// Use the default HTTP client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}
	defer resp.Body.Close() // Ensure the response body is closed

	if resp.StatusCode == http.StatusOK {
		return &events.APIGatewayProxyResponse{
			StatusCode:      200,
			Headers:         map[string]string{"Content-Type": "application/json"},
			Body:            "{\"success\": true\"}",
			IsBase64Encoded: false,
		}, nil
	} else {
		return &events.APIGatewayProxyResponse{
			StatusCode:      resp.StatusCode,
			Headers:         map[string]string{"Content-Type": "text/plain"},
			Body:            "There was a problem requesting this code example; please contact the Dev Docs team directly",
			IsBase64Encoded: false,
		}, nil
	}
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
