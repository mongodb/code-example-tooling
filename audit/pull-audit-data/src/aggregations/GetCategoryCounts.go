package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"pull-audit-data/types"
)

func GetCategoryCounts(db *mongo.Database, collectionName string, categoryCountMap map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	categoryPipeline := mongo.Pipeline{
		{{"$match", bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}}},
		{{"$unwind", bson.D{{"path", "$nodes"}}}},
		{{"$group", bson.D{{"_id", "$nodes.category"}, {"count", bson.D{{"$sum", 1}}}}}},
	}
	cursor, err := collection.Aggregate(ctx, categoryPipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	// Process results and update countMap

	categoryCount := 0
	for cursor.Next(ctx) {
		var result types.CountResult
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
		// Accumulate the counts for each _id
		categoryCountMap[result.ID] += result.Count
		categoryCount += result.Count
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	if _, exists := categoryCountMap["total"]; exists {
		categoryCountMap["total"] += categoryCount
	} else {
		categoryCountMap["total"] = categoryCount
	}
	return categoryCountMap
}
