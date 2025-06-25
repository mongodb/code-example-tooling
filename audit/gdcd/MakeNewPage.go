package main

import (
	"common"
	"context"
	add_code_examples "gdcd/add-code-examples"
	"gdcd/snooty"
	"gdcd/types"
	"gdcd/utils"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

func MakeNewPage(data types.PageWrapper, siteUrl string, report types.ProjectReport, llm *ollama.LLM, ctx context.Context) (common.DocsPage, types.ProjectReport) {
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
	newAppliedUsageExampleCount := 0
	var newCodeNodes []common.CodeNode
	for _, node := range incomingCodeNodes {
		newNode := snooty.MakeCodeNodeFromSnootyAST(node, llm, ctx, isDriversProject)
		newCodeNodes = append(newCodeNodes, newNode)
		if add_code_examples.IsNewAppliedUsageExample(newNode) {
			newAppliedUsageExampleCount++
		}
	}
	maybeKeywords := snooty.GetMetaKeywords(data.Data.AST.Children)

	languagesArrayValues := MakeLanguagesArray(newCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes)

	// Report relevant details for the new page
	report = UpdateProjectReportForNewPage(incomingCodeNodeCount, incomingLiteralIncludeNodeCount, incomingIoCodeNodeCount, len(newCodeNodes), newAppliedUsageExampleCount, pageId, report)

	return common.DocsPage{
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
