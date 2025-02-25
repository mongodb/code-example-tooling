package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"pull-audit-data/types"
)

func GetProductLanguageCounts(db *mongo.Database, collectionName string, productLanguageMap map[string]map[string]int, ctx context.Context) map[string]map[string]int {
	collection := db.Collection(collectionName)
	languagePipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},
		{{"$unwind", "$languages"}},
		{{"$addFields", bson.D{{"languageKey", bson.D{{"$arrayElemAt", bson.A{
			bson.D{{"$objectToArray", "$languages"}}, 0,
		}}}}}}},
		{{"$match", bson.D{{"languageKey.v.total", bson.D{{"$gt", 0}}}}}},
		{{"$group", bson.D{
			{"_id", bson.D{
				{"product", "$product"},
				{"language", "$languageKey.k"},
			}},
			{"totalSum", bson.D{{"$sum", "$languageKey.v.total"}}},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, languagePipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result types.ProductLanguageCount
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		// Aggregate the results
		product := result.ID.Product
		language := result.ID.Language
		total := result.TotalSum
		if _, exists := productLanguageMap[product]; !exists {
			productLanguageMap[product] = make(map[string]int)
		}
		productLanguageMap[product][language] += total
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	return productLanguageMap
}
