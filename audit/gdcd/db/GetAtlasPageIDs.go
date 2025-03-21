package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
)

func GetAtlasPageIDs(collectionName string) []string {
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
	// Define the filter to exclude documents with "_id" equal to "summaries"
	filter := bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}
	// Project only the "_id" field
	projection := bson.D{{Key: "_id", Value: 1}}

	findOptions := options.Find().SetProjection(projection)
	// Query the collection
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	// Slice to store the _id
	var ids []string
	// Iterate over the cursor
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Failed to decode document: %v\n", err)
		}
		if id, ok := result["_id"].(string); ok { // Ensure the _id is a string
			ids = append(ids, id)
		} else {
			log.Println("Found non-string _id, skipping...")
		}
	}
	if err := cursor.Err(); err != nil {
		log.Printf("Failed to cursor: %v/n", err)
	}
	return ids
}
