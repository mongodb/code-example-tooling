package main

import (
	"common"
	"gdcd/db"
	"gdcd/types"
	"gdcd/utils"
)

func CheckForExistingPage(collectionName string, incomingPage types.PageWrapper) *common.DocsPage {
	atlasDocId := utils.ConvertSnootyPageIdToAtlasPageId(incomingPage.Data.PageID)
	return db.GetAtlasPageData(collectionName, atlasDocId)
}
