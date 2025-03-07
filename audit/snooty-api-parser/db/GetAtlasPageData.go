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

func GetAtlasPageData(collectionName string, docId string) *types.DocsPage {
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
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()
	// Define the database and collection
	collection := client.Database(dbName).Collection(collectionName)
	filter := bson.D{{Key: "_id", Value: docId}}
	// Create a DocsPage object to hold the result
	var result types.DocsPage
	// Execute the query
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("Didn't find matching document for page %v - need to make a new one\n", docId)
		} else {
			log.Printf("Error: can't find a matching document for page %v, %v\n", docId, err)
		}
	}
	return &result
}
