package main

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"pull-audit-data/aggregations"
)

// PerformAggregation executes several different aggregation operations for every collection in the DB, and logs the output to console.
func PerformAggregation(db *mongo.Database, ctx context.Context) {
	// The aggregations in this project use one of these data structures. Uncomment the corresponding data structure,
	// and/or make duplicates with appropriate names as needed
	simpleMap := make(map[string]int)
	//lengthCountMap := make(map[string]types.CodeLengthStats)
	//nestedOneLevelMap := make(map[string]map[string]int)
	//nestedTwoLevelMap := make(map[string]map[string]map[string]int)

	// If you just need to get data for a single collection, perform the aggregation using the collection name
	//simpleMap = aggregations.GetLanguageCounts(db, "pymongo", simpleMap, ctx)

	// If you need to get data across all the collections in the `code-examples` database, iterate through the collections
	emptyFilter := bson.D{}
	collectionNames, err := db.ListCollectionNames(ctx, emptyFilter)
	if err != nil {
		panic(err)
	}

	for _, collectionName := range collectionNames {
		//simpleMap = aggregations.GetCategoryCounts(db, collectionName, simpleMap, ctx)
		//simpleMap = aggregations.GetLanguageCounts(db, collectionName, simpleMap, ctx)
		//lengthCountMap[collectionName] = aggregations.GetMinMedianMaxCodeLength(db, collectionName, ctx)
		//nestedOneLevelMap = aggregations.GetCategoryLanguageCounts(db, collectionName, nestedOneLevelMap, ctx)
		//nestedOneLevelMap = aggregations.GetProductCategoryCounts(db, collectionName, nestedOneLevelMap, ctx)
		//nestedTwoLevelMap = aggregations.GetSubProductCategoryCounts(db, collectionName, nestedTwoLevelMap, ctx)
		//simpleMap = aggregations.GetOneLineUsageExampleCounts(db, collectionName, oneLineUsageExampleCountMap, totalOneLineUsageExampleCount, ctx)
		//simpleMap = aggregations.GetOneLinerCounts(db, collectionName, simpleMap, ctx)
		//nestedOneLevelMap = aggregations.GetProductLanguageCounts(db, collectionName, nestedOneLevelMap, ctx)
		//nestedTwoLevelMap = aggregations.GetSubProductLanguageCounts(db, collectionName, nestedTwoLevelMap, ctx)
		simpleMap = aggregations.GetCollectionCount(db, collectionName, simpleMap, ctx)
	}

	// TODO: Refactor this to a nice Print function so we don't need all the accumulator boilerplate here
	//minAccumulator := 0
	//medianAccumulator := 0
	//maxAccumulator := 0
	//collectionCount := 0
	//shortCodeCount := 0
	//for name, stats := range lengthCountMap {
	//	fmt.Printf("Collection: %s, Min: %d, Median: %d, Max: %d, One-Liners: %d\n", name, stats.Min, stats.Median, stats.Max, stats.ShortCodeCount)
	//	minAccumulator += stats.Min
	//	medianAccumulator += stats.Median
	//	maxAccumulator += stats.Max
	//	shortCodeCount += stats.ShortCodeCount
	//	collectionCount++
	//}
	//fmt.Printf("Aggregate min: %d\n", minAccumulator/collectionCount)
	//fmt.Printf("Aggregate median: %d\n", medianAccumulator/collectionCount)
	//fmt.Printf("Aggregate max: %d\n", maxAccumulator/collectionCount)
	//fmt.Printf("Total one-liner count across collections: %d\n", shortCodeCount)

	simpleTableLabel := "Collection"
	simpleTableColumnNames := []interface{}{"Collection", "Count"}
	simpleTableColumnWidths := []int{23, 15}
	PrintSimpleCountDataToConsole(simpleMap, simpleTableLabel, simpleTableColumnNames, simpleTableColumnWidths)

	//nestedOneLevelTableLabel := "Product Language"
	//nestedOneLevelTableColumnNames := []interface{}{"Language", "Count"}
	//nestedOneLevelTableColumnWidths := []int{20, 15}
	//PrintNestedOneLevelCountDataToConsole(nestedOneLevelMap, nestedOneLevelTableLabel, nestedOneLevelTableColumnNames, nestedOneLevelTableColumnWidths)
	//
	//nestedTwoLevelTableLabel := "Sub-Product Language"
	//nestedTwoLevelTableColumnNames := []interface{}{"Language", "Count"}
	//nestedTwoLevelTableColumnWidths := []int{20, 15}
	//PrintNestedTwoLevelCountDataToConsole(nestedTwoLevelMap, nestedTwoLevelTableLabel, nestedTwoLevelTableColumnNames, nestedTwoLevelTableColumnWidths)
}
