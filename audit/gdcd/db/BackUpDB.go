package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func BackUpDb() {
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
	sourceDb := client.Database(dbName)

	// Create a db name for today's backup
	now := time.Now()
	// Format the date as "Month_Day"
	dateStr := fmt.Sprintf("%s_%d", now.Month(), now.Day())
	targetDBName := "backup_code_metrics_" + dateStr
	targetDb := client.Database(targetDBName)

	// List all collections in the source database
	collectionNames, err := sourceDb.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Fatalf("Error listing collections: %v", err)
	}

	log.Println("Backing up database...")
	// Iterate over each collection
	for _, collName := range collectionNames {
		sourceColl := sourceDb.Collection(collName)
		targetColl := targetDb.Collection(collName)
		// Fetch all documents from the source collection
		cursor, err := sourceColl.Find(ctx, bson.D{})
		if err != nil {
			log.Fatalf("Error finding documents in collection %s: %v", collName, err)
		}
		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err := cursor.Close(ctx)
			if err != nil {
				log.Fatalf("Error closing cursor: %v", err)
			}
		}(cursor, ctx)
		var documents []interface{}
		for cursor.Next(ctx) {
			var doc bson.M
			if err = cursor.Decode(&doc); err != nil {
				log.Fatalf("Error decoding document in collection %s: %v", collName, err)
			}
			documents = append(documents, doc)
		}
		if len(documents) > 0 {
			_, err = targetColl.InsertMany(ctx, documents)
			if err != nil {
				log.Fatalf("Error inserting documents into collection %s: %v", collName, err)
			}
			log.Printf("Copied %d documents to collection %s\n", len(documents), collName)
		}
	}
	log.Println("Successfully backed up database")

	// Drop the oldest backup. Get a list of backup names so we can find the oldest backup.
	backupNames := getBackupDbNames(client, ctx)
	oldestBackup := findOldestBackup(backupNames)
	// Get a handle for the database
	dbToDrop := client.Database(oldestBackup)

	// Drop the database
	err = dbToDrop.Drop(ctx)
	if err != nil {
		log.Fatalf("Failed to drop database %v: %v", oldestBackup, err)
	}
	log.Printf("Oldest backup database '%s' dropped successfully\n", oldestBackup)
}

// The cluster contains a mix of databases - some are backups, and some are other databases.
// We want to get a slice of only the backup database names.
func getBackupDbNames(client *mongo.Client, ctx context.Context) []string {
	var backupNames []string
	// List the database names in the cluster
	databaseNames, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		log.Fatalf("Failed to list database names: %v", err)
	}
	// Get only the DB names for the backup databases
	for _, databaseName := range databaseNames {
		if strings.HasPrefix(databaseName, "backup_code_metrics") {
			backupNames = append(backupNames, databaseName)
		}
	}
	return backupNames
}

// Parse the dates from the backup database names to find the oldest backup database.
func findOldestBackup(backupNames []string) string {
	// Define a reference year (we need a year to work with Go's time package)
	const year = 2025 // Arbitrary year for comparison purposes

	// Variables to track the oldest date and its corresponding string
	var oldestDate time.Time
	var oldestBackupName string

	// Iterate over the strings to extract and compare dates
	for _, entry := range backupNames {
		// Split the string and find the month and day (assume the format is fixed)
		parts := strings.Split(entry, "_")
		if len(parts) < 4 {
			continue // Skip invalid strings
		}
		monthStr := parts[len(parts)-2] // Second-to-last part is the month
		dayStr := parts[len(parts)-1]   // Last part is the day

		// Convert the day string to an integer
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			fmt.Println("Error converting day:", err)
			continue
		}

		// Parse the month using the time.Month enum
		month, err := parseMonth(monthStr)
		if err != nil {
			fmt.Println("Error parsing month:", err)
			continue
		}

		// Create a time.Time object for the given month and day
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

		// Compare dates to find the earliest one
		if oldestBackupName == "" || date.Before(oldestDate) {
			oldestDate = date
			oldestBackupName = entry
		}
	}
	return oldestBackupName
}

// Helper function to parse month names into time.Month
func parseMonth(month string) (time.Month, error) {
	month = strings.ToLower(month) // Make the string case-insensitive
	switch month {
	case "january":
		return time.January, nil
	case "february":
		return time.February, nil
	case "march":
		return time.March, nil
	case "april":
		return time.April, nil
	case "may":
		return time.May, nil
	case "june":
		return time.June, nil
	case "july":
		return time.July, nil
	case "august":
		return time.August, nil
	case "september":
		return time.September, nil
	case "october":
		return time.October, nil
	case "november":
		return time.November, nil
	case "december":
		return time.December, nil
	default:
		return time.Month(0), fmt.Errorf("invalid month: %s", month)
	}
}
