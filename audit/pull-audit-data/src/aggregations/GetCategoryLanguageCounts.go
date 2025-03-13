package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// GetCategoryLanguageCounts returns a `nestedOneLevelMap` data structure as defined in the PerformAggregation function.
// The top-level map key is the category name, the nested-map string key is the language name, and the int is the count of
// code examples in that language for that category.
func GetCategoryLanguageCounts(db *mongo.Database, collectionName string, categoryLanguageCountMap map[string]map[string]int, ctx context.Context) map[string]map[string]int {
	collection := db.Collection(collectionName)
	categoryLanguagePipeline := mongo.Pipeline{
		{{"$match", bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}}},
		{{"$match", bson.D{
			{"nodes", bson.D{
				{"$ne", bson.A{}}, // Ensure nodes array is not empty
				{"$ne", nil},      // Ensure nodes array is not null
			}},
		}}},
		{{"$unwind", bson.D{{"path", "$nodes"}, {"preserveNullAndEmptyArrays", false}}}},
		{{"$group", bson.D{
			{"_id", bson.D{
				{"category", "$nodes.category"},
				{"language", "$nodes.language"},
			}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, categoryLanguagePipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	// Process results and update countMap
	for cursor.Next(ctx) {
		var result struct {
			ID struct {
				Category string `bson:"category"`
				Language string `bson:"language"`
			} `bson:"_id"`
			Count int `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		// Initialize map for the category if it doesn't exist
		if _, ok := categoryLanguageCountMap[result.ID.Category]; !ok {
			categoryLanguageCountMap[result.ID.Category] = make(map[string]int)
		}
		// Accumulate the count for each language within the category
		categoryLanguageCountMap[result.ID.Category][result.ID.Language] += result.Count
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	return categoryLanguageCountMap
}
