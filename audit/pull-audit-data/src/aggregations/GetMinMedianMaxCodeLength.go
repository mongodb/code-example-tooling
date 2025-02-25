package aggregations

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"pull-audit-data/types"
)

func GetMinMedianMaxCodeLength(db *mongo.Database, collectionName string, ctx context.Context) types.CodeLengthStats {
	collection := db.Collection(collectionName)
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"nodes", bson.D{
				{"$ne", bson.A{}}, // Ensure nodes array is not empty
				{"$ne", nil},      // Ensure nodes array is not null
			}},
		}}},
		{{"$unwind", bson.D{{"path", "$nodes"}}}},
		{{"$match", bson.D{
			{"nodes.code", bson.D{{"$type", "string"}}}, // Ensure code is a string
		}}},
		{{"$project", bson.D{
			{"codeLength", bson.D{{"$strLenCP", "$nodes.code"}}},
			{"shortCode", bson.D{{"$lt", bson.A{bson.D{{"$strLenCP", "$nodes.code"}}, 80}}}}, // Determine if code length < 80
		}}},
		{{"$group", bson.D{
			{"_id", nil},
			{"minLength", bson.D{{"$min", "$codeLength"}}},
			{"maxLength", bson.D{{"$max", "$codeLength"}}},
			{"allLengths", bson.D{{"$push", "$codeLength"}}},
			{"shortCodeCount", bson.D{{"$sum", bson.D{{"$cond", bson.A{"$shortCode", 1, 0}}}}}}, // Sum shortCode instances
		}}},
		// Sort lengths and calculate median
		{{"$project", bson.D{
			{"minLength", 1},
			{"maxLength", 1},
			{"sortedLengths", bson.D{{"$sortArray", bson.D{{"input", "$allLengths"}, {"sortBy", 1}}}}},
			{"count", bson.D{{"$size", "$allLengths"}}},
			{"shortCodeCount", 1}, // Include shortCodeCount in output
		}}},
		{{"$addFields", bson.D{
			{"medianLength", bson.D{{
				"$cond", bson.D{
					{"if", bson.D{{"$eq", bson.A{"$count", 0}}}},
					{"then", nil},
					{"else", bson.D{{"$arrayElemAt", bson.A{"$sortedLengths", bson.D{{"$floor", bson.D{{"$divide", bson.A{"$count", 2}}}}}}}}},
				},
			}}},
		}}},
	}
	//pipeline := mongo.Pipeline{
	//	{{"$match", bson.D{
	//		{"nodes", bson.D{
	//			{"$ne", bson.A{}}, // Ensure nodes array is not empty
	//			{"$ne", nil},      // Ensure nodes array is not null
	//		}},
	//	}}},
	//	{{"$unwind", bson.D{{"path", "$nodes"}}}},
	//	{{"$match", bson.D{
	//		{"nodes.code", bson.D{{"$type", "string"}}}, // Ensure code is a string
	//	}}},
	//	{{"$project", bson.D{{"codeLength", bson.D{{"$strLenCP", "$nodes.code"}}}}}},
	//	{{"$group", bson.D{
	//		{"_id", nil},
	//		{"minLength", bson.D{{"$min", "$codeLength"}}},
	//		{"maxLength", bson.D{{"$max", "$codeLength"}}},
	//		{"allLengths", bson.D{{"$push", "$codeLength"}}},
	//	}}},
	//	// Sort lengths and calculate median
	//	{{"$project", bson.D{
	//		{"minLength", 1},
	//		{"maxLength", 1},
	//		{"sortedLengths", bson.D{{"$sortArray", bson.D{{"input", "$allLengths"}, {"sortBy", 1}}}}},
	//		{"count", bson.D{{"$size", "$allLengths"}}},
	//	}}},
	//	{{"$addFields", bson.D{
	//		{"medianLength", bson.D{{
	//			"$cond", bson.D{
	//				{"if", bson.D{{"$eq", bson.A{"$count", 0}}}},
	//				{"then", nil},
	//				{"else", bson.D{{"$arrayElemAt", bson.A{"$sortedLengths", bson.D{{"$floor", bson.D{{"$divide", bson.A{"$count", 2}}}}}}}}},
	//			},
	//		}}},
	//	}}},
	//}
	// Execute aggregation
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal("Failed to execute aggregation:", err)
	}
	defer cursor.Close(ctx)
	var result struct {
		MinLength      int `bson:"minLength"`
		MaxLength      int `bson:"maxLength"`
		MedianLength   int `bson:"medianLength"`
		ShortCodeCount int `bson:"shortCodeCount"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error in collection %s: %v", collectionName, err)
	}
	return types.CodeLengthStats{
		Min:            result.MinLength,
		Median:         result.MedianLength,
		Max:            result.MaxLength,
		ShortCodeCount: result.ShortCodeCount,
	}
}
