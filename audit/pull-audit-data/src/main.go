package main

import (
	"context"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
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

	// To add product names to new docs pages
	//updates.AddProductNames(db, ctx)

	// To perform aggregations
	PerformAggregation(db, ctx)

	// To rename a field in the document
	//updates.RenameField(db, ctx)

	// To change the value of a field in the CodeNode object
	//updates.RenameValue(db, ctx)
}
