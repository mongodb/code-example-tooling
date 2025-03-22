package aggregations

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"pull-audit-data/types"
)

// GetCollectionCount returns a `simpleMap` data structure as defined in the PerformAggregation function. Each key is
// the collection name, and the int value is the count of code examples in that collection. This map also contains
// a `types.Total` key whose value is the aggregate count of all code examples across all collections.
func GetCollectionCount(db *mongo.Database, collectionName string, collectionSums map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	countPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},
		{{"$group", bson.D{
			{"_id", nil},
			{"totalCodeNodes", bson.D{{"$sum", "$code_nodes_total"}}},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, countPipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	// Process results and update countMap

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatalf("Failed to access cursor in collection %s: %v", collectionName, err)
	}

	if len(results) > 0 {
		if total, ok := results[0]["totalCodeNodes"].(int32); ok {
			collectionSums[collectionName] = int(total)
			if collectionSums[types.Total] > 0 {
				collectionSums[types.Total] += int(total)
			} else {
				collectionSums[types.Total] = int(total)
			}
		} else {
			fmt.Println("Error: Could not read total code nodes.")
		}
	}
	return collectionSums
}
