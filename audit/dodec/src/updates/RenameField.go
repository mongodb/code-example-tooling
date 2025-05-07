package updates

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RenameField changes a field name from oldFieldName to newFieldName for every document in all the collections.
func RenameField(db *mongo.Database, ctx context.Context) {
	// List collection names
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	oldFieldName := "new_field"   // Replace with the actual old field name
	newFieldName := "sub_product" // Replace with the desired new field name
	// Iterate through each collection
	for _, collectionName := range collections {
		collection := db.Collection(collectionName)
		// Build the update statement to rename the field
		update := bson.D{
			{"$rename", bson.D{
				{oldFieldName, newFieldName},
			}},
		}
		// Update documents in the collection
		result, err := collection.UpdateMany(ctx, bson.D{}, update)
		if err != nil {
			log.Printf("Error updating collection %s: %v\n", collectionName, err)
			continue
		}
		fmt.Printf("Updated %d documents in collection %s\n", result.ModifiedCount, collectionName)
	}
}
