package updates

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
)

// ChangeProjectName sets the `project_name` field value to a new value that you specify for all documents in the given
// collection. Then, it renames the collection. This should match the project name defined in `common`
func ChangeProjectName(client *mongo.Client, dbName string, ctx context.Context) {

	// ===== CONFIGURATION: Set these values before running =====
	oldProjectName := "cluster-sync" // Existing collection to update (this should match the old project name in `common`)
	newProjectName := "mongosync"    // New project name to apply to the documents/rename the collection
	// ==========================================================

	codeMetricsDb := client.Database(dbName)
	collection := codeMetricsDb.Collection(oldProjectName)

	// Omit the summary document, as the `$set` operator would add this field to the doc
	filter := bson.M{
		"_id": bson.M{
			"$ne": "summaries",
		},
	}

	// Define the update to set the ProjectName field value
	update := bson.M{
		"$set": bson.M{
			"project_name": newProjectName,
		},
	}

	// Perform the update
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Fatalf("Failed to update documents: %v", err)
	}

	// Print the result
	fmt.Printf("Matched %d documents and modified %d documents\n", result.MatchedCount, result.ModifiedCount)

	// Then, rename the collection. The Go Driver does not provide a method to do this directly, so using the RunCommand method
	// Form the BSON document to create the rename command
	command := bson.D{
		{"renameCollection", fmt.Sprintf("%s.%s", dbName, oldProjectName)},
		{"to", fmt.Sprintf("%s.%s", dbName, newProjectName)},
	}

	// Execute the renameCollection command
	adminDB := client.Database("admin") // The renameCollection command must be run on the admin database
	renameResult := adminDB.RunCommand(ctx, command)

	// Check for errors and handle the result
	if err := renameResult.Err(); err != nil {
		log.Fatal("Failed to rename collection:", err)
	} else {
		fmt.Printf("Collection renamed successfully from '%s' to '%s'\n", oldProjectName, newProjectName)
	}
}
