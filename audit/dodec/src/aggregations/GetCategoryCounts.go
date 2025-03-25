package aggregations

import (
	"common"
	"context"
	"dodec/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// GetCategoryCounts returns a `simpleMap` data structure as defined in the PerformAggregation function. Each key
// is the category name, and the int value is the count of code examples for that category. The map also contains
// a types.ComplexExamples key whose value is the count of Usage Examples where the code example character count is greater
// than 500 characters, and a types.Total key whose value is an aggregate count of all code examples regardless of category.
func GetCategoryCounts(db *mongo.Database, collectionName string, categoryCountMap map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	categoryPipeline := mongo.Pipeline{
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
		{{"$match", bson.D{
			{"nodes.code", bson.D{
				{"$exists", true},
				{"$ne", ""},
			}},
		}}},
		{{"$addFields", bson.D{{"isLongCode", bson.D{{"$gte", bson.A{bson.D{{"$strLenCP", "$nodes.code"}}, 500}}}}}}},
		{{"$group", bson.D{
			{"_id", "$nodes.category"},
			{"count", bson.D{{"$sum", 1}}},
			{"longCodeCount", bson.D{{"$sum", bson.D{
				{"$cond", bson.A{
					bson.D{
						{"$and", bson.A{
							bson.D{{"$eq", bson.A{"$nodes.category", common.UsageExample}}},
							"$isLongCode",
						}},
					},
					1,
					0,
				}},
			}}},
			}}},
		},
	}
	cursor, err := collection.Aggregate(ctx, categoryPipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)
	// Process results and update countMap

	categoryCount := 0
	complexExampleCount := 0
	for cursor.Next(ctx) {
		var result struct {
			ID            string `bson:"_id"`
			Count         int    `bson:"count"`
			LongCodeCount int    `bson:"longCodeCount"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
		// Accumulate the counts for each _id
		categoryCountMap[result.ID] += result.Count
		categoryCount += result.Count
		complexExampleCount += result.LongCodeCount
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	if _, exists := categoryCountMap[types.Total]; exists {
		categoryCountMap[types.Total] += categoryCount
	} else {
		categoryCountMap[types.Total] = categoryCount
	}
	if _, exists := categoryCountMap[types.ComplexExamples]; exists {
		categoryCountMap[types.ComplexExamples] += complexExampleCount
	} else {
		categoryCountMap[types.ComplexExamples] = complexExampleCount
	}
	return categoryCountMap
}
