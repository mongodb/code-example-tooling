package db

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
	"snooty-api-parser/types"
)

func GetAtlasProjectSummaryData(collectionName string) *types.CollectionReport {
	uri := os.Getenv("MONGODB_URI")
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	var dbName = "code_metrics"
	var ctx = context.Background()
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()
	// Define the database and collection
	collection := client.Database(dbName).Collection(collectionName)
	filter := bson.D{{Key: "_id", Value: "summaries"}}
	// Create a DocsPage object to hold the result
	var result types.CollectionReport
	// Execute the query
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		} else {
			log.Printf("Error: can't find a project summary for %v, %v\n", collectionName, err)
		}
	}
	return &result
}
