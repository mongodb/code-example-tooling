package aggregations

import (
	"common"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetLangCountFromLangArrayManually processes the `Languages` field locally on device to get a count of all code examples
// in each programming language within the collection. This function was used to debug discrepancies between the node language
// count and the languages array count returned by other aggregations.
func GetLangCountFromLangArrayManually(db *mongo.Database, collectionName string, languageCountMap map[string]int, ctx context.Context) map[string]int {
	collection := db.Collection(collectionName)
	// Define the aggregation pipeline
	filter := bson.D{{"_id", bson.D{{"$ne", "summaries"}}}, // Exclude documents with _id "summaries"
		{"nodes", bson.D{{"$ne", bson.A{}}}}, // Exclude documents with nodes array empty
		{"nodes", bson.D{{"$ne", nil}}},      // Exclude documents with nodes array null
	} // Empty filter to get all documents

	// Find all documents
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	// Iterate over the cursor to access each document
	var docs []common.DocsPage
	if err = cursor.All(ctx, &docs); err != nil {
		log.Fatal(err)
	}
	// Print each document
	for _, doc := range docs {
		for _, object := range doc.Languages {
			for lang, counts := range object {
				languageCountMap[lang] = languageCountMap[lang] + counts.Total
			}
		}
	}
	return languageCountMap
}
