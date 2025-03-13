package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// GetSpecificLanguageCount returns the int count for the specified language within the collection.
func GetSpecificLanguageCount(db *mongo.Database, collectionName string, language string, ctx context.Context) int {
	collection := db.Collection(collectionName)
	languagePipeline := mongo.Pipeline{
		{{"$match", bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}}},
		{{"$unwind", bson.D{{"path", "$languages"}}}},
		{{"$project", bson.D{{"languagesArray", bson.D{{"$objectToArray", "$languages"}}}}}},
		{{"$unwind", bson.D{{"path", "$languagesArray"}}}},
		{{"$match", bson.D{{"languagesArray.k", language}}}}, // Replace 'specificLanguage' with the desired language
		{{"$group", bson.D{{"_id", "$languagesArray.k"}, {"count", bson.D{{"$sum", "$languagesArray.v.total"}}}}}},
	}
	cursor, err := collection.Aggregate(ctx, languagePipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	langCount := 0
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
		langCount = result.Count
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	return langCount
}
