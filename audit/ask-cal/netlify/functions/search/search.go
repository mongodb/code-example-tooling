package main

import (
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
)

type RequestBody struct {
	QueryString   string `json:"queryString"`
	LanguageFacet string `json:"languageFacet"`
	CategoryFacet string `json:"categoryFacet"`
	DocsSet       string `json:"docsSet"`
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

func getSearchResultsFromAtlas(queryString string, languageFacet string, categoryFacet string, docsSet string) []ReshapedCodeNode {
	ctx := context.Background()
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
					{"$regex", queryString},
					{"$options", "i"}, // Case-insensitive string matching
				}},
			}},
		},
	}

	// Add the default string matching pipeline to the main pipeline
	pipeline = append(pipeline, stringMatchingPipeline...)

	// Conditionally add a `$match` stage for `languageFacet`
	if languageFacet != "" {
		languageFacetStage := bson.D{
			{"$match", bson.D{
				{"nodes.language", languageFacet},
			}},
		}
		pipeline = append(pipeline, languageFacetStage)
	}

	// Conditionally add a `$match` stage for `categoryFacet`
	if categoryFacet != "" {
		categoryFacetStage := bson.D{
			{"$match", bson.D{
				{"nodes.category", categoryFacet},
			}},
		}
		pipeline = append(pipeline, categoryFacetStage)
	}

	// Conditionally add a `$match` stage for `docsSet`
	if categoryFacet != "" {
		categoryFacetStage := bson.D{
			{"$match", bson.D{
				{"project_name", docsSet},
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
	return results
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

	queryResults := getSearchResultsFromAtlas(body.QueryString, body.LanguageFacet, body.CategoryFacet, body.DocsSet)
	var codeExamples []CodeExampleResult
	for _, result := range queryResults {
		title, description := getPageNameAndDescription(result.PageURL)
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
	codeExamplesAsJson, _ := json.Marshal(codeExamples)

	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json"},
		Body:            string(codeExamplesAsJson),
		IsBase64Encoded: false,
	}, nil

}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
