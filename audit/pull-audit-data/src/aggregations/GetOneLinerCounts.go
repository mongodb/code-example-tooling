package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

func GetOneLinerCounts(db *mongo.Database, collectionName string, oneLinerCountMap map[string]int, totalCount int, ctx context.Context) (map[string]int, int) {
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
			{"nodes.category", "Task-based usage"}, // Ensure category is "Usage Example"
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
	return oneLinerCountMap, totalCount
}
