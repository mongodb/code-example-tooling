package db

import (
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

func HandleMissingPageIds(collectionName string, incomingPageIds map[string]bool, report types.ProjectReport) (int, types.ProjectReport) {
	// Get a slice of all the page IDs for pages that are currently in Atlas
	existingPageIds := GetAtlasPageIDs(collectionName)
	var missingPageIds []string
	if existingPageIds != nil {
		// Compare the pages that are currently in Atlas with pages coming in from the Snooty Data API. If the page exists
		// in Atlas but isn't coming in from the Snooty Data API, grab the ID so we can remove the page in Atlas.
		for _, existingId := range existingPageIds {
			if !incomingPageIds[existingId] {
				missingPageIds = append(missingPageIds, existingId)
			}
		}
	}
	for _, missingPageId := range missingPageIds {
		pageRemoved := RemovePageFromAtlas(collectionName, missingPageId)
		if pageRemoved {
			report = utils.ReportChanges(types.PageRemoved, report, missingPageId)
		} else {
			report = utils.ReportIssues(types.PageNotRemovedIssue, report, missingPageId)
		}
	}
	return len(existingPageIds), report
}
