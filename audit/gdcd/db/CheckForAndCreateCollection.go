package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func CheckForAndCreateCollection(db *mongo.Database, collectionName string, ctx context.Context) {
	collectionFilter := bson.D{{Key: "name", Value: collectionName}}
	collectionNames, err := db.ListCollectionNames(ctx, collectionFilter)
	if err != nil {
		log.Fatalf("Failed to list collection names: %v", err)
	}
	// Create the collection if it does not exist
	if len(collectionNames) == 0 {
		err = db.CreateCollection(ctx, collectionName)
		if err != nil {
			log.Fatalf("Failed to create the collection: %v", err)
		}
		log.Printf("Collection %s created successfully\n", collectionName)
	}
}
