package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// GetSubProductCategoryCounts uses the `nestedTwoLevelMap` data structure in the `PerformAggregation` function
func GetSubProductCategoryCounts(db *mongo.Database, collectionName string, subProductCategoryMap map[string]map[string]map[string]int, ctx context.Context) map[string]map[string]map[string]int {
	collection := db.Collection(collectionName)
	categoryPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},
		{{"$unwind", bson.D{{"path", "$nodes"}}}},
		{{
			"$match", bson.D{
				{"sub_product", bson.D{{"$exists", true}}},
			},
		}},
		{{
			"$group", bson.D{
				{"_id", bson.D{
					{"product", "$product"},
					{"subProduct", "$sub_product"},
					{"category", "$nodes.category"},
				}},
				{"codeNodeCount", bson.D{{"$sum", 1}}},
			},
		}},
	}
	cursor, err := collection.Aggregate(ctx, categoryPipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID struct {
				Product    string `bson:"product"`
				SubProduct string `bson:"subProduct"`
				Category   string `bson:"category"`
			} `bson:"_id"`
			CodeNodeCount int `bson:"codeNodeCount"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		if _, exists := subProductCategoryMap[result.ID.Product]; !exists {
			subProductCategoryMap[result.ID.Product] = make(map[string]map[string]int)
		}
		if _, exists := subProductCategoryMap[result.ID.Product][result.ID.SubProduct]; !exists {
			subProductCategoryMap[result.ID.Product][result.ID.SubProduct] = make(map[string]int)
		}
		subProductCategoryMap[result.ID.Product][result.ID.SubProduct][result.ID.Category] += result.CodeNodeCount
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	return subProductCategoryMap
}
