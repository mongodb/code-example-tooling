package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
	"time"
)

func MakeNewDocsPage(data types.PageWrapper, siteUrl string, report types.ProjectReport, llm *ollama.LLM, ctx context.Context) (types.DocsPage, types.ProjectReport) {
	incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes := snooty.GetCodeExamplesFromIncomingData(data.Data.AST)
	incomingCodeNodeCount := len(incomingCodeNodes)
	incomingLiteralIncludeNodeCount := len(incomingLiteralIncludeNodes)
	incomingIoCodeNodeCount := len(incomingIoCodeBlockNodes)
	pageId := utils.ConvertSnootyPageIdToAtlasPageId(data.Data.PageID)
	pageUrl := utils.ConvertSnootyPageIdToProductionUrl(data.Data.PageID, siteUrl)
	product, subProduct := GetProductSubProduct(report.ProjectName, pageUrl)
	var isDriversProject bool
	if product == "Drivers" {
		isDriversProject = true
	} else {
		isDriversProject = false
	}
	var newCodeNodes []types.CodeNode
	for _, node := range incomingCodeNodes {
		newNode := snooty.MakeCodeNodeFromSnootyAST(node, llm, ctx, isDriversProject)
		newCodeNodes = append(newCodeNodes, newNode)
	}
	maybeKeywords := snooty.GetMetaKeywords(data.Data.AST.Children)

	languagesArrayValues := MakeLanguagesArray(newCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes)

	// Report relevant details for the new page
	report.Counter = IncrementProjectCountsForNewPage(incomingCodeNodeCount, incomingLiteralIncludeNodeCount, incomingIoCodeNodeCount, len(newCodeNodes), report.Counter)

	report.Counter.NewPagesCount += 1
	newPageChange := types.Change{
		Type: types.PageCreated,
		Data: fmt.Sprintf("Page ID: %s", pageId),
	}
	report.Changes = append(report.Changes, newPageChange)

	newCodeExamplesChange := types.Change{
		Type: types.PageCreated,
		Data: fmt.Sprintf("Page ID: %s, created %d new code examples", pageId, len(newCodeNodes)),
	}
	report.Changes = append(report.Changes, newCodeExamplesChange)

	newCodeNodeCount := len(newCodeNodes)
	if incomingCodeNodeCount != newCodeNodeCount {
		issue := types.Issue{
			Type: 1,
			Data: fmt.Sprintf("Page ID: %s, incoming code node count: %d, does not match new code node count: %d", pageId, incomingCodeNodeCount, newCodeNodeCount),
		}
		report.Issues = append(report.Issues, issue)
	}

	return types.DocsPage{
		ID:                   pageId,
		CodeNodesTotal:       incomingCodeNodeCount,
		DateAdded:            time.Now(),
		DateLastUpdated:      time.Now(),
		IoCodeBlocksTotal:    incomingIoCodeNodeCount,
		Languages:            languagesArrayValues,
		LiteralIncludesTotal: incomingLiteralIncludeNodeCount,
		Nodes:                &newCodeNodes,
		PageURL:              pageUrl,
		ProjectName:          report.ProjectName,
		Product:              product,
		SubProduct:           subProduct,
		Keywords:             maybeKeywords,
	}, report
}
