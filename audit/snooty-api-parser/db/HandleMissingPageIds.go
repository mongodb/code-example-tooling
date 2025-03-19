package db

func HandleMissingPageIds(collectionName string, incomingPageIds map[string]bool) int {
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
		RemovePageFromAtlas(collectionName, missingPageId)
	}
	return len(existingPageIds)
}
