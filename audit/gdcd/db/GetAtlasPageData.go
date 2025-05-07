package db

import (
	"common"
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAtlasPageData(collectionName string, docId string) *common.DocsPage {
	uri := os.Getenv("MONGODB_URI")
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	var dbName = os.Getenv("DB_NAME")
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
	filter := bson.D{{Key: "_id", Value: docId}}
	// Create a DocsPage object to hold the result
	var result common.DocsPage

	const maxRetries = 3
	const retryDelay = 2 * time.Second
	retryableErrorPrefix := "connection() error occurred during connection handshake"

	for attempts := 0; attempts < maxRetries; attempts++ {
		err := collection.FindOne(ctx, filter).Decode(&result)

		if err == nil {
			// Successful fetch
			return &result
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}

		if isRetryableError(err, retryableErrorPrefix) {
			log.Printf("Attempt %d: transient error occurred, retrying: %v", attempts+1, err)
			time.Sleep(retryDelay)
			continue
		} else {
			log.Printf("Error: can't find a matching document for page %v, %v\n", docId, err)
			return nil
		}
	}

	log.Printf("Failed after %d attempts to find document for page %v", maxRetries, docId)
	return nil
}

// Helper function to determine if an error is transient and retryable
func isRetryableError(err error, prefix string) bool {
	return err != nil && err.Error()[:len(prefix)] == prefix
}
