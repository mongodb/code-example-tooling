package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
	"sort"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("Set your 'DB_NAME' environment variable. ")
	}
	db := client.Database(dbName)

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("Set your 'COLLECTION_NAME' environment variable. ")
	}
	coll := db.Collection(collectionName)

	// Define the substrings to search for - strings potentially related to code examples
	include := []string{"code", "example", "deprecated", "api", "method", "function", "parameter", "doesn't work", "does not work", "broken", "fails", "failed", "error", "outdated", "warning"}
	// Build the $or condition with $regex for each substring
	var regexIncludeConditions bson.A
	for _, substring := range include {
		regexIncludeConditions = append(regexIncludeConditions, bson.D{
			{"comment", bson.D{
				{"$regex", substring},
				{"$options", "i"}, // Add this line to make the regex case-insensitive
			}},
		})
	}

	// Define the substrings to exclude - strings potentially related to broken links
	exclude := []string{"link", "url"}
	var regexExcludeConditions bson.A
	for _, substring := range exclude {
		regexExcludeConditions = append(regexExcludeConditions, bson.D{
			{"comment", bson.D{
				{"$regex", substring},
				{"$options", "i"}, // Add this line to make the regex case-insensitive
			}},
		})
	}

	// Define the aggregation pipeline
	pipeline := mongo.Pipeline{
		{
			{"$match", bson.D{
				{"$and", bson.A{
					bson.D{
						{"comment", bson.D{
							{"$exists", true}, // Ensure the "comment" field exists
						}},
					},
					bson.D{
						// Exclude comments that match any of the specified regex exclusion conditions
						{"$nor", regexExcludeConditions},
					},
					bson.D{
						{"$or", regexIncludeConditions}, // Match any of the specified regex conditions
					},
				}},
			}},
		},
	}
	fmt.Println("Performing aggregations to run report. This may take a moment.")
	// Execute the aggregation pipeline
	cur, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var results []Feedback
	for cur.Next(ctx) {
		var result Feedback
		if err = cur.Decode(&result); err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}
	if err = cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Get the total count of feedback in the collection. Used to create percentages when breaking down code-related counts.
	filter := bson.D{}
	var totalDocumentCount int64
	totalDocumentCount, err = coll.CountDocuments(ctx, filter)
	fmt.Printf("Total current count of feedback in collection: %d\n", totalDocumentCount)
	fmt.Printf("Total count of feedback related to code examples: %d\n", len(results))

	// Sort the results based on DocsProperty
	sort.Slice(results, func(i, j int) bool {
		return results[i].Page.DocsProperty < results[j].Page.DocsProperty
	})
	// Group the results by DocsProperty
	groupedResults := make(map[string][]Feedback)
	for _, result := range results {
		groupedResults[result.Page.DocsProperty] = append(groupedResults[result.Page.DocsProperty], result)
	}
	// Count results for each DocsProperty
	counts := make(map[string]int)
	for docsProperty, feedbacks := range groupedResults {
		counts[docsProperty] = len(feedbacks)
	}
	// Print counts for each DocsProperty
	fmt.Printf("\nCounts:\n")
	for docsProperty, count := range counts {
		fmt.Printf("%s: %d\n", docsProperty, count)
	}

	// Open a file for writing
	file, err := os.Create("report.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	// Write header to CSV
	writer.Write([]string{"EntryNumber", "DocsProperty", "URL", "Comment"})
	entryNumber := 1
	for docsProperty, feedbacks := range groupedResults {
		for _, feedback := range feedbacks {
			// Write each feedback as a row in the CSV
			writer.Write([]string{
				fmt.Sprintf("%d", entryNumber),
				docsProperty,
				feedback.Page.URL,
				feedback.Comment,
			})
			entryNumber++
		}
	}
	// Ensure all data is flushed to the file
	writer.Flush()
	// Check for any error during the write operation
	if err = writer.Error(); err != nil {
		fmt.Println("Error writing to CSV:", err)
	}
}
