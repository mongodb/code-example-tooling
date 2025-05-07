package aggregations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func FindDocsMissingProduct(db *mongo.Database, collectionName string, pageIdsMissingProduct map[string][]string, ctx context.Context) map[string][]string {
	var pageIds []string
	collection := db.Collection(collectionName)
	filter := bson.M{
		"$and": []bson.M{
			{"_id": bson.M{"$ne": "summaries"}},
			{"$or": []bson.M{
				{"product": bson.M{"$exists": false}},
				{"product": ""},
			}},
		},
	}
	// Define projection to get only the _id field
	projection := bson.M{"_id": 1}
	// Find documents
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var result bson.M
		if err = cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		if id, ok := result["_id"].(string); ok {
			pageIds = append(pageIds, id)
		}
	}
	if err = cursor.Err(); err != nil {
		log.Fatal(err)
	}
	if len(pageIds) > 0 {
		pageIdsMissingProduct[collectionName] = pageIds
	}
	return pageIdsMissingProduct
}
