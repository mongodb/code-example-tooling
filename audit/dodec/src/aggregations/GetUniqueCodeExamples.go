package aggregations

import (
	"common"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetUniqueCodeExampleUpdatesForDocsSection looks for pages within the docs section whose code nodes have been added
// or had updates in the last week, omits any removed code nodes, and then does some post-processing to get counts
// broken down by unique and aggregate appearances per page, per category, and in total.
// NOTE: This func does not return data to print in our nicely-formatted tables; instead, it logs directly to console.
func GetUniqueCodeExampleUpdatesForDocsSection(db *mongo.Database, ctx context.Context) {
	// ------ CONFIGURATION: Set these values for your docs set and section ----------
	collectionName := "cloud-docs"                // Replace this with the name of the docs set you want to search within
	docsSectionSubstring := "atlas-vector-search" // Replace this with a substring that represents the docs section where you want to focus results
	// ------ END CONFIGURATION --------------------------------------------------

	// Define a struct to match the aggregation output
	type AggregatedDocsPage struct {
		ID    string            `bson:"_id"`   // Grouping field (DocsPage ID from `$group`)
		Nodes []common.CodeNode `bson:"nodes"` // Grouped array of nodes
	}

	// Calculate last week's date range
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	collection := db.Collection(collectionName)
	// Define the aggregation pipeline
	pipeline := mongo.Pipeline{
		// Step 1: Unwind the `nodes` array.
		{{"$unwind", bson.D{{"path", "$nodes"}}}},

		// Step 2: Match relevant recently added or updated `CodeNode` objects.
		{{"$match", bson.D{
			{"$and", bson.A{
				bson.D{{"$or", bson.A{
					bson.D{{"nodes.date_updated", bson.M{"$gte": oneWeekAgo}}},
					bson.D{{"nodes.date_added", bson.M{"$gte": oneWeekAgo}}},
				}}},
				// Filter out any removed code examples - we only care about net new or updated code examples for this piece
				bson.D{{"$or", bson.A{
					bson.D{{"nodes.is_removed", bson.M{"$exists": false}}},
					bson.D{{"nodes.is_removed", false}},
				}}},
			}},
		}}},

		// Step 3: Group `CodeNode` objects back into arrays per `DocsPage`.
		{{"$group", bson.D{
			{"_id", "$_id"},                        // Group by DocsPage ID.
			{"nodes", bson.D{{"$push", "$nodes"}}}, // Gather filtered nodes into an array.
		}}},

		// Step 4: Match `_id` whose string value contains the docs section substring to omit results for any other pages - i.e. `atlas-search`
		bson.D{{"$match", bson.D{
			{"_id", bson.D{
				{"$regex", docsSectionSubstring},
				{"$options", "i"},
			}},
		}}},
	}

	// Execute the aggregation
	cur, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		fmt.Println("Error during aggregation:", err)
		return
	}
	defer cur.Close(ctx)

	// Process the aggregation results
	var docsPages []AggregatedDocsPage
	if err := cur.All(ctx, &docsPages); err != nil {
		fmt.Println("Error decoding aggregation results:", err)
		return
	}

	// Map to track distinct SHA-256 hashes across all pages.
	sha256Map := make(map[string]int)

	// Map to store category-specific counts
	categoryResults := make(map[string]struct {
		UniqueCount        int
		AggregateInstances int
	})

	// Map to store results for each page.
	pageResults := make(map[string]struct {
		DistinctCount   int
		InstancesOnPage int
	})

	// Iterate through each DocsPage
	for _, page := range docsPages {
		pageID := page.ID
		distinctHashes := make(map[string]bool) // Track unique hashes for this page.
		totalInstances := 0                     // Track total `instances_on_page` for this page.

		// Iterate through the `nodes` array in current DocsPage
		for _, node := range page.Nodes {
			if !distinctHashes[node.SHA256Hash] {
				distinctHashes[node.SHA256Hash] = true
			}

			// Increment the totalInstances count directly to account for nodes without an InstancesOnPage field.
			if node.InstancesOnPage > 0 {
				totalInstances += node.InstancesOnPage
			} else {
				totalInstances++ // Increment by 1 if InstancesOnPage is missing or is 0
			}

			// Add this node's InstancesOnPage (or 1 if missing) to the global sha256Map.
			if node.InstancesOnPage > 0 {
				sha256Map[node.SHA256Hash] += node.InstancesOnPage
			} else {
				sha256Map[node.SHA256Hash]++
			}

			// Track category-specific statistics
			if _, exists := categoryResults[node.Category]; !exists {
				categoryResults[node.Category] = struct {
					UniqueCount        int
					AggregateInstances int
				}{0, 0}
			}
			categoryStats := categoryResults[node.Category]
			categoryStats.UniqueCount++
			if node.InstancesOnPage > 0 {
				categoryStats.AggregateInstances += node.InstancesOnPage
			} else {
				categoryStats.AggregateInstances++ // Increment by 1 if InstancesOnPage is missing or 0
			}
			categoryResults[node.Category] = categoryStats
		}

		// Track results for this page.
		pageResults[pageID] = struct {
			DistinctCount   int
			InstancesOnPage int
		}{
			DistinctCount:   len(distinctHashes),
			InstancesOnPage: totalInstances,
		}
	}

	// Output: Log page-specific results.
	/* Example:

	Page Results:
	Page ID: atlas-search|field-types|number-type, Unique code examples: 78, Total code examples: 162
	Page ID: atlas-search|field-types|uuid-type, Unique code examples: 39, Total code examples: 97
	Page ID: atlas-search|geoWithin, Unique code examples: 68, Total code examples: 54
	*/
	fmt.Println("Page Results:")
	for pageID, result := range pageResults {
		fmt.Printf("Page ID: %s, Unique code examples: %d, Total code examples: %d\n",
			pageID, result.DistinctCount, result.InstancesOnPage)
	}

	// Output: Log global `sha256Map` results.
	/* Example:

	Global Results:
	Unique code example count: 444, Total Instances Across All Pages: 1524
	*/
	distinctShaCount := len(sha256Map)
	totalInstances := 0
	for _, count := range sha256Map {
		totalInstances += count
	}
	fmt.Printf("\nGlobal Results:\nUnique code example count: %d, Total Instances Across All Pages: %d\n",
		distinctShaCount, totalInstances)

	// Output: Log category-specific results
	/* Example:

	Category Breakdown:
	Category: Example configuration object, Unique code example count: 159, Total Instances: 268
	Category: Usage example, Unique code example count: 261, Total Instances: 561
	Category: Syntax example, Unique code example count: 169, Total Instances: 453
	Category: Non-MongoDB command, Unique code example count: 93, Total Instances: 195
	Category: Example return object, Unique code example count: 40, Total Instances: 47
	*/
	fmt.Println("\nCategory Breakdown:")
	for category, stats := range categoryResults {
		fmt.Printf("Category: %s, Unique code example count: %d, Total Instances: %d\n",
			category, stats.UniqueCount, stats.AggregateInstances)
	}
}
