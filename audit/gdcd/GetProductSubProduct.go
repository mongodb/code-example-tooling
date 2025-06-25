package main

import (
	"strings"
)

// GetProductSubProduct returns the product taxonomy for a given page in a project, which corresponds to collection in Atlas.
// It uses predefined mappings to determine the product and sub-product, if any, based on the project name and page URL.
func GetProductSubProduct(project string, page string) (string, string) {

	// Maps a project to a product name. Every project should have a corresponding product.
	// Keys are the project names (in alpha order), and values are product names (from Docs taxonomy).
	// This sets the `product` field for all documents in the project's collection in Atlas; otherwise, the field is left empty.
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
		"django":                   "Django Integration",
		"docs":                     "Server",
		"docs-k8s-operator":        "Enterprise Kubernetes Operator",
		"docs-relational-migrator": "Relational Migrator",
		"entity-framework":         "Entity Framework Core Provider", // DOCSP-50997 to add to taxonomy
		"golang":                   "Drivers",
		"java":                     "Drivers",
		"java-rs":                  "Drivers",
		"kafka-connector":          "Kafka Connector",
		"kotlin":                   "Drivers",
		"kotlin-sync":              "Drivers",
		"laravel":                  "Drivers",
		"mck":                      "Enterprise Kubernetes Operator",
		"mcp-server":               "MongoDB MCP Server", // DOCSP-50997 to add to taxonomy
		"mongoid":                  "Drivers",
		"mongodb-shell":            "MongoDB Shell",
		"mongocli":                 "MongoDB CLI",
		"node":                     "Drivers",
		"ops-manager":              "Ops Manager",
		"php-library":              "Drivers", // DOCSP-51020 to add to taxonomy/programmatic tagging
		"pymongo":                  "Drivers",
		"pymongo-arrow":            "Drivers",
		"ruby-driver":              "Drivers",
		"rust":                     "Drivers",
		"scala":                    "Drivers",
		"spark-connector":          "Spark Connector",
	}

	// Maps a project to a sub-product, when applicable.
	// Keys are the project names (in alpha order), and values are sub-product names (from Docs taxonomy).
	// This sets the `sub_product` field for all documents in the project's collection in Atlas; otherwise, the field is omitted.
	collectionSubProducts := map[string]string{
		"atlas-cli":      "Atlas CLI",
		"atlas-operator": "Kubernetes Operator",
		"charts":         "Charts",
	}

	// Maps a subdirectory in the `cloud-docs` project to a sub-product, when applicable.
	// Keys are the subdirectory names (in alpha order), and values are sub-product names (from Docs taxonomy).
	// This sets the `sub_product` field for documents in the `cloud-docs` collection in Atlas whose page URL contains the subdirectory string; otherwise, the field is omitted.
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
