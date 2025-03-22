package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

func GetLangCountsFromNodes(db *mongo.Database, collectionName string, languageCountMap map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	languagePipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},
		{{"$unwind", bson.D{{"path", "$nodes"}}}},
		// Filter to omit nodes that have been removed from a docs page
		{{"$match", bson.D{
			{"$or", bson.A{
				bson.D{{"nodes.is_removed", bson.D{{"$exists", false}}}},
				bson.D{{"nodes.is_removed", false}},
			}},
		}}},
		// Filter for nodes with a non-empty, existing code field
		{{"$group", bson.D{
			{"_id", "$nodes.language"},
			{"count", bson.D{{"$sum", 1}}},
		}},
		},
	}
	cursor, err := collection.Aggregate(ctx, languagePipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	// Process results and update countMap
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err = cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
		// Accumulate the counts for each _id
		languageCountMap[result.ID] += result.Count
	}
	if err = cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	return languageCountMap
}
