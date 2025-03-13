package updates

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"pull-audit-data/types"
)

// RenameValue looks for any document where a field whose name and old value match the filter you define, and sets the
// field's value to the new value you define.
func RenameValue(db *mongo.Database, ctx context.Context) {
	// List collection names
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	oldValue := "Task-based usage"
	newValue := types.UsageExample
	// Define the filter and update document
	// The filter matches documents that have a 'nodes' array with at least one element with a fieldName matching the oldValue
	filter := bson.M{"nodes": bson.M{"$elemMatch": bson.M{"category": oldValue}}}
	// The update operation uses array filters to update only the elements of the 'nodes' array that match the condition
	update := bson.M{
		"$set": bson.M{"nodes.$[elem].category": newValue},
	}
	// Use the UpdateManyOptionsBuilder to set array filters
	arrayFilters := []interface{}{bson.M{"elem.category": oldValue}}
	updateOptions := options.UpdateMany().SetArrayFilters(arrayFilters)
	// Iterate over each collection and perform the update
	for _, collectionName := range collections {
		collection := db.Collection(collectionName)
		result, err := collection.UpdateMany(ctx, filter, update, updateOptions)
		if err != nil {
			log.Printf("Failed to update documents in collection %s: %v", collectionName, err)
			continue
		}
		log.Printf("Updated %d documents in collection %s", result.ModifiedCount, collectionName)
	}
}
