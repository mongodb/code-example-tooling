package db

import (
	"common"
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
)

func BatchUpdateCollection(collectionName string, newPages []common.DocsPage, updatedPages []common.DocsPage, updatedSummaries common.CollectionReport) {
	uri := os.Getenv("MONGODB_URI")
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	var dbName = os.Getenv("DB_NAME")
	var ctx = context.Background()
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()
	// Define the database and collection
	db := client.Database(dbName)
	// If the collection doesn't exist already, we need to create it.
	CheckForAndCreateCollection(db, collectionName, ctx)
	// Make models to perform the updates as a bulk operation for the collection
	collection := db.Collection(collectionName)
	models := make([]mongo.WriteModel, 0)
	for _, newPage := range newPages {
		model := mongo.NewInsertOneModel().SetDocument(newPage)
		models = append(models, model)
	}
	for _, updatedPage := range updatedPages {
		filter := bson.D{{"_id", updatedPage.ID}}
		model := mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(updatedPage).SetUpsert(false)
		models = append(models, model)
	}
	summaryModel := mongo.NewReplaceOneModel().SetFilter(bson.D{{"_id", "summaries"}}).SetReplacement(updatedSummaries).SetUpsert(true)
	models = append(models, summaryModel)
	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(ctx, models, opts)
	if err != nil {
		log.Printf("Failed to perform bulk write for collection %s: %v", collectionName, err)
	}
	log.Printf("Atlas: For collection %s: Inserted %v documents, modified %v documents\n", collectionName, result.InsertedCount, result.ModifiedCount)
}
