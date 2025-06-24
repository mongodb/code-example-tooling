package aggregations

import (
	"context"
	"dodec/types"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetDocsIdsWithRecentActivity returns a slice of types.PageIdChangedCounts for the given collection. This aggregation
// finds all docs pages with activity within the last week, and appends the page ID along with a count of each change type
// to a list of pages with changes in the given collection. The string map key is the collection name.
func GetDocsIdsWithRecentActivity(db *mongo.Database, collectionName string, aggregatePageIdCounts map[string][]types.PageIdChangedCounts, ctx context.Context) map[string][]types.PageIdChangedCounts {
	// Calculate last week's date range
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	// Define the aggregation pipeline
	pipeline := mongo.Pipeline{
		{{"$unwind", bson.D{{"path", "$nodes"}}}},
		{{"$match", bson.D{
			{"$or", bson.A{
				bson.D{{"nodes.date_updated", bson.M{"$gte": oneWeekAgo}}},
				bson.D{{"nodes.date_added", bson.M{"$gte": oneWeekAgo}}},
				bson.D{{"nodes.date_removed", bson.M{"$gte": oneWeekAgo}}},
			}},
		}}},
		{{"$group", bson.D{
			{"_id", "$_id"},
			{"added_count", bson.D{{"$sum", bson.D{
				{"$cond", bson.A{
					bson.D{{"$gte", bson.A{"$nodes.date_added", oneWeekAgo}}},
					1,
					0,
				}},
			}}}},
			{"updated_count", bson.D{{"$sum", bson.D{
				{"$cond", bson.A{
					bson.D{{"$gte", bson.A{"$nodes.date_updated", oneWeekAgo}}},
					1,
					0,
				}},
			}}}},
			{"removed_count", bson.D{{"$sum", bson.D{
				{"$cond", bson.A{
					bson.D{{"$gte", bson.A{"$nodes.date_removed", oneWeekAgo}}},
					1,
					0,
				}},
			}}}},
		}}},
	}
	// Execute the aggregation
	collection := db.Collection(collectionName)
	cur, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(ctx)
	var results []types.PageIdChangedCounts
	// Read all results into the results slice
	if err = cur.All(ctx, &results); err != nil {
		fmt.Println(err)
	}
	if len(results) > 0 {
		aggregatePageIdCounts[collectionName] = results
	}
	return aggregatePageIdCounts
}
