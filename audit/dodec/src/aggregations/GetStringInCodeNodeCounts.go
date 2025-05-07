package aggregations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetStringInCodeNodeCounts returns the int count for the specified string present in code nodes within the collection.
func GetStringInCodeNodeCounts(db *mongo.Database, collectionName string, stringCountMap map[string]int, ctx context.Context, substring string) map[string]int {
	collection := db.Collection(collectionName)
	stringMatchingPipeline := mongo.Pipeline{
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
		{{"$match", bson.D{
			{"nodes.code", bson.D{
				{"$exists", true},
				{"$type", "string"}, // Ensure code is a string
			}},
		}}},
		{{"$match", bson.D{
			{"nodes.code", bson.D{{"$regex", substring}, {"$options", "i"}}},
		}}},
		{{"$count", "matchingCodeExamples"}},
	}
	cursor, err := collection.Aggregate(ctx, stringMatchingPipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatalf("Error reading cursor: %v", err)
	}
	if len(results) == 0 {
		stringCountMap[collectionName] = 0
	} else {
		stringCountMap[collectionName] = int(results[0]["matchingCodeExamples"].(int32))
	}
	return stringCountMap
}
