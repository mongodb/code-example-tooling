package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
	"pull-audit-data/aggregations"
	"pull-audit-data/types"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	ctx := context.Background()
	db := client.Database("code_metrics")
	emptyFilter := bson.D{}
	collectionNames, err := db.ListCollectionNames(ctx, emptyFilter)
	if err != nil {
		panic(err)
	}

	categoryCountMap := make(map[string]int)
	langCountMap := make(map[string]int)
	lengthCountMap := make(map[string]types.CodeLengthStats)

	for _, collectionName := range collectionNames {
		categoryCountMap = aggregations.GetCategoryCounts(db, collectionName, categoryCountMap, ctx)
		langCountMap = aggregations.GetLanguageCounts(db, collectionName, langCountMap, ctx)
		lengthCountMap[collectionName] = aggregations.GetMinMedianMaxCodeLength(db, collectionName, ctx)
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
	for name, stats := range lengthCountMap {
		fmt.Printf("Collection: %s, Min: %d, Median: %d, Max: %d\n", name, stats.Min, stats.Median, stats.Max)
		minAccumulator += stats.Min
		medianAccumulator += stats.Median
		maxAccumulator += stats.Max
		collectionCount++
	}
	fmt.Printf("Aggregate min: %d\n", minAccumulator/collectionCount)
	fmt.Printf("Aggregate median: %d\n", medianAccumulator/collectionCount)
	fmt.Printf("Aggregate max: %d\n", maxAccumulator/collectionCount)
}
