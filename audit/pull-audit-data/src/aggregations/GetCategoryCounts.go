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
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},
		{{"$unwind", bson.D{{"path", "$nodes"}}}},
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
							bson.D{{"$eq", bson.A{"$nodes.category", "Task-based usage"}}},
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
		var result types.CountResult
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
	if _, exists := categoryCountMap["total"]; exists {
		categoryCountMap["total"] += categoryCount
	} else {
		categoryCountMap["total"] = categoryCount
	}
	if _, exists := categoryCountMap["complex examples"]; exists {
		categoryCountMap["complex examples"] += complexExampleCount
	} else {
		categoryCountMap["complex examples"] = complexExampleCount
	}
	return categoryCountMap
}
