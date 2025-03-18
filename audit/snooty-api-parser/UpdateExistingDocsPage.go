package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/compare-code-examples"
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
	"time"
)

func UpdateExistingDocsPage(existingPage types.DocsPage, data types.PageWrapper, report types.ProjectReport, llm *ollama.LLM, ctx context.Context) (*types.DocsPage, types.ProjectReport) {
	atlasDocCodeNodeCount := existingPage.CodeNodesTotal
	incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes := snooty.GetCodeExamplesFromIncomingData(data.Data.AST)
	incomingCodeNodePageCount := len(incomingCodeNodes)
	incomingLiteralIncludeNodeCount := len(incomingLiteralIncludeNodes)
	incomingIoCodeBlockNodeCount := len(incomingIoCodeBlockNodes)
	report = IncrementProjectCountsForExistingPage(incomingCodeNodePageCount, incomingLiteralIncludeNodeCount, incomingIoCodeBlockNodeCount, existingPage, report)
	if incomingCodeNodePageCount == atlasDocCodeNodeCount {
		// TODO: update keywords for the page even if there are no code nodes changes (should also update date updated date)
		// TODO: Add new change type for keywords update
		// The page doesn't have any changes - don't bother returning the page, but do return the updated project counter
		report.Counter.UnchangedCodeNodesCount += atlasDocCodeNodeCount
		return nil, report
	}

	// If the incoming page node count does not equal the existing atlas doc node count, we need to update the page
	updatedDocsPage := existingPage
	var isDriversProject bool
	if existingPage.Product == "Drivers" {
		isDriversProject = true
	} else {
		isDriversProject = false
	}
	updatedDocsPage.Keywords = snooty.GetMetaKeywords(data.Data.AST.Children)

	// If examples exist already and we are getting no incoming examples from the API, the existing examples have been removed from the incoming page
	if existingPage.Nodes != nil && incomingCodeNodePageCount == 0 {
		removedNodeCount := len(*existingPage.Nodes)
		// Mark all nodes as removed
		updatedCodeNodes := make([]types.CodeNode, 0)
		for _, node := range *existingPage.Nodes {
			node.DateRemoved = time.Now()
			node.IsRemoved = true
			updatedCodeNodes = append(updatedCodeNodes, node)
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

		// Add relevant entries to the project report
		pageUpdatedChange := types.Change{
			Type: types.PageUpdated,
			Data: fmt.Sprintf("Page ID: %s", existingPage.ID),
		}
		report.Changes = append(report.Changes, pageUpdatedChange)

		if removedNodeCount > 0 {
			report.Counter.RemovedCodeNodesCount += removedNodeCount
			removedExamplesChange := types.Change{
				Type: types.CodeExampleRemoved,
				Data: fmt.Sprintf("Page ID: %s, removed %d examples, now %d", existingPage.ID, atlasDocCodeNodeCount, incomingCodeNodePageCount),
			}
			report.Changes = append(report.Changes, removedExamplesChange)
		}

		if oldCodeNodeCount != 0 {
			codeNodeCountChange := types.Change{
				Type: types.CodeNodeCountChange,
				Data: fmt.Sprintf("Page ID: %s, code node count was: %d, now 0", existingPage.ID, oldCodeNodeCount),
			}
			report.Changes = append(report.Changes, codeNodeCountChange)
		}
		if oldLiteralIncludeCount != 0 {
			literalIncludeCountChange := types.Change{
				Type: types.LiteralIncludeCountChange,
				Data: fmt.Sprintf("Page ID: %s, literalinclude count was %d, now 0", existingPage.ID, oldLiteralIncludeCount),
			}
			report.Changes = append(report.Changes, literalIncludeCountChange)
		}
		if oldIoCodeBlockCount != 0 {
			ioCodeBlockCountChange := types.Change{
				Type: types.IoCodeBlockCountChange,
				Data: fmt.Sprintf("Page ID: %s, io-code-block count was %d, now 0", existingPage.ID, oldIoCodeBlockCount),
			}
			report.Changes = append(report.Changes, ioCodeBlockCountChange)
		}

	} else if existingPage.Nodes == nil && incomingCodeNodePageCount > 0 {
		// There are no existing code examples - they're all new - so just make new code examples
		newCodeNodes := make([]types.CodeNode, 0)
		for _, snootyNode := range incomingCodeNodes {
			newNode := snooty.MakeCodeNodeFromSnootyAST(snootyNode, llm, ctx, isDriversProject)
			newCodeNodes = append(newCodeNodes, newNode)
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

		// Add relevant entries to the project report
		pageUpdatedChange := types.Change{
			Type: types.PageUpdated,
			Data: fmt.Sprintf("Page ID: %s", existingPage.ID),
		}
		report.Changes = append(report.Changes, pageUpdatedChange)
		if newCodeNodeCount > 0 {
			report.Counter.NewCodeNodesCount += newCodeNodeCount
			newExamplesChange := types.Change{
				Type: types.CodeExampleCreated,
				Data: fmt.Sprintf("Page ID: %s, %d new code examples added", existingPage.ID, newCodeNodeCount)}
			report.Changes = append(report.Changes, newExamplesChange)
		}
	} else if existingPage.Nodes == nil && incomingCodeNodePageCount == 0 {
		// No code examples to deal with here - just return nil and the unchanged report
		return nil, report
	} else {
		// Add an entry to the project report for updating the page. Adding this first so it precedes individual changes.
		// Note we're not reporting on any changes here - any count changes are reported through
		// CompareExistingIncomingCodeExampleSlices()
		pageUpdatedChange := types.Change{
			Type: types.PageUpdated,
			Data: fmt.Sprintf("Page ID: %s", existingPage.ID),
		}
		report.Changes = append(report.Changes, pageUpdatedChange)

		// If some examples exist already, and some examples are coming in from snooty, they might be updated, new, removed, or unchanged.
		// Handle those distinct cases.
		var updatedCodeNodes []types.CodeNode
		updatedCodeNodes, report = compare_code_examples.CompareExistingIncomingCodeExampleSlices(*existingPage.Nodes, incomingCodeNodes, report, existingPage.ID, llm, ctx, isDriversProject)
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

		// Update the report for changes to the code node count, literalinclude count, or io-code-block count
		oldCodeNodeCount := existingPage.CodeNodesTotal
		oldLiteralIncludeCount := existingPage.LiteralIncludesTotal
		oldIoCodeBlockCount := existingPage.IoCodeBlocksTotal
		if oldCodeNodeCount != incomingCodeNodePageCount {
			codeNodeCountChange := types.Change{
				Type: types.CodeNodeCountChange,
				Data: fmt.Sprintf("Page ID: %s, code node total for page was %d, now %d", existingPage.ID, oldCodeNodeCount, incomingCodeNodePageCount),
			}
			report.Changes = append(report.Changes, codeNodeCountChange)
		}
		if oldLiteralIncludeCount != incomingLiteralIncludeNodeCount {
			literalIncludeCountChange := types.Change{
				Type: types.LiteralIncludeCountChange,
				Data: fmt.Sprintf("Page ID: %s, literalinclude total for page was %d, now %d", existingPage.ID, oldLiteralIncludeCount, incomingLiteralIncludeNodeCount),
			}
			report.Changes = append(report.Changes, literalIncludeCountChange)
		}
		if oldIoCodeBlockCount != incomingIoCodeBlockNodeCount {
			ioCodeBlockCountChange := types.Change{
				Type: types.IoCodeBlockCountChange,
				Data: fmt.Sprintf("Page ID: %s, io-code-block total for page was %d, now %d", existingPage.ID, oldIoCodeBlockCount, incomingIoCodeBlockNodeCount),
			}
			report.Changes = append(report.Changes, ioCodeBlockCountChange)
		}
	}
	return &updatedDocsPage, report
}
