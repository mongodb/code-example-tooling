package main

import (
	"strings"
)

func GetProductSubProduct(project string, page string) (string, string) {
	collectionProducts := map[string]string{
		"atlas-cli":                "Atlas",
		"atlas-operator":           "Atlas",
		"atlas-architecture":       "Atlas Architecture Center",
		"bi-connector":             "BI Connector",
		"c":                        "Drivers",
		"charts":                   "Atlas",
		"cloud-docs":               "Atlas",
		"cloud-manager":            "Cloud Manager",
		"cloudgov":                 "Atlas", // Missing from taxonomy/this is a guess
		"cluster-sync":             "Cluster-to-Cluster sync",
		"compass":                  "Compass",
		"cpp-driver":               "Drivers",
		"csharp":                   "Drivers",
		"database-tools":           "Database Tools",
		"django":                   "Drivers",
		"docs":                     "Server",
		"docs-k8s-operator":        "Enterprise Kubernetes Operator",
		"docs-relational-migrator": "Relational Migrator",
		"entity-framework":         "Drivers", // Missing from taxonomy/this is a guess
		"golang":                   "Drivers",
		"java":                     "Drivers",
		"java-rs":                  "Drivers",
		"kafka-connector":          "Kafka Connector",
		"kotlin":                   "Drivers",
		"kotlin-sync":              "Drivers",
		"laravel":                  "Drivers",
		"mck":                      "Enterprise Kubernetes Operator",
		"mongoid":                  "Drivers",
		"mongodb-shell":            "MongoDB Shell",
		"mongocli":                 "MongoDB CLI",
		"node":                     "Drivers",
		"ops-manager":              "Ops Manager",
		"php-library":              "Drivers", // Missing from taxonomy/this is a guess
		"pymongo":                  "Drivers",
		"pymongo-arrow":            "Drivers",
		"ruby-driver":              "Drivers",
		"rust":                     "Drivers",
		"scala":                    "Drivers",
		"spark-connector":          "Spark Connector",
	}

	collectionSubProducts := map[string]string{
		"atlas-cli":      "Atlas CLI",
		"atlas-operator": "Kubernetes Operator",
		"charts":         "Charts",
	}

	atlasCollectionSubProductByDir := map[string]string{
		"atlas-stream-processing": "Stream Processing",
		"atlas-search":            "Search",
		"atlas-vector-search":     "Vector Search",
		"data-federation":         "Data Federation",
		"online-archive":          "Online Archive",
		"triggers":                "Triggers",
	}

	product := collectionProducts[project]
	subProduct := ""

	if project == "cloud-docs" {
		for directory, displayName := range atlasCollectionSubProductByDir {
			if strings.Contains(page, directory) {
				subProduct = displayName
			}
		}
	} else {
		// If it's a collection that directly correlates to one of the collectionSubProducts, it's an Atlas product
		// and has the relevant sub-product display name
		displayName, exists := collectionSubProducts[project]
		if exists {
			product = "Atlas"
			subProduct = displayName
			return product, subProduct
		}
	}
	return product, subProduct
}
