package main

import (
	"snooty-api-parser/db"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

func CheckForExistingPage(collectionName string, incomingPage types.PageWrapper) *types.DocsPage {
	atlasDocId := utils.ConvertSnootyPageIdToAtlasPageId(incomingPage.Data.PageID)
	return db.GetAtlasPageData(collectionName, atlasDocId)
}
