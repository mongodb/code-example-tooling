package main

import (
	"snooty-api-parser/db"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

// HandleDeletedIncomingPages checks whether a page that has the `"deleted":true` flag when it comes in from the Snooty Data API
// has a corresponding page in Atlas. If it does, we delete it.
func HandleDeletedIncomingPages(collectionName string, deletedPage types.PageWrapper) {
	maybeAtlasId := utils.ConvertSnootyPageIdToAtlasPageId(deletedPage.Data.PageID)
	maybeAtlasDocument := db.GetAtlasPageData(collectionName, maybeAtlasId)
	if maybeAtlasDocument != nil {
		db.RemovePageFromAtlas(collectionName, maybeAtlasDocument.ID)
	}
}
