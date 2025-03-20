package utils

import (
	"log"
	"strings"
)

func ConvertSnootyPageIdToAtlasPageId(snootyPageId string) string {
	parts := strings.Split(snootyPageId, "/")
	atlasPageId := ""
	// Check if the path has at least three parts to slice
	if len(parts) > 3 {
		// Omit the first three elements
		remainingParts := parts[3:]
		// Join the remaining parts back into a string with "|" separator
		atlasPageId = strings.Join(remainingParts, "|")
	} else {
		log.Println("The Snooty 'page_id'", snootyPageId, "does not have more than three parts to omit.")
	}
	return atlasPageId
}
