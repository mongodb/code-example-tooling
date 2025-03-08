package main

import (
	"log"
	"strings"
)

func ConvertPageIdToProductionUrl(pageId string, siteUrl string) string {
	parts := strings.Split(pageId, "/")
	pageUrl := ""
	// Check if the path has at least three parts to slice
	if len(parts) > 3 {
		// Omit the first three elements
		remainingParts := parts[3:]
		// Join the remaining parts back into a string with "/" separator
		result := strings.Join(remainingParts, "/")
		// Append the page path to the production site URL
		pageUrl = siteUrl + "/" + result
	} else {
		log.Println("The path", pageId, "does not have more than three parts to omit.")
	}
	return pageUrl
}
