package aggregations

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// FindURLContainingString returns the URL for the specified string present in any `page_url` within the collection.
func FindURLContainingString(db *mongo.Database, collectionName string, pageURLMap map[string][]string, ctx context.Context, substring string) map[string][]string {
	collection := db.Collection(collectionName)
	fmt.Printf("Checking for substring '%s' in %s collection\n", substring, collectionName)
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"page_url", bson.D{
				{"$regex", substring}, // Match documents containing the substring.
				{"$options", "i"},     // Optionally make the search case-insensitive.
			}},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation in collection %s: %v", collectionName, err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		fmt.Println("Substring not found in any page URL in the collection")
	}

	// Iterate through the results
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		// Append the URL to the array of matching URLs in the collection, and set the new array as the value for the collection
		existingURLs := pageURLMap[collectionName]
		existingURLs = append(existingURLs, result["page_url"].(string))
		pageURLMap[collectionName] = existingURLs
	}
	return pageURLMap
}
