package main

import (
	"common"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
)

type UserFeedback struct {
	ID        string `json:"_id"`
	IsHelpful bool   `json:"isHelpful"`
	Comment   string `json:"comment,omitempty"`
	Category  string `json:"category,omitempty"`
	Type      string `json:"type"` // Value should be either "results" or "summary"
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var requestPayload UserFeedback
	err := json.Unmarshal([]byte(request.Body), &requestPayload)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode:      422,
			Headers:         map[string]string{"Content-Type": "text/plain"},
			Body:            "Invalid input",
			IsBase64Encoded: false,
		}, err
	}

	// Prpare the update
	var update bson.M
	if requestPayload.Type == "results" {
		feedback := common.ResultsFeedback{
			IsHelpful: requestPayload.IsHelpful,
		}
		if requestPayload.Comment != "" {
			feedback.Comment = requestPayload.Comment
		}
		update = bson.M{
			"$set": bson.M{"results_feedback": feedback},
		}
	} else if requestPayload.Type == "summary" {
		feedback := common.SummaryFeedback{
			IsHelpful: requestPayload.IsHelpful,
		}
		if requestPayload.Category != "" {
			feedback.Category = requestPayload.Category
		}
		if requestPayload.Comment != "" {
			feedback.Comment = requestPayload.Comment
		}
		update = bson.M{
			"$set": bson.M{"summary_feedback": feedback},
		}
	}

	objectId, err := bson.ObjectIDFromHex(requestPayload.ID)

	// Write update to the DB
	ctx := context.Background()
	uri := os.Getenv("ANALYTICS_CONNECTION_STRING")
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			fmt.Printf("Failed to disconnect from mongodb: %v\n", err)
		}
	}(client, ctx)

	collection := client.Database("analytics").Collection("v1")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectId}, update)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode:      500,
			Headers:         map[string]string{"Content-Type": "text/plain"},
			Body:            "Failed to update feedback",
			IsBase64Encoded: false,
		}, err
	}

	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json"},
		Body:            fmt.Sprintf("{\"success\": true, \"updated\": \"%s\"}", requestPayload.Type),
		IsBase64Encoded: false,
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
