package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"pull-audit-data/types"
)

func GetSubProductLanguageCounts(db *mongo.Database, collectionName string, subProductLanguageMap map[string]map[string]map[string]int, ctx context.Context) map[string]map[string]map[string]int {
	collection := db.Collection(collectionName)
	categoryPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},
		{{"$unwind", "$languages"}},
		{{
			"$match", bson.D{
				{"sub_product", bson.D{{"$exists", true}}},
			},
		}},
		{{"$addFields", bson.D{{"languageKey", bson.D{{"$arrayElemAt", bson.A{
			bson.D{{"$objectToArray", "$languages"}}, 0,
		}}}}}}},
		{{"$match", bson.D{{"languageKey.v.total", bson.D{{"$gt", 0}}}}}},
		{{"$group", bson.D{
			{"_id", bson.D{
				{"product", "$product"},
				{"subProduct", bson.D{{"$ifNull", bson.A{"$sub_product", ""}}}},
				{"language", "$languageKey.k"},
			}},
			{"totalSum", bson.D{{"$sum", "$languageKey.v.total"}}},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, categoryPipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result types.SubProductLanguageResult
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		if _, exists := subProductLanguageMap[result.ID.Product]; !exists {
			subProductLanguageMap[result.ID.Product] = make(map[string]map[string]int)
		}
		if _, exists := subProductLanguageMap[result.ID.Product][result.ID.SubProduct]; !exists {
			subProductLanguageMap[result.ID.Product][result.ID.SubProduct] = make(map[string]int)
		}
		subProductLanguageMap[result.ID.Product][result.ID.SubProduct][result.ID.Language] += result.TotalSum
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	return subProductLanguageMap
}
