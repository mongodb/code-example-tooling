package aggregations

import (
	"common"
	"context"
	"dodec/types"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

func FindNewAppliedUsageExamples(db *mongo.Database, collectionName string, pageIdsWithNewUsageExamples map[string][]types.PageIdNewAppliedUsageExamples, ctx context.Context) map[string][]types.PageIdNewAppliedUsageExamples {
	// Calculate last week's date range
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
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
				// Filter for nodes that have been added or updated in the last week
				bson.D{{"$or", bson.A{
					bson.D{{"nodes.date_added", bson.D{{"$gte", oneWeekAgo}}}},
					//bson.D{{"nodes.date_updated", bson.D{{"$gte", oneWeekAgo}}}},
				}}},
				// Consider only usage examples
				bson.D{{"nodes.category", common.UsageExample}},
				// Specify a minimum code character count to consider it a new applied usage example
				bson.D{{"$expr", bson.D{
					{"$gt", bson.A{bson.D{{"$strLenCP", "$nodes.code"}}, 300}},
				}}},
			}},
		}}},

		// Group documents by _id and collect matching nodes
		bson.D{{"$group", bson.D{
			{"_id", "$_id"},
			{"matchingNodes", bson.D{{"$push", "$nodes"}}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	// Execute the aggregation
	collection := db.Collection(collectionName)
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		println(fmt.Errorf("failed to execute aggregate query: %v", err))
		return pageIdsWithNewUsageExamples
	}
	defer cursor.Close(ctx)
	collectionPagesWithNewAppliedUsageExamples := make([]types.PageIdNewAppliedUsageExamples, 0)
	for cursor.Next(ctx) {
		var result types.PageIdNewAppliedUsageExamples
		if err := cursor.Decode(&result); err != nil {
			println(fmt.Errorf("failed to decode result document: %v", err))
			return pageIdsWithNewUsageExamples
		}
		collectionPagesWithNewAppliedUsageExamples = append(collectionPagesWithNewAppliedUsageExamples, result)
	}
	if err := cursor.Err(); err != nil {
		println(fmt.Errorf("cursor encountered an error: %v", err))
		return pageIdsWithNewUsageExamples
	}
	if collectionPagesWithNewAppliedUsageExamples != nil && len(collectionPagesWithNewAppliedUsageExamples) > 0 {
		pageIdsWithNewUsageExamples[collectionName] = collectionPagesWithNewAppliedUsageExamples
	}
	return pageIdsWithNewUsageExamples
}
