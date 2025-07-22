package updates

import (
	"common"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"regexp"
)

// AddProductNames adds a human-readable `product` and `sub_product` (where applicable) fields to documents with values that correspond
// to the Docs Taxonomy. If the document in the collection already has the applicable field(s), no change is made.
func AddProductNames(db *mongo.Database, ctx context.Context) {
	emptyFilter := bson.D{}
	collectionNames, err := db.ListCollectionNames(ctx, emptyFilter)

	if err != nil {
		log.Fatal("Could not retrieve collection names from the database: ", err)
	}

	for _, collectionName := range collectionNames {
		collection := db.Collection(collectionName)
		productInfo := common.GetProductInfo(collectionName)
		var update bson.D

		// Skip the "summaries" document because there should be no product or sub-product
		filter := bson.D{{"_id", bson.D{{"$ne", "summaries"}}}}

		switch productInfo.ProductType {
		case common.CollectionIsProduct:
			// If every document in the collection is a specific product with no sub-product (CollectionIsProduct),
			// set the "product" field for every document
			update = bson.D{{"$set", bson.D{{"product", productInfo.ProductName}}}}
		case common.CollectionIsSubProduct:
			// If every document in the collection is a sub-product, set both the product and sub-product field for
			// every document in the collection
			update = bson.D{
				{"$set", bson.D{
					{"product", productInfo.ProductName},
					{"sub_product", productInfo.SubProduct},
				}},
			}
		case common.DirSubProduct:
			// Should not be able to hit this case because none of the collection names map to DirSubProduct strings
			update = bson.D{}
		default:
			// If we hit this case, it's because we don't have a value matching the collection name, so we don't know
			// what "product" key to assign and therefore we don't want to make an update
			update = bson.D{}
		}

		updateResult, err := collection.UpdateMany(ctx, filter, update)
		if err != nil {
			log.Printf("Could not update documents to add a Product field in collection [%s]: %v", collectionName, err)
			continue
		}

		// Print a descriptive update depending on the type of operation we just completed
		switch productInfo.ProductType {
		case common.CollectionIsProduct:
			fmt.Printf("Added a Product field '%s' to %d documents in the [%s] collection\n", productInfo.ProductName, updateResult.ModifiedCount, collectionName)
		case common.CollectionIsSubProduct:
			fmt.Printf("Added a Product field '%s' and Sub-Product field '%s' to %d documents in the [%s] collection\n", productInfo.ProductName, productInfo.SubProduct, updateResult.ModifiedCount, collectionName)
		case common.DirSubProduct:
			fmt.Printf("Added a Product field '%s' to %d documents in the [%s] collection\n", productInfo.ProductName, updateResult.ModifiedCount, collectionName)
		default:
			fmt.Printf("Could not retrieve a matching ProductInfo for [%s] collection, so no updates were made.\n", collectionName)
		}

		// In the Atlas (cloud) docs, some of the docs subdirectories correspond to specific sub-products. Add the relevant
		// sub-product field to any page whose URL contains one of the relevant subdirectories
		if collectionName == "cloud-docs" {
			for _, dirPath := range common.SubProductDirs {
				subProductInfo := common.GetProductInfo(dirPath)

				dirPathFilter := bson.M{
					"page_url": bson.M{
						"$regex":   regexp.QuoteMeta(dirPath), // Properly escape special regex characters
						"$options": "i",                       // Case-insensitive match
					},
				}

				// Define the update to add a new field with the desired value
				dirSubProductUpdate := bson.M{
					"$set": bson.M{"sub_product": subProductInfo.SubProduct},
				}

				// Apply the update to all matching documents
				dirSubProductUpdateResult, err := collection.UpdateMany(ctx, dirPathFilter, dirSubProductUpdate)
				if err != nil {
					log.Printf("Could not update documents for path substring [%s]: %v", dirPath, err)
					continue
				}
				fmt.Printf("Added a 'sub_product' field value '%s' to %d documents in 'cloud-docs' for path substring [%s]\n", subProductInfo.SubProduct, dirSubProductUpdateResult.ModifiedCount, dirPath)
			}
		}
	}
}
