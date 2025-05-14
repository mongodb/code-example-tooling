package main

import (
	"common"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type QueryResult struct {
	CodeExamples []ReshapedCodeNode `json:"code_examples"`
	AnalyticsID  string             `json:"analytics_id"`
}

type ReshapedCodeNode struct {
	Code        string `bson:"code" json:"code"`
	Language    string `bson:"language" json:"language"`
	Category    string `bson:"category" json:"category"`
	PageURL     string `bson:"page_url" json:"pageUrl"`
	ProjectName string `bson:"project_name" json:"projectName"`
}

type CodeExampleResult struct {
	Code            string `json:"code"`
	Language        string `json:"language"`
	Category        string `json:"category"`
	PageURL         string `json:"pageUrl"`
	ProjectName     string `json:"projectName"`
	PageTitle       string `json:"pageTitle"`
	PageDescription string `json:"pageDescription"`
}

type ResponseBody struct {
	QueryId      string              `json:"queryId"`
	CodeExamples []CodeExampleResult `json:"codeExamples"`
}

//func getSearchResultsFromAtlas(queryString string, languageFacet string, categoryFacet string, docsSet string) []ReshapedCodeNode {
//	ctx := context.Background()
//	uri := os.Getenv("CODE_SNIPPETS_CONNECTION_STRING")
//	client, err := mongo.Connect(options.Client().
//		ApplyURI(uri))
//	if err != nil {
//		log.Fatalf("Failed to connect to MongoDB: %v", err)
//	}
//	defer func(client *mongo.Client, ctx context.Context) {
//		err := client.Disconnect(ctx)
//		if err != nil {
//			fmt.Printf("Failed to disconnect from mongodb: %v", err)
//		}
//	}(client, ctx)
//
//	collection := client.Database("ask_cal").Collection("consolidated_examples")
//
//	// Define the aggregation pipeline with search and multiple facet filters
//	pipeline := mongo.Pipeline{
//		// `$search` with text search and multiple filters - MUST BE THE FIRST STAGE
//		{
//			{"$search", bson.D{
//				{"index", "ask_cal"},
//				{"compound", bson.D{
//					{"should", bson.A{
//						bson.D{
//							{"text", bson.D{
//								{"query", queryString},
//								{"path", "nodes.code"},
//							}},
//						},
//					}},
//					{"filter", bson.A{
//						bson.D{
//							{"equals", bson.D{
//								{"path", "languages_facet"},
//								{"value", languageFacet},
//							}},
//						},
//						bson.D{
//							{"equals", bson.D{
//								{"path", "categories_facet"},
//								{"value", categoryFacet},
//							}},
//						},
//						bson.D{
//							{"equals", bson.D{
//								{"path", "project_name"},
//								{"value", docsSet},
//							}},
//						},
//					}},
//				}},
//			}},
//		},
//		// Reshape the data using a `$project` stage
//		{
//			{"$project", bson.D{
//				{"code", "$nodes.code"},
//				{"language", "$nodes.language"},
//				{"category", "$nodes.category"},
//				{"page_url", "$page_url"},
//				{"project_name", "$project_name"},
//			}},
//		},
//	}
//
//	// Execute the aggregation pipeline
//	cursor, err := collection.Aggregate(ctx, pipeline)
//
//	// Iterate through the results
//	var results []ReshapedCodeNode
//	if err != nil {
//		log.Fatalf("Failed to execute aggregation pipeline: %v", err)
//	}
//	defer cursor.Close(ctx)
//
//	if err := cursor.All(ctx, &results); err != nil {
//		log.Fatalf("Failed to decode aggregation results: %v", err)
//	}
//
//	return results
//}

func getSearchResultsFromAtlas(query common.QueryRequestBody, ctx context.Context) QueryResult {
	queryStartTime := time.Now()
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

	collection := client.Database("ask_cal").Collection("consolidated_examples")
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
	queryCompletionTime := time.Now()
	queryTimeElapsedInNanoseconds := queryCompletionTime.Sub(queryStartTime)
	queryTimeInSeconds := queryTimeElapsedInNanoseconds.Seconds()
	analyticsObjectId := createAnalyticsReport(query, ctx, queryTimeInSeconds)

	return QueryResult{
		CodeExamples: results,
		AnalyticsID:  analyticsObjectId,
	}
}

func getPageNameAndDescription(pageURL string) (string, string) {
	// Step 1: Fetch the HTML content of the webpage
	resp, err := http.Get(pageURL)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ""
	}

	// Step 2: Parse the HTML content with goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", ""
	}

	// Step 3: Extract the `<title>` tag
	title := doc.Find("title").Text()
	// Substring to find and trim
	substring := "arrow-"

	// Trim the string
	trimmedTitle := trimStartingFromSubstring(title, substring)

	// Step 4: Extract the meta description from `<meta name="description">`
	description := ""
	metaDescription := doc.Find(`meta[name="description"]`)
	if metaDescription.Length() > 0 {
		description, _ = metaDescription.Attr("content")
	}

	return trimmedTitle, description
}

func trimStartingFromSubstring(input string, substring string) string {
	// Find the index where the substring appears
	index := strings.Index(input, substring)
	if index == -1 {
		// If the substring is not found, return the original string
		return input
	}

	// Return the string trimmed up to the index of the substring
	return input[:index]
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
	queryResults := getSearchResultsFromAtlas(requestPayload, ctx)
	var codeExamples []CodeExampleResult
	for index, result := range queryResults.CodeExamples {
		var title string
		var description string
		// Only get the title and description for the first 10 results
		if index < 10 {
			title, description = getPageNameAndDescription(result.PageURL)
		}
		completeResult := CodeExampleResult{
			Code:            result.Code,
			Language:        result.Language,
			Category:        result.Category,
			PageURL:         result.PageURL,
			ProjectName:     result.ProjectName,
			PageTitle:       title,
			PageDescription: description,
		}
		codeExamples = append(codeExamples, completeResult)
	}

	responseBody := ResponseBody{
		QueryId:      queryResults.AnalyticsID,
		CodeExamples: codeExamples,
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
