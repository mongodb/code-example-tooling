package main

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	add_code_examples "snooty-api-parser/add-code-examples"
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
	newAppliedUsageExampleCount := 0
	var newCodeNodes []types.CodeNode
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
	report.Counter = IncrementProjectCountsForNewPage(incomingCodeNodeCount, incomingLiteralIncludeNodeCount, incomingIoCodeNodeCount, len(newCodeNodes), report.Counter)
	report = utils.ReportChanges(types.PageCreated, report, pageId)
	if len(newCodeNodes) > 0 {
		report = utils.ReportChanges(types.CodeExampleCreated, report, pageId, len(newCodeNodes))
	}

	if newAppliedUsageExampleCount > 0 {
		report.Counter.NewAppliedUsageExamplesCount += newAppliedUsageExampleCount
		report = utils.ReportChanges(types.AppliedUsageExampleAdded, report, pageId, newAppliedUsageExampleCount)
	}

	newCodeNodeCount := len(newCodeNodes)
	if incomingCodeNodeCount != newCodeNodeCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, pageId, incomingCodeNodeCount, newCodeNodeCount)
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
