package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// GetCollectionCount uses the `simpleMap` data structure in the `PerformAggregation` function
func GetCollectionCount(db *mongo.Database, collectionName string, collectionSums map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	countPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},
		{{"$project", bson.D{
			{"nodeCount", bson.D{{"$size", "$nodes"}}},
		}}},
		{{"$group", bson.D{
			{"_id", nil},
			{"totalNodes", bson.D{{"$sum", "$nodeCount"}}},
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
		count := results[0]["totalNodes"].(int32)
		collectionSums[collectionName] = int(count)
		if collectionSums["total"] > 0 {
			collectionSums["total"] += int(count)
		} else {
			collectionSums["total"] = int(count)
		}
	}
	return collectionSums
}
