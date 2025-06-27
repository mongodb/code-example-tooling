package utils

import (
	"strings"
)

func ConvertAtlasPageIdToProductionUrl(pageId string, siteUrl string) string {
	// The Atlas ID has `|`-separated segments. Replace with `/` to use it in a URL.
	pageIdAsUrlSegments := strings.ReplaceAll(pageId, "|", "/")
	// Append the page path to the production site URL
	return siteUrl + "/" + pageIdAsUrlSegments
}
