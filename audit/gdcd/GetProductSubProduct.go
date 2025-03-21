package main

import (
	"strings"
)

func GetProductSubProduct(project string, page string) (string, string) {
	collectionProducts := map[string]string{
		"atlas-cli":             "Atlas",
		"atlas-operator":        "Atlas",
		"atlas-architecture":    "Atlas",
		"bi-connector":          "BI Connector",
		"c":                     "Drivers",
		"charts":                "Atlas",
		"cloud-docs":            "Atlas",
		"cloud-manager":         "Cloud Manager",
		"cloudgov":              "Atlas", // Missing from taxonomy/this is a guess
		"cluster-sync":          "Cluster-to-Cluster sync",
		"compass":               "Compass",
		"cpp-driver":            "Drivers",
		"csharp":                "Drivers",
		"database-tools":        "Database Tools",
		"docs":                  "Server",
		"docs-django":           "Drivers",
		"docs-entity-framework": "Drivers", // Missing from taxonomy/this is a guess
		"docs-golang":           "Drivers",
		"docs-java":             "Drivers",
		"docs-java-rs":          "Drivers",
		"docs-k8s-operator":     "Enterprise Kubernetes Operator",
		"kafka-connector":       "Kafka Connector",
		"kotlin":                "Drivers",
		"kotlin-sync":           "Drivers",
		"laravel":               "Drivers",
		"mongoid":               "Drivers",
		"mongodb-shell":         "MongoDB Shell",
		"mongocli":              "MongoDB CLI",
		"node":                  "Drivers",
		"ops-manager":           "Ops Manager",
		"php-library":           "Drivers", // Missing from taxonomy/this is a guess
		"pymongo":               "Drivers",
		"pymongo-arrow":         "Drivers",
		"relational-migrator":   "Relational Migrator",
		"ruby-driver":           "Drivers",
		"rust":                  "Drivers",
		"scala":                 "Drivers",
		"spark-connector":       "Spark Connector",
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
