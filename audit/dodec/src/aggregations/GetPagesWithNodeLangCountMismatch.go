package aggregations

import (
	"common"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

func GetPagesWithNodeLangCountMismatch(db *mongo.Database, collectionName string, pageIdsWithNodeLangCountMismatch map[string][]string, ctx context.Context) map[string][]string {
	var pageIdsWithCountMismatch []string
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
		pageNodeLangCount := 0
		pageLangArrayCount := 0
		nodeLangCounts := make(map[string]int)
		for _, node := range *doc.Nodes {
			if node.IsRemoved {
				continue
			} else {
				nodeLangCounts[node.Language]++
				pageNodeLangCount++
			}
		}
		for _, object := range doc.Languages {
			for _, counts := range object {
				if counts.Total > 0 {
					pageLangArrayCount += counts.Total
				}
			}
		}
		if pageNodeLangCount != pageLangArrayCount {
			pageIdsWithCountMismatch = append(pageIdsWithCountMismatch, doc.ID)
		}
	}
	if len(pageIdsWithCountMismatch) > 0 {
		pageIdsWithNodeLangCountMismatch[collectionName] = pageIdsWithCountMismatch
	}
	fmt.Printf("I am in collection %s, and found %d page IDs where there is a node count mismatch.\n", collectionName, len(pageIdsWithNodeLangCountMismatch))
	return pageIdsWithNodeLangCountMismatch
}
