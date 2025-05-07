package db

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RemovePageFromAtlas deletes a common.DocsPage from Atlas. We don't need to update the collection summaries document,
// because that will be overwritten with existing page count and code node count at the end of this run.
func RemovePageFromAtlas(collectionName string, pageId string) bool {
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
	db := client.Database(dbName)
	coll := db.Collection(collectionName)
	filter := bson.M{"_id": pageId}
	// Delete the document
	var deleteResult *mongo.DeleteResult
	deleteResult, err = coll.DeleteOne(ctx, filter)
	if err != nil {
		log.Printf("Failed to delete MongoDB document for pageId %s: %v\n", pageId, err)
	}
	if deleteResult.DeletedCount == 1 {
		return true
	} else {
		return false
	}
}
