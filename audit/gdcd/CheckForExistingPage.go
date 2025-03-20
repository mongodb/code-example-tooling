package main

import (
	"gdcd/db"
	"gdcd/types"
	"gdcd/utils"
)

func CheckForExistingPage(collectionName string, incomingPage types.PageWrapper) *types.DocsPage {
	atlasDocId := utils.ConvertSnootyPageIdToAtlasPageId(incomingPage.Data.PageID)
	return db.GetAtlasPageData(collectionName, atlasDocId)
}
