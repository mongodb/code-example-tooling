package aggregations

import (
	"common"
	"context"
	"dodec/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// GetOneLineUsageExampleCounts returns a `simpleMap` data structure as defined in the PerformAggregation function. The
// key is the collection name, and the int value is the count of one-line code examples in the Usage Example category.
// One-line code example is defined as a code example whose character count is fewer than 80 characters. The map also
// contains a `types.Total` key whose value is the aggregate count of all one-line code examples in the Usage Example category across all collections.
func GetOneLineUsageExampleCounts(db *mongo.Database, collectionName string, oneLinerCountMap map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"nodes", bson.D{
				{"$ne", bson.A{}}, // Ensure nodes array is not empty
				{"$ne", nil},      // Ensure nodes array is not null
			}},
		}}},
		{{"$unwind", bson.D{{"path", "$nodes"}}}}, // Unwind the nodes array to handle each CodeNode separately
		{{"$match", bson.D{
			{"nodes.code", bson.D{
				{"$exists", true},
				{"$type", "string"}, // Ensure code is a string
			}},
			{"nodes.category", common.UsageExample}, // Ensure category is "Usage Example"
		}}},
		// Filter to omit nodes that have been removed from a docs page
		{{"$match", bson.D{
			{"$or", bson.A{
				bson.D{{"nodes.is_removed", bson.D{{"$exists", false}}}},
				bson.D{{"nodes.is_removed", false}},
			}},
		}}},
		{{"$project", bson.D{
			{"codeLength", bson.D{{"$strLenCP", "$nodes.code"}}},
		}}},
		{{"$match", bson.D{
			{"codeLength", bson.D{{"$lt", 80}}}, // Filter where code length is less than 80
		}}},
		{{"$count", "shortCodeCount"}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	// Process results and update countMap
	totalCount := 0
	for cursor.Next(ctx) {
		var result struct {
			ShortCodeCount int `bson:"shortCodeCount"`
		}
		// Decode a single result document
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
		// Add to the total and individual collection counts
		oneLinerCountMap[collectionName] += result.ShortCodeCount
		totalCount += result.ShortCodeCount
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}

	if oneLinerCountMap[types.Total] != 0 {
		oneLinerCountMap[types.Total] += totalCount
	} else {
		oneLinerCountMap[types.Total] = totalCount
	}
	return oneLinerCountMap
}
