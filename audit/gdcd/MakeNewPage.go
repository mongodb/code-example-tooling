package main

import (
	"common"
	"context"
	"gdcd/snooty"
	"gdcd/types"
	"gdcd/utils"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

func MakeNewPage(data types.PageMetadata, projectName string, siteUrl string, llm *ollama.LLM, ctx context.Context) common.DocsPage {
	incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes := snooty.GetCodeExamplesFromIncomingData(data.AST)
	incomingCodeNodeCount := len(incomingCodeNodes)
	incomingLiteralIncludeNodeCount := len(incomingLiteralIncludeNodes)
	incomingIoCodeNodeCount := len(incomingIoCodeBlockNodes)
	pageId := utils.ConvertSnootyPageIdToAtlasPageId(data.PageID)
	pageUrl := utils.ConvertSnootyPageIdToProductionUrl(data.PageID, siteUrl)
	product, subProduct := GetProductSubProduct(projectName, pageUrl)
	var isDriversProject bool
	if product == "Drivers" {
		isDriversProject = true
	} else {
		isDriversProject = false
	}

	// Some of the new code examples coming in from the page may be duplicates. So we first make Sha256 hashes of the
	// incoming code examples, and count the number of times the hash appears on the page.
	snootySha256Hashes := make(map[string]int)
	snootySha256ToAstNodeMap := make(map[string]types.ASTNode)

	for _, node := range incomingCodeNodes {
		// This makes a hash from the whitespace-trimmed AST node. We trim whitespace on AST nodes before adding
		// them to the DB, so this ensures an incoming node hash can match a whitespace-trimmed existing node hash.
		hash := snooty.MakeSha256HashForCode(node.Value)

		// Add the hash as an entry in the map, and increment its counter. If the hash does not already exist in the map,
		// this will create it. If it does already exist, this will just increment its counter.
		snootySha256Hashes[hash]++
		snootySha256ToAstNodeMap[hash] = node
	}

	// Then, we go through the hashes, create the corresponding codeNodes, and set the `InstancesOnPage` if the example
	// appears more than once on the page.
	var newCodeNodes []common.CodeNode
	for hash, count := range snootySha256Hashes {
		newNode := snooty.MakeCodeNodeFromSnootyAST(snootySha256ToAstNodeMap[hash], llm, ctx, isDriversProject)
		if count > 1 {
			newNode.InstancesOnPage = count
		}
		newCodeNodes = append(newCodeNodes, newNode)
	}

	maybeKeywords := snooty.GetMetaKeywords(data.AST.Children)

	languagesArrayValues := MakeLanguagesArray(newCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes)

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
		ProjectName:          projectName,
		Product:              product,
		SubProduct:           subProduct,
		Keywords:             maybeKeywords,
	}
}
