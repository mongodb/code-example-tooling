package updates

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AddProductNames adds `product` and `sub_product` (where applicable) fields to documents with values that correspond
// to the Docs Taxonomy. If the document in the collection already has the applicable field(s), no change is made.
func AddProductNames(db *mongo.Database, ctx context.Context) {
	collectionProducts := map[string]string{
		"atlas-cli":             "Atlas",
		"atlas-operator":        "Atlas",
		"bi-connector":          "BI Connector",
		"c":                     "Drivers",
		"charts":                "Atlas",
		"cloud-docs":            "Atlas",
		"cloud-manager":         "Cloud Manager",
		"cloudgov":              "Atlas", // Missing from taxonomy/this is a guess
		"cluster-sync":          "Cluster-to-Cluster sync",
		"compass":               "Compass",
		"cpp-driver":            "Drivers",
		"csharp":                "Drivers",
		"database-tools":        "Database Tools",
		"docs":                  "Server",
		"docs-django":           "Drivers",
		"docs-entity-framework": "Drivers", // Missing from taxonomy/this is a guess
		"docs-golang":           "Drivers",
		"docs-java":             "Drivers",
		"docs-java-rs":          "Drivers",
		"docs-k8s-operator":     "Enterprise Kubernetes Operator",
		"kafka-connector":       "Kafka Connector",
		"kotlin":                "Drivers",
		"kotlin-sync":           "Drivers",
		"laravel":               "Drivers",
		"mongoid":               "Drivers",
		"mongodb-shell":         "MongoDB Shell",
		"mongocli":              "MongoDB CLI",
		"node":                  "Drivers",
		"ops-manager":           "Ops Manager",
		"php-library":           "Drivers", // Missing from taxonomy/this is a guess
		"pymongo":               "Drivers",
		"pymongo-arrow":         "Drivers",
		"relational-migrator":   "Relational Migrator",
		"ruby-driver":           "Drivers",
		"rust":                  "Drivers",
		"scala":                 "Drivers",
		"spark-connector":       "Spark Connector",
	}

	collectionSubProducts := map[string]string{
		"atlas-cli":      "Atlas CLI",
		"atlas-operator": "Kubernetes Operator",
		"charts":         "Charts",
	}

	atlasCollectionSubProductByDir := map[string]string{
		"atlas-stream-processing": "Stream Processing",
		"atlas-search":            "Search",
		"atlas-vector-search":     "Vector Search",
		"data-federation":         "Data Federation",
		"online-archive":          "Online Archive",
		"triggers":                "Triggers",
	}

	for collectionName, productString := range collectionProducts {
		collection := db.Collection(collectionName)

		// Use UpdateMany to add the Product field
		filter := bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}
		update := bson.D{{"$set", bson.D{{"product", productString}}}}
		updateResult, err := collection.UpdateMany(ctx, filter, update)
		if err != nil {
			log.Printf("Could not update documents to add a Product field in collection [%s]: %v", collectionName, err)
			continue
		}

		fmt.Printf("Added a Product field to %d documents in the [%s] collection\n", updateResult.ModifiedCount, collectionName)
	}

	for collectionName, subProductString := range collectionSubProducts {
		collection := db.Collection(collectionName)

		// Define the filter to exclude documents with "_id" equal to "summaries"
		filter := bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}

		// Define the update to add the "sub_product" field
		update := bson.D{{"$set", bson.D{{"sub_product", subProductString}}}}

		// Update all documents that match the filter
		updateResult, err := collection.UpdateMany(ctx, filter, update)
		if err != nil {
			log.Printf("Could not update documents to add a sub-product field in collection [%s]: %v", collectionName, err)
			continue
		}

		// Log the number of documents modified
		fmt.Printf("Added a sub_product field to %d documents in the [%s] collection\n", updateResult.ModifiedCount, collectionName)
	}

	for directorySubpath, subProductString := range atlasCollectionSubProductByDir {
		// These are specifically all subdirectories in Cloud Docs, so hardcoding this here for convenience
		collection := db.Collection("cloud-docs")
		filter := bson.M{
			"page_url": bson.M{
				"$regex":   regexp.QuoteMeta(directorySubpath), // Properly escape special regex characters
				"$options": "i",                                // Case-insensitive match
			},
		}
		// Define the update to add a new field with the desired value
		update := bson.M{
			"$set": bson.M{"sub_product": subProductString},
		}
		// Apply the update to all matching documents
		updateResult, err := collection.UpdateMany(context.TODO(), filter, update)
		if err != nil {
			log.Printf("Could not update documents for substring [%s]: %v", directorySubpath, err)
			continue
		}
		fmt.Printf("Added a sub_product field to %d documents in the collection for substring [%s]\n", updateResult.ModifiedCount, directorySubpath)
	}
}
