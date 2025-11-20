package aggregations

import (
	"common"
	"context"
	"dodec/types"
	"dodec/utils"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// FindUsageExamplesForMonth looks for docs pages in Atlas that have had a new usage example added during the target month.
// We get a count of new usage examples matching this criteria, return the count and the page_id, and
// track the product and sub-product in the types.NewAppliedUsageExampleCounterByProductSubProduct
func FindUsageExamplesForMonth(db *mongo.Database, collectionName string, productSubProductCounter types.NewAppliedUsageExampleCounterByProductSubProduct, monthForReporting time.Month, ctx context.Context) types.NewAppliedUsageExampleCounterByProductSubProduct {
	// Target a specific month (example for November 2025):
	targetYear := 2025
	monthStart := time.Date(targetYear, monthForReporting, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0) // First day of next month
	// Define the aggregation pipeline
	pipeline := mongo.Pipeline{
		// Find only page documents where the `nodes` value is not null
		{{"$match", bson.D{
			{"_id", bson.D{{"$ne", "summaries"}}},
			{"nodes", bson.D{{"$ne", nil}}}, // Ensure nodes is not null
		}}},

		// Unwind the `nodes` value to match on specific node elements
		{{"$unwind", bson.D{{"path", "$nodes"}}}},

		{{"$match", bson.D{
			{"$and", bson.A{
				// Filter for nodes that have been added during the target month
				bson.D{{"$and", bson.A{
					bson.D{{"nodes.date_added", bson.D{{"$gte", monthStart}}}},
					bson.D{{"nodes.date_added", bson.D{{"$lt", monthEnd}}}},
				}}},
				// Consider only usage examples
				bson.D{{"nodes.category", common.UsageExample}},
			}},
		}}},

		// First group by Product and SubProduct
		bson.D{{"$group", bson.D{
			{"_id", bson.D{
				{"product", "$product"},
				{"subProduct", bson.D{{"$ifNull", bson.A{"$sub_product", "None"}}}},
			}},
			{"nodesPerProduct", bson.D{{"$push", bson.D{
				{"_id", "$_id"},     // Preserve original document _id
				{"nodes", "$nodes"}, // Collect nodes
			}}}},
		}}},
		// Unwind after the first group to regroup by original _id
		bson.D{{"$unwind", bson.D{{"path", "$nodesPerProduct"}}}},
		// Regroup by original document _id within each Product and SubProduct
		bson.D{{"$group", bson.D{
			{"_id", bson.D{
				{"product", "$_id.product"},
				{"subProduct", "$_id.subProduct"},
				{"documentId", "$nodesPerProduct._id"},
			}},
			{"new_applied_usage_examples", bson.D{{"$push", "$nodesPerProduct.nodes"}}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		// Optionally sort by count in descending order
		bson.D{{"$sort", bson.D{{"count", -1}}}},
	}
	// Execute the aggregation
	collection := db.Collection(collectionName)
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		println(fmt.Errorf("failed to execute aggregate query: %v", err))
		return productSubProductCounter
	}
	defer cursor.Close(ctx)
	collectionPagesWithNewAppliedUsageExamples := make([]types.PageIdNewAppliedUsageExamples, 0)
	for cursor.Next(ctx) {
		var result types.PageIdNewAppliedUsageExamples
		if err = cursor.Decode(&result); err != nil {
			println(fmt.Errorf("failed to decode result document: %v", err))
			return productSubProductCounter
		}

		// If a sub-product map for the product does not exist yet, create one
		if _, ok := productSubProductCounter.ProductSubProductCounts[result.ID.Product]; !ok {
			productSubProductCounter.ProductSubProductCounts[result.ID.Product] = make(map[string]int)
		}

		// The docs org would like to see a breakdown of focus areas. For the purpose of reporting this result, I'm arbitrarily
		// assigning some of the key focus areas as "sub-product" if a page ID contains a substring related to these focus
		// areas. That makes it easy to report on these things as sub-products even if they're not officially sub-products.
		resultAdjustedForFocusAreas := utils.GetFocusAreaAsSubProduct(result)
		if resultAdjustedForFocusAreas.ID.SubProduct != "None" {
			productSubProductCounter.ProductSubProductCounts[result.ID.Product][resultAdjustedForFocusAreas.ID.SubProduct] += resultAdjustedForFocusAreas.Count

			// Add the adjusted for focus area count to the product accumulator
			productSubProductCounter.ProductAggregateCount[result.ID.Product] += resultAdjustedForFocusAreas.Count
		} else {
			// If the subproduct is "None", just append the original count
			productSubProductCounter.ProductSubProductCounts[result.ID.Product][result.ID.SubProduct] += result.Count
			// Add the non-adjusted subproduct count to the product accumulator
			productSubProductCounter.ProductAggregateCount[result.ID.Product] += result.Count
		}
		collectionPagesWithNewAppliedUsageExamples = append(collectionPagesWithNewAppliedUsageExamples, resultAdjustedForFocusAreas)
	}
	if err = cursor.Err(); err != nil {
		println(fmt.Errorf("cursor encountered an error: %v", err))
		return productSubProductCounter
	}
	if collectionPagesWithNewAppliedUsageExamples != nil && len(collectionPagesWithNewAppliedUsageExamples) > 0 {
		productSubProductCounter.PagesInCollections[collectionName] = collectionPagesWithNewAppliedUsageExamples
	}
	return productSubProductCounter
}
