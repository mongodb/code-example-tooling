package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"pull-audit-data/types"
)

// GetLanguageCounts uses the `simpleMap` data structure in the `PerformAggregation` function
func GetLanguageCounts(db *mongo.Database, collectionName string, languageCountMap map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	languagePipeline := mongo.Pipeline{
		{{"$match", bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}}},
		{{"$unwind", bson.D{{"path", "$languages"}}}},
		{{"$project", bson.D{{"languagesArray", bson.D{{"$objectToArray", "$languages"}}}}}},
		{{"$unwind", bson.D{{"path", "$languagesArray"}}}},
		{{"$group", bson.D{{"_id", "$languagesArray.k"}, {"count", bson.D{{"$sum", "$languagesArray.v.total"}}}}}},
	}
	cursor, err := collection.Aggregate(ctx, languagePipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	langCount := 0
	// Process results and update countMap
	for cursor.Next(ctx) {
		var result types.CountResult
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
		// Accumulate the counts for each _id
		languageCountMap[result.ID] += result.Count
		langCount += result.Count
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	if _, exists := languageCountMap["total"]; exists {
		languageCountMap["total"] += langCount
	} else {
		languageCountMap["total"] = langCount
	}
	return languageCountMap
}
