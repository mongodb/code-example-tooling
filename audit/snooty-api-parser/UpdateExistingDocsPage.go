package main

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/add-code-examples"
	"snooty-api-parser/compare-code-examples"
	"snooty-api-parser/db"
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
	"time"
)

func UpdateExistingDocsPage(existingPage types.DocsPage, data types.PageWrapper, projectReport types.ProjectReport, llm *ollama.LLM, ctx context.Context) (*types.DocsPage, types.ProjectReport) {
	var existingCurrentCodeNodes []types.CodeNode
	var existingRemovedCodeNodes []types.CodeNode
	// Some of the existing Nodes on the page could have been previously removed from the page. So we need to know which
	// nodes are "currently" on the page, and which nodes have already been removed. The ones that are "currently" on the
	// page should be used to compare code examples, but the ones that have already been removed from the page will be
	// appended to the Nodes array without changes after making all the other updates.
	if existingPage.Nodes != nil {
		existingCurrentCodeNodes, existingRemovedCodeNodes = db.GetCurrentRemovedAtlasCodeNodes(*existingPage.Nodes)
	}
	atlasDocCurrentCodeNodeCount := len(existingCurrentCodeNodes)
	incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes := snooty.GetCodeExamplesFromIncomingData(data.Data.AST)
	maybePageKeywords := snooty.GetMetaKeywords(data.Data.AST.Children)
	newAppliedUsageExampleCount := 0
	incomingCodeNodePageCount := len(incomingCodeNodes)
	incomingLiteralIncludeNodeCount := len(incomingLiteralIncludeNodes)
	incomingIoCodeBlockNodeCount := len(incomingIoCodeBlockNodes)
	projectReport = IncrementProjectCountsForExistingPage(incomingCodeNodePageCount, incomingLiteralIncludeNodeCount, incomingIoCodeBlockNodeCount, existingPage, projectReport)
	var pageWithUpdatedKeywords *types.DocsPage
	if len(maybePageKeywords) > 0 {
		// If the page has keywords, and it's not the same number of keywords that are coming in from Snooty, update the keywords
		if len(existingPage.Keywords) != len(maybePageKeywords) {
			pageWithUpdatedKeywords = &existingPage
			pageWithUpdatedKeywords.Keywords = maybePageKeywords
			pageWithUpdatedKeywords.DateLastUpdated = time.Now()
			projectReport = utils.ReportChanges(types.KeywordsUpdated, projectReport, existingPage.ID)
		}
	}

	if incomingCodeNodePageCount == atlasDocCurrentCodeNodeCount {
		// The page doesn't have any code changes we can return a page with updated keywords (if it exists) and an updated projectReport
		projectReport.Counter.UnchangedCodeNodesCount += atlasDocCurrentCodeNodeCount
		return pageWithUpdatedKeywords, projectReport
	}

	// If the incoming page node count does not equal the existing atlas doc node count, we need to update the page
	var updatedDocsPage types.DocsPage
	if pageWithUpdatedKeywords != nil {
		updatedDocsPage = *pageWithUpdatedKeywords
	} else {
		updatedDocsPage = existingPage
	}
	var isDriversProject bool
	if existingPage.Product == "Drivers" {
		isDriversProject = true
	} else {
		isDriversProject = false
	}

	// If examples exist already and we are getting no incoming examples from the API, the existing examples have been removed from the incoming page
	if existingCurrentCodeNodes != nil && incomingCodeNodePageCount == 0 {
		newRemovedNodeCount := len(existingCurrentCodeNodes)
		// Mark all nodes as removed
		updatedCodeNodes := make([]types.CodeNode, 0)
		for _, node := range existingCurrentCodeNodes {
			node.DateRemoved = time.Now()
			node.IsRemoved = true
			updatedCodeNodes = append(updatedCodeNodes, node)
		}
		// Some removed nodes may already exist on the page. We don't want to count those in the "new removed nodes" count,
		// but we do need to add them to the `Nodes` array if we don't want them to disappear.
		if existingRemovedCodeNodes != nil && len(existingRemovedCodeNodes) > 0 {
			for _, node := range existingRemovedCodeNodes {
				updatedCodeNodes = append(updatedCodeNodes, node)
			}
		}
		updatedDocsPage.Nodes = &updatedCodeNodes

		oldCodeNodeCount := existingPage.CodeNodesTotal
		oldLiteralIncludeCount := existingPage.LiteralIncludesTotal
		oldIoCodeBlockCount := existingPage.IoCodeBlocksTotal

		// Update the code node count, io-block-count and literalinclude count
		updatedDocsPage.CodeNodesTotal = 0
		updatedDocsPage.LiteralIncludesTotal = 0
		updatedDocsPage.IoCodeBlocksTotal = 0

		// Update the language counts array (set all values for the page to 0)
		updatedLanguagesArray := MakeLanguagesArray([]types.CodeNode{}, []types.ASTNode{}, []types.ASTNode{})
		updatedDocsPage.Languages = updatedLanguagesArray

		// Update the date_last_updated time
		updatedDocsPage.DateLastUpdated = time.Now()

		// Add relevant entries to the projectReport
		projectReport = utils.ReportChanges(types.PageUpdated, projectReport, existingPage.ID)

		if newRemovedNodeCount > 0 {
			projectReport.Counter.RemovedCodeNodesCount += newRemovedNodeCount
			projectReport = utils.ReportChanges(types.CodeExampleRemoved, projectReport, existingPage.ID, newRemovedNodeCount)
		}

		if oldCodeNodeCount != incomingCodeNodePageCount {
			projectReport = utils.ReportChanges(types.CodeNodeCountChange, projectReport, existingPage.ID, oldCodeNodeCount, incomingCodeNodePageCount)
		}
		if oldLiteralIncludeCount != incomingLiteralIncludeNodeCount {
			projectReport = utils.ReportChanges(types.LiteralIncludeCountChange, projectReport, existingPage.ID, oldLiteralIncludeCount, incomingLiteralIncludeNodeCount)
		}
		if oldIoCodeBlockCount != incomingIoCodeBlockNodeCount {
			projectReport = utils.ReportChanges(types.IoCodeBlockCountChange, projectReport, existingPage.ID, oldIoCodeBlockCount, incomingIoCodeBlockNodeCount)
		}
	} else if atlasDocCurrentCodeNodeCount == 0 && incomingCodeNodePageCount > 0 {
		// There are no existing code examples - they're all new - so just make new code examples
		newCodeNodes := make([]types.CodeNode, 0)
		for _, snootyNode := range incomingCodeNodes {
			newNode := snooty.MakeCodeNodeFromSnootyAST(snootyNode, llm, ctx, isDriversProject)
			newCodeNodes = append(newCodeNodes, newNode)
			if add_code_examples.IsNewAppliedUsageExample(newNode) {
				newAppliedUsageExampleCount++
			}
		}
		newCodeNodeCount := len(newCodeNodes)
		updatedDocsPage.Nodes = &newCodeNodes

		// Update the code node count, io-block-count and literalinclude count
		updatedDocsPage.CodeNodesTotal = newCodeNodeCount
		updatedDocsPage.LiteralIncludesTotal = len(incomingLiteralIncludeNodes)
		updatedDocsPage.IoCodeBlocksTotal = len(incomingIoCodeBlockNodes)

		// Add language counts
		updatedLanguagesArray := MakeLanguagesArray(newCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes)
		updatedDocsPage.Languages = updatedLanguagesArray

		// Update the date_last_updated time
		updatedDocsPage.DateLastUpdated = time.Now()

		// Add relevant entries to the project projectReport
		projectReport = utils.ReportChanges(types.PageUpdated, projectReport, existingPage.ID)
		if newCodeNodeCount > 0 {
			projectReport.Counter.NewCodeNodesCount += newCodeNodeCount
			projectReport = utils.ReportChanges(types.CodeExampleCreated, projectReport, existingPage.ID, newCodeNodeCount)
		}
		if newAppliedUsageExampleCount > 0 {
			projectReport.Counter.NewAppliedUsageExamplesCount += newAppliedUsageExampleCount
			projectReport = utils.ReportChanges(types.AppliedUsageExampleAdded, projectReport, existingPage.ID, newAppliedUsageExampleCount)
		}
	} else if atlasDocCurrentCodeNodeCount == 0 && incomingCodeNodePageCount == 0 {
		// No code examples to deal with here - just return nil and the unchanged projectReport
		return nil, projectReport
	} else {
		// Add an entry to the projectReport for updating the page. Adding this first so it precedes individual changes.
		// Note we're not reporting on any changes here - any count changes are reported through
		// CompareExistingIncomingCodeExampleSlices()
		projectReport = utils.ReportChanges(types.PageUpdated, projectReport, existingPage.ID)

		// If some examples exist already, and some examples are coming in from snooty, they might be updated, new, removed, or unchanged.
		// Handle those distinct cases.
		var updatedCodeNodes []types.CodeNode
		updatedCodeNodes, projectReport = compare_code_examples.CompareExistingIncomingCodeExampleSlices(existingCurrentCodeNodes, existingRemovedCodeNodes, incomingCodeNodes, projectReport, existingPage.ID, llm, ctx, isDriversProject)
		updatedDocsPage.Nodes = &updatedCodeNodes

		// Update the code node count, io-block-count and literalinclude count
		updatedDocsPage.CodeNodesTotal = incomingCodeNodePageCount
		updatedDocsPage.LiteralIncludesTotal = len(incomingLiteralIncludeNodes)
		updatedDocsPage.IoCodeBlocksTotal = len(incomingIoCodeBlockNodes)

		// Update the language counts for the page based on the updated code nodes.
		updatedLanguagesArray := MakeLanguagesArray(updatedCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes)
		updatedDocsPage.Languages = updatedLanguagesArray

		// Update the date_last_updated time
		updatedDocsPage.DateLastUpdated = time.Now()

		// Update the projectReport for changes to the code node count, literalinclude count, or io-code-block count
		oldCodeNodeCount := existingPage.CodeNodesTotal
		oldLiteralIncludeCount := existingPage.LiteralIncludesTotal
		oldIoCodeBlockCount := existingPage.IoCodeBlocksTotal
		if oldCodeNodeCount != incomingCodeNodePageCount {
			projectReport = utils.ReportChanges(types.CodeNodeCountChange, projectReport, existingPage.ID, oldCodeNodeCount, incomingCodeNodePageCount)
		}
		if oldLiteralIncludeCount != incomingLiteralIncludeNodeCount {
			projectReport = utils.ReportChanges(types.LiteralIncludeCountChange, projectReport, existingPage.ID, oldLiteralIncludeCount, incomingLiteralIncludeNodeCount)
		}
		if oldIoCodeBlockCount != incomingIoCodeBlockNodeCount {
			projectReport = utils.ReportChanges(types.IoCodeBlockCountChange, projectReport, existingPage.ID, oldIoCodeBlockCount, incomingIoCodeBlockNodeCount)
		}
	}
	return &updatedDocsPage, projectReport
}
