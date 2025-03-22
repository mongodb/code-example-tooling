package updates

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

func CopyDBForTesting(client *mongo.Client, ctx context.Context) {
	sourceDb := client.Database("code_metrics")
	targetDb := client.Database("backup_code_metrics")
	// List all collections in the source database
	collectionNames, err := sourceDb.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Fatalf("Error listing collections: %v", err)
	}
	// Iterate over each collection
	for _, collName := range collectionNames {
		sourceColl := sourceDb.Collection(collName)
		targetColl := targetDb.Collection(collName)
		// Fetch all documents from the source collection
		cursor, err := sourceColl.Find(ctx, bson.D{})
		if err != nil {
			log.Fatalf("Error finding documents in collection %s: %v", collName, err)
		}
		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err := cursor.Close(ctx)
			if err != nil {
				log.Fatalf("Error closing cursor: %v", err)
			}
		}(cursor, ctx)
		var documents []interface{}
		for cursor.Next(ctx) {
			var doc bson.M
			if err = cursor.Decode(&doc); err != nil {
				log.Fatalf("Error decoding document in collection %s: %v", collName, err)
			}
			documents = append(documents, doc)
		}
		if len(documents) > 0 {
			_, err = targetColl.InsertMany(ctx, documents)
			if err != nil {
				log.Fatalf("Error inserting documents into collection %s: %v", collName, err)
			}
			log.Printf("Copied %d documents to collection %s", len(documents), collName)
		}
	}
	log.Println("All collections copied successfully")
}
