package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"pull-audit-data/types"
)

// GetLanguageCounts returns a `simpleMap` data structure as defined in the PerformAggregation function. Each key is the
// name of a programming language, and the int value is the count of code examples in that programming language. The map
// also contains a types.Total key whose value is the aggregate count of all the code examples across all the collections.
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
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
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
	if _, exists := languageCountMap[types.Total]; exists {
		languageCountMap[types.Total] += langCount
	} else {
		languageCountMap[types.Total] = langCount
	}
	return languageCountMap
}
