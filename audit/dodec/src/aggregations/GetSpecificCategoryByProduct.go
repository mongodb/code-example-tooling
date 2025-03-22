package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// GetSpecificCategoryByProduct returns a `simpleMap` data structure as defined in the PerformAggregation function. The
// key is the product name, and the int is the count of code examples in the given category across the product.
// `types.Constants` has constants for all the category names.
func GetSpecificCategoryByProduct(db *mongo.Database, collectionName string, categoryName string, productSums map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	countPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
			{"nodes.category", categoryName},
		}}},
		{{"$unwind", "$nodes"}},
		// Filter to omit nodes that have been removed from a docs page
		{{"$match", bson.D{
			{"$or", bson.A{
				bson.D{{"nodes.is_removed", bson.D{{"$exists", false}}}},
				bson.D{{"nodes.is_removed", false}},
			}},
		}}},
		{{"$match", bson.D{{"nodes.category", categoryName}}}},
		{{"$group", bson.D{
			{"_id", "$product"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, countPipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	// Process results and update countMap

	// Process aggregation results
	for cursor.Next(context.TODO()) {
		var result struct {
			Product string `bson:"_id"`
			Count   int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		productSums[result.Product] += result.Count
	}
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
	return productSums
}
