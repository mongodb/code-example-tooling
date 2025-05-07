package main

import (
	"context"
	"dodec/aggregations"
	"dodec/types"
	"dodec/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// PerformAggregation executes several different aggregation operations for every collection in the DB, and logs the output to console.
func PerformAggregation(db *mongo.Database, ctx context.Context) {
	// The aggregations in this project use one of these data structures. Uncomment the corresponding data structure,
	// or make duplicates with appropriate names as needed
	//simpleMap := make(map[string]int)
	//codeLengthMap := make(map[string]types.CodeLengthStats)
	//nestedOneLevelMap := make(map[string]map[string]int)
	//nestedTwoLevelMap := make(map[string]map[string]map[string]int)
	//pageIdChangesCountMap := make(map[string][]types.PageIdChangedCounts)
	//pageIdsWithNodeLangCountMismatch := make(map[string][]string)
	productSubProductCounter := types.NewAppliedUsageExampleCounterByProductSubProduct{
		ProductSubProductCounts: make(map[string]map[string]int),
		ProductAggregateCount:   make(map[string]int),
		PagesInCollections:      make(map[string][]types.PageIdNewAppliedUsageExamples),
	}

	// If you just need to get data for a single collection, perform the aggregation using the collection name
	//simpleMap = aggregations.GetLanguageCounts(db, "pymongo", simpleMap, ctx)

	// If you need to get data across all the collections in the `code-examples` database, iterate through the collections
	emptyFilter := bson.D{}
	collectionNames, err := db.ListCollectionNames(ctx, emptyFilter)
	if err != nil {
		panic(err)
	}
	//substringToFindInCodeExamples := "defaultauthdb"

	for _, collectionName := range collectionNames {
		//simpleMap = aggregations.GetCategoryCounts(db, collectionName, simpleMap, ctx)
		//simpleMap = aggregations.GetLanguageCounts(db, collectionName, simpleMap, ctx)
		//simpleMap = aggregations.GetStringInCodeNodeCounts(db, collectionName, simpleMap, ctx, substringToFindInCodeExamples)
		//simpleMap = aggregations.GetLangCountsFromNodes(db, collectionName, simpleMap, ctx)
		//simpleMap = aggregations.GetLangCountFromLangArrayManually(db, collectionName, simpleMap, ctx)
		//codeLengthMap = aggregations.GetCodeLengths(db, collectionName, codeLengthMap, ctx)
		//nestedOneLevelMap = aggregations.GetCategoryLanguageCounts(db, collectionName, nestedOneLevelMap, ctx)
		//nestedOneLevelMap = aggregations.GetProductCategoryCounts(db, collectionName, nestedOneLevelMap, ctx)
		//nestedTwoLevelMap = aggregations.GetSubProductCategoryCounts(db, collectionName, nestedTwoLevelMap, ctx)
		//simpleMap = aggregations.GetOneLineUsageExampleCounts(db, collectionName, simpleMap, ctx)
		//nestedOneLevelMap = aggregations.GetProductLanguageCounts(db, collectionName, nestedOneLevelMap, ctx)
		//nestedTwoLevelMap = aggregations.GetSubProductLanguageCounts(db, collectionName, nestedTwoLevelMap, ctx)
		//simpleMap = aggregations.GetCollectionCount(db, collectionName, simpleMap, ctx)
		//simpleMap = aggregations.GetSpecificCategoryByProduct(db, collectionName, common.UsageExample, simpleMap, ctx)
		//langCount := aggregations.GetSpecificLanguageCount(db, collectionName, common.Go, ctx)
		//pageIdChangesCountMap = aggregations.GetDocsIdsWithRecentActivity(db, collectionName, pageIdChangesCountMap, ctx)
		//pageIdsWithNodeLangCountMismatch = aggregations.GetPagesWithNodeLangCountMismatch(db, collectionName, pageIdsWithNodeLangCountMismatch, ctx)
		//pageIdsWithNodeLangCountMismatch = aggregations.FindDocsMissingProduct(db, collectionName, pageIdsWithNodeLangCountMismatch, ctx)
		productSubProductCounter = aggregations.FindNewAppliedUsageExamples(db, collectionName, productSubProductCounter, ctx)
	}

	//simpleTableLabel := "Collection"
	//simpleTableColumnNames := []interface{}{"Collection", "Count"}
	//simpleTableColumnWidths := []int{30, 15}
	//utils.PrintSimpleCountDataToConsole(simpleMap, simpleTableLabel, simpleTableColumnNames, simpleTableColumnWidths)

	//nestedOneLevelTableLabel := "Product Language"
	//nestedOneLevelTableColumnNames := []interface{}{"Language", "Count"}
	//nestedOneLevelTableColumnWidths := []int{20, 15}
	//utils.PrintNestedOneLevelCountDataToConsole(nestedOneLevelMap, nestedOneLevelTableLabel, nestedOneLevelTableColumnNames, nestedOneLevelTableColumnWidths)

	//nestedTwoLevelTableLabel := "Sub-Product Language"
	//nestedTwoLevelTableColumnNames := []interface{}{"Language", "Count"}
	//nestedTwoLevelTableColumnWidths := []int{20, 15}
	//utils.PrintNestedTwoLevelCountDataToConsole(nestedTwoLevelMap, nestedTwoLevelTableLabel, nestedTwoLevelTableColumnNames, nestedTwoLevelTableColumnWidths)

	// The length count map is a very specific fixed data structure, so this function has hard-coded title and column names/widths
	//utils.PrintCodeLengthMapToConsole(codeLengthMap)
	//utils.PrintPageIdChangesCountMap(pageIdChangesCountMap)
	//utils.PrintPageIdsWithNodeLangCountMismatch(pageIdsWithNodeLangCountMismatch)
	utils.PrintPageIdNewAppliedUsageExampleCounts(productSubProductCounter)
}
