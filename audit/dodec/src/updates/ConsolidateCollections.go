package updates

import (
	"common"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

func ConsolidateCollections(client *mongo.Client, ctx context.Context) {
	sourceDb := client.Database("code_metrics")
	targetDbName := "ask_cal"
	targetDb := client.Database(targetDbName)
	targetCollName := "consolidated_examples"
	targetColl := targetDb.Collection(targetCollName)
	// List all collections in the source database
	collectionNames, err := sourceDb.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Fatalf("Error listing collections: %v", err)
	}
	// Iterate over each collection
	for _, collName := range collectionNames {
		sourceColl := sourceDb.Collection(collName)
		// Fetch all documents from the source collection
		cursor, err := sourceColl.Find(ctx, bson.D{})
		if err != nil {
			log.Fatalf("Error finding documents in collection %s: %v", collName, err)
		}
		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err := cursor.Close(ctx)
			if err != nil {
				log.Fatalf("Error closing cursor: %v", err)
			}
		}(cursor, ctx)
		var updatedDocuments []common.CalDocsPage
		for cursor.Next(ctx) {
			var doc bson.M
			if err = cursor.Decode(&doc); err != nil {
				log.Fatalf("Error decoding document in collection %s: %v", collName, err)
			}
			idValue, ok := doc["_id"].(string)
			if ok {
				// Skip documents where '_id' is "summaries"
				if idValue == "summaries" {
					continue
				} else {
					// Deserialize into DocsPage
					var docsPage common.DocsPage
					if err := cursor.Decode(&docsPage); err != nil {
						log.Fatalf("Error decoding document into DocsPage: %v", err)
					}

					// If the page has been removed from Snooty/is no longer live, skip it - we don't want those examples in Ask Cal
					if docsPage.IsRemoved {
						continue
					}

					// If the page has no code examples, skip it - we only care about pages with code examples for Ask Cal
					if docsPage.CodeNodesTotal == 0 {
						continue
					}

					newID := bson.NewObjectID() // Generate a new unique ObjectID for the page
					// Atlas Search can only facet on a top-level field, so we need to create a top-level field of languages for faceting
					var languagesFacet []string
					if docsPage.Nodes != nil && len(*docsPage.Nodes) > 0 {
						for _, node := range *docsPage.Nodes {
							if !node.IsRemoved {
								if !Contains(languagesFacet, node.Language) {
									languagesFacet = append(languagesFacet, node.Language)
								}
							}
						}
					}

					// Convert the DocsPage into a modified version of the page with an ObjectID identifier and the origin collection name
					updatedDoc := common.CalDocsPage{
						ID:                   newID,
						CodeNodesTotal:       docsPage.CodeNodesTotal,
						DateAdded:            docsPage.DateAdded,
						DateLastUpdated:      docsPage.DateLastUpdated,
						IoCodeBlocksTotal:    docsPage.IoCodeBlocksTotal,
						Languages:            docsPage.Languages,
						LiteralIncludesTotal: docsPage.LiteralIncludesTotal,
						Nodes:                docsPage.Nodes,
						PageURL:              docsPage.PageURL,
						ProjectName:          docsPage.ProjectName,
						Product:              docsPage.Product,
						SubProduct:           docsPage.SubProduct,
						Keywords:             docsPage.Keywords,
						DateRemoved:          docsPage.DateRemoved,
						IsRemoved:            docsPage.IsRemoved,
					}

					if languagesFacet != nil {
						updatedDoc.LanguagesFacet = languagesFacet
					}
					updatedDocuments = append(updatedDocuments, updatedDoc)
				}
			} else {
				fmt.Println("Field '_id' does not exist or is not a string in the document")
				continue
			}
		}
		if len(updatedDocuments) > 0 {
			_, err = targetColl.InsertMany(ctx, updatedDocuments)
			if err != nil {
				log.Fatalf("Error inserting documents into target DB %s, collection %s: %v", targetDbName, targetCollName, err)
			}
			log.Printf("Copied %d documents from %s to collection %s", len(updatedDocuments), collName, targetCollName)
		}
	}
	log.Println("All collections copied successfully")
}

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
