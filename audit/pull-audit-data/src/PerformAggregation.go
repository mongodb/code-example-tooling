package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"pull-audit-data/aggregations"
	"pull-audit-data/types"
)

// PerformAggregation executes several different aggregation operations for every collection in the DB, and logs the output to console.
func PerformAggregation(db *mongo.Database, ctx context.Context) {
	emptyFilter := bson.D{}
	collectionNames, err := db.ListCollectionNames(ctx, emptyFilter)
	if err != nil {
		panic(err)
	}
	categoryCountMap := make(map[string]int)
	langCountMap := make(map[string]int)
	lengthCountMap := make(map[string]types.CodeLengthStats)
	//productCategoryMap := make(map[string]map[string]int)
	//subProductCategoryMap := make(map[string]map[string]map[string]int)
	//productLanguageMap := make(map[string]map[string]int)
	//subProductLanguageMap := make(map[string]map[string]map[string]int)
	oneLinerCountMap := make(map[string]int)
	totalOneLinerCount := 0

	for _, collectionName := range collectionNames {
		categoryCountMap = aggregations.GetCategoryCounts(db, collectionName, categoryCountMap, ctx)
		langCountMap = aggregations.GetLanguageCounts(db, collectionName, langCountMap, ctx)
		lengthCountMap[collectionName] = aggregations.GetMinMedianMaxCodeLength(db, collectionName, ctx)
		//productCategoryMap = aggregations.GetProductCategoryCounts(db, collectionName, productCategoryMap, ctx)
		//subProductCategoryMap = aggregations.GetSubProductCategoryCounts(db, collectionName, subProductCategoryMap, ctx)
		oneLinerCountMap, totalOneLinerCount = aggregations.GetOneLinerCounts(db, collectionName, oneLinerCountMap, totalOneLinerCount, ctx)
		//productLanguageMap = aggregations.GetProductLanguageCounts(db, collectionName, productLanguageMap, ctx)
		//subProductLanguageMap = aggregations.GetSubProductLanguageCounts(db, collectionName, subProductLanguageMap, ctx)
	}

	for category, count := range categoryCountMap {
		fmt.Printf("%s: %d\n", category, count)
	}
	for language, count := range langCountMap {
		fmt.Printf("%s: %d\n", language, count)
	}
	minAccumulator := 0
	medianAccumulator := 0
	maxAccumulator := 0
	collectionCount := 0
	shortCodeCount := 0
	for name, stats := range lengthCountMap {
		fmt.Printf("Collection: %s, Min: %d, Median: %d, Max: %d, One-Liners: %d\n", name, stats.Min, stats.Median, stats.Max, stats.ShortCodeCount)
		minAccumulator += stats.Min
		medianAccumulator += stats.Median
		maxAccumulator += stats.Max
		shortCodeCount += stats.ShortCodeCount
		collectionCount++
	}
	fmt.Printf("Aggregate min: %d\n", minAccumulator/collectionCount)
	fmt.Printf("Aggregate median: %d\n", medianAccumulator/collectionCount)
	fmt.Printf("Aggregate max: %d\n", maxAccumulator/collectionCount)
	fmt.Printf("Total one-liner count across collections: %d\n", shortCodeCount)

	for name, count := range oneLinerCountMap {
		fmt.Printf("Collection: %s, One-Liner Usage Examples: %d\n", name, count)
	}
	fmt.Printf("Total one liner count across collections: %d\n", totalOneLinerCount)

	//PrintProductCategoryData(productCategoryMap)
	//PrintSubProductCategoryData(subProductCategoryMap)
	//PrintProductLanguageData(productLanguageMap)
	//PrintSubProductLanguageData(subProductLanguageMap)
}
