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
	"time"
)

type QueryResult struct {
	CodeExamples []ReshapedCodeNode `json:"code_examples"`
	AnalyticsID  string             `json:"analytics_id"`
}

type ReshapedCodeNode struct {
	Code            string `bson:"code" json:"code"`
	Language        string `bson:"language" json:"language"`
	Category        string `bson:"category" json:"category"`
	PageURL         string `bson:"page_url" json:"pageUrl"`
	ProjectName     string `bson:"project_name" json:"projectName"`
	PageTitle       string `bson:"page_title" json:"pageTitle"`
	PageDescription string `bson:"page_description" json:"pageDescription"`
}

type ResponseBody struct {
	QueryId      string             `json:"queryId"`
	CodeExamples []ReshapedCodeNode `json:"codeExamples"`
}

// This functionality uses Atlas Search, which gives us full-text matching and additional search options
func performAtlasSearchQuery(query common.QueryRequestBody, ctx context.Context) []ReshapedCodeNode {
	uri := os.Getenv("CODE_SNIPPETS_CONNECTION_STRING")
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			fmt.Printf("Failed to disconnect from mongodb: %v", err)
		}
	}(client, ctx)

	collection := client.Database("ask_cal").Collection("consolidated_examples_v2")
	// Initialize the pipeline
	var pipeline mongo.Pipeline

	// Build the $search stage of the query - this stage MUST come first
	searchStage := mongo.Pipeline{
		bson.D{
			{"$search", bson.D{
				{"index", "ask_cal_v2"},
				{"compound", bson.D{
					{"must", bson.A{
						bson.D{
							{"text", bson.D{
								{"query", query.QueryString},
								{"path", bson.A{
									"page_title",
									"page_description",
									"nodes.code",
									"page_url",
									"sub_product",
								}},
							}},
						},
					}},
				}},
			}},
		},
	}

	pipeline = append(pipeline, searchStage...)

	// Add `$sort` stage to order by relevance score
	sortStage := bson.D{{"$sort", bson.D{{"_score", -1}}}}
	pipeline = append(pipeline, sortStage)

	// Unwind the nodes array to filter on individual node values
	unwindStage := mongo.Pipeline{
		{{"$unwind", bson.D{{"path", "$nodes"}}}}, // Unwind the `nodes` array
	}
	pipeline = append(pipeline, unwindStage...)

	// Conditionally add a `$match` stage for `languageFacet`
	if query.LanguageFacet != "" {
		languageFacetStage := bson.D{
			{"$match", bson.D{
				{"nodes.language", query.LanguageFacet},
			}},
		}
		pipeline = append(pipeline, languageFacetStage)
	}

	// Conditionally add a `$match` stage for `categoryFacet`
	if query.CategoryFacet != "" {
		categoryFacetStage := bson.D{
			{"$match", bson.D{
				{"nodes.category", query.CategoryFacet},
			}},
		}
		pipeline = append(pipeline, categoryFacetStage)
	}

	// Conditionally add a `$match` stage for `docsSet`
	if query.DocsSet != "" {
		docsSetFacetStage := bson.D{
			{"$match", bson.D{
				{"project_name", query.DocsSet},
			}},
		}
		pipeline = append(pipeline, docsSetFacetStage)
	}

	projectStage := bson.D{
		{"$project", bson.D{
			{"code", "$nodes.code"},
			{"language", "$nodes.language"},
			{"category", "$nodes.category"},
			{"page_url", "$page_url"},
			{"project_name", "$project_name"},
			{"page_title", "$page_title"},
			{"page_description", "$page_description"},
		}},
	}
	pipeline = append(pipeline, projectStage)

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation: %v", err)
	}
	defer cursor.Close(ctx)

	// Iterate through the results
	var results []ReshapedCodeNode
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &results); err != nil {
		log.Fatalf("Failed to decode aggregation results: %v", err)
	}
	return results
}

// This functionality uses an aggregation pipeline with exact substring match only in the code example field
// Try first to see if we have some exact matches; otherwise, fall back to Atlas Search to expand the options
// With enough tweaking, we could probably get this functionality *in* Atlas Search, but it's taking too much time for a quick PoC
func performExactMatchQuery(query common.QueryRequestBody, ctx context.Context) []ReshapedCodeNode {
	uri := os.Getenv("CODE_SNIPPETS_CONNECTION_STRING")
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			fmt.Printf("Failed to disconnect from mongodb: %v", err)
		}
	}(client, ctx)

	collection := client.Database("ask_cal").Collection("consolidated_examples_v2")
	// Initialize the pipeline
	var pipeline mongo.Pipeline

	// Default string matching pipeline (always present)
	stringMatchingPipeline := mongo.Pipeline{
		{{"$unwind", bson.D{{"path", "$nodes"}}}}, // Unwind the `nodes` array
		{
			{"$match", bson.D{
				{"nodes.code", bson.D{
					{"$exists", true},
					{"$type", "string"}, // Ensure `code` is a string
				}},
			}},
		},
		{
			{"$match", bson.D{
				{"nodes.code", bson.D{
					{"$regex", query.QueryString},
					{"$options", "i"}, // Case-insensitive string matching
				}},
			}},
		},
	}

	// Add the default string matching pipeline to the main pipeline
	pipeline = append(pipeline, stringMatchingPipeline...)

	// Conditionally add a `$match` stage for `languageFacet`
	if query.LanguageFacet != "" {
		languageFacetStage := bson.D{
			{"$match", bson.D{
				{"nodes.language", query.LanguageFacet},
			}},
		}
		pipeline = append(pipeline, languageFacetStage)
	}

	// Conditionally add a `$match` stage for `categoryFacet`
	if query.CategoryFacet != "" {
		categoryFacetStage := bson.D{
			{"$match", bson.D{
				{"nodes.category", query.CategoryFacet},
			}},
		}
		pipeline = append(pipeline, categoryFacetStage)
	}

	// Conditionally add a `$match` stage for `docsSet`
	if query.DocsSet != "" {
		categoryFacetStage := bson.D{
			{"$match", bson.D{
				{"project_name", query.DocsSet},
			}},
		}
		pipeline = append(pipeline, categoryFacetStage)
	}

	projectStage := bson.D{
		{"$project", bson.D{
			{"code", "$nodes.code"},
			{"language", "$nodes.language"},
			{"category", "$nodes.category"},
			{"page_url", "$page_url"},
			{"project_name", "$project_name"},
			{"page_title", "$page_title"},
			{"page_description", "$page_description"},
		}},
	}
	pipeline = append(pipeline, projectStage)

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation: %v", err)
	}
	defer cursor.Close(ctx)

	// Iterate through the results
	var results []ReshapedCodeNode
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &results); err != nil {
		log.Fatalf("Failed to decode aggregation results: %v", err)
	}
	return results
}

func createAnalyticsReport(query common.QueryRequestBody, ctx context.Context, queryTimeElapsed float64) string {
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
	reportId := bson.NewObjectID()
	feedback := common.AnalyticsReport{
		ID:                     reportId,
		Query:                  query,
		CreatedDate:            time.Now(),
		QueryDurationInSeconds: queryTimeElapsed,
		ResultsFeedback:        nil,
		SummaryFeedback:        nil,
	}
	result, err := collection.InsertOne(ctx, feedback)
	if err != nil {
		fmt.Printf("Failed to insert the document: %v\n", err)
	}
	if result.InsertedID != nil && result.InsertedID != reportId {
		fmt.Printf("The inserted document ID %s does not match the document ID %s\n", result.InsertedID, reportId)
	}
	return reportId.Hex()
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var requestPayload common.QueryRequestBody
	err := json.Unmarshal([]byte(request.Body), &requestPayload)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode:      422,
			Headers:         map[string]string{"Content-Type": "text/plain"},
			Body:            "Invalid input",
			IsBase64Encoded: false,
		}, err
	}

	ctx := context.Background()

	queryStartTime := time.Now()

	// The exact match query only searches for an exact text match within the nodes.code fields. It may not return enough results.
	// If we get fewer than, say, 5 results, flesh out the list with some extended Atlas Search results
	queryResults := performExactMatchQuery(requestPayload, ctx)
	if len(queryResults) < 5 {
		expandedQueryResults := performAtlasSearchQuery(requestPayload, ctx)
		queryResults = append(queryResults, expandedQueryResults...)
	}
	queryCompletionTime := time.Now()
	queryTimeElapsedInNanoseconds := queryCompletionTime.Sub(queryStartTime)
	queryTimeInSeconds := queryTimeElapsedInNanoseconds.Seconds()
	analyticsObjectId := createAnalyticsReport(requestPayload, ctx, queryTimeInSeconds)

	// We don't want a large number of responses, so limit the number of responses we're sending to the front end
	var responseBody ResponseBody
	resultsLimit := 30
	if len(queryResults) < resultsLimit {
		responseBody = ResponseBody{
			QueryId:      analyticsObjectId,
			CodeExamples: queryResults,
		}
	} else {
		limitedResultsSet := queryResults[:resultsLimit]
		responseBody = ResponseBody{
			QueryId:      analyticsObjectId,
			CodeExamples: limitedResultsSet,
		}
	}
	responseAsJson, _ := json.Marshal(responseBody)

	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json"},
		Body:            string(responseAsJson),
		IsBase64Encoded: false,
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
