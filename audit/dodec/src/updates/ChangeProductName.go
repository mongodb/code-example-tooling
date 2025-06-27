package updates

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// ChangeProductName sets the `product` field value to a new value that you specify for all documents in the given collection.
// Run this function to populate any changes made to the mapped product names in `GetProductSubProduct` in GDCD.
func ChangeProductName(db *mongo.Database, ctx context.Context) {

	// ===== CONFIGURATION: Set these values before running =====
	collection := db.Collection("atlas-architecture") // collection to update (this should match the project name in GDCD)
	newProductName := "Atlas Architecture Center"     // New product name to apply to the documents
	// ==========================================================

	// Omit the summary document, as the `$set` operator would add this field to the doc
	filter := bson.M{
		"_id": bson.M{
			"$ne": "summaries",
		},
	}

	// Define the update to set the Product field value
	update := bson.M{
		"$set": bson.M{
			"product": newProductName,
		},
	}

	// Perform the update
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Fatalf("Failed to update documents: %v", err)
	}

	// Print the result
	fmt.Printf("Matched %d documents and modified %d documents\n", result.MatchedCount, result.ModifiedCount)
}
