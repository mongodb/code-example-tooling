package main

import (
	"log"
	"snooty-api-parser/types"
	"strings"
	"time"
)

const (
	updated               = "updated"
	newExample            = "new"
	unchanged             = "unchanged"
	percentChangeAccepted = float64(50)
)

func CompareExistingIncomingCodeExampleSlices(existingNodes []types.CodeNode, incomingNodes []types.ASTNode, projectCounter types.ProjectCounts, pageId string) ([]types.CodeNode, types.ProjectCounts) {
	var updatedPageNodes []types.ASTNode
	var newPageNodes []types.ASTNode
	var unchangedPageNodes []types.ASTNode

	snootySha256Hashes := make(map[string]int)
	existingSha256Hashes := make(map[string]int)
	existingSha256ToCodeNodeMap := make(map[string]types.CodeNode)
	incomingUnchangedSha256ToCodeNodeMap := make(map[string]types.CodeNode)
	incomingUpdatedSha256ToCodeNodeMap := make(map[string]types.CodeNode)
	for _, node := range incomingNodes {
		incomingNodeSha256Hash := MakeSha256HashForCode(node.Value)
		snootySha256Hashes[incomingNodeSha256Hash]++
	}
	for _, node := range existingNodes {
		existingSha256Hashes[node.SHA256Hash]++
		existingSha256ToCodeNodeMap[node.SHA256Hash] = node
	}

	for _, node := range incomingNodes {
		hash := MakeSha256HashForCode(node.Value)
		bucketName, existingNode := ChooseBucketForNode(existingNodes, existingSha256Hashes, node)
		switch bucketName {
		case unchanged:
			if existingNode != nil {
				incomingUnchangedSha256ToCodeNodeMap[hash] = *existingNode
			}
			unchangedPageNodes = append(unchangedPageNodes, node)
		case updated:
			if existingNode != nil {
				incomingUpdatedSha256ToCodeNodeMap[hash] = *existingNode
			}
			updatedPageNodes = append(updatedPageNodes, node)
		case newExample:
			newPageNodes = append(newPageNodes, node)
		default:
			log.Printf("Bucket '%s' not found in existing nodes\n", bucketName)
		}
	}
	removedNodes := FindRemovedNodes(existingSha256ToCodeNodeMap, unchangedPageNodes, updatedPageNodes, newPageNodes)

	codeNodesToReturn := make([]types.CodeNode, 0)
	codeNodesToReturn, projectCounter = MakeUpdatedCodeNodesArray(removedNodes, unchangedPageNodes, incomingUnchangedSha256ToCodeNodeMap, updatedPageNodes, incomingUpdatedSha256ToCodeNodeMap, newPageNodes, existingNodes, projectCounter, pageId)
	codeNodeChangesMinusRemoved := len(codeNodesToReturn) - len(removedNodes)
	if codeNodeChangesMinusRemoved != len(incomingNodes) {
		log.Printf("ISSUE: page %s, code node changes minus removed: %d does not match incoming node count %d\n", pageId, codeNodeChangesMinusRemoved, len(incomingNodes))
	}
	return codeNodesToReturn, projectCounter
}

func ChooseBucketForNode(existingNodes []types.CodeNode, existingSha256Hashes map[string]int, node types.ASTNode) (string, *types.CodeNode) {
	whitespaceTrimmedString := strings.TrimSpace(node.Value)
	hash := MakeSha256HashForCode(whitespaceTrimmedString)
	_, exists := existingSha256Hashes[hash]
	if !exists {
		// Could be updated, could be new
		for _, existingNode := range existingNodes {
			isUpdated := DiffCodeExamples(existingNode.Code, whitespaceTrimmedString, percentChangeAccepted)
			if isUpdated {
				return updated, &existingNode
			}
		}
		return newExample, nil
	} else {
		return unchanged, nil
	}
}

func MakeUpdatedCodeNodesArray(removedCodeNodes []types.CodeNode, unchangedPageNodes []types.ASTNode, unchangedPageNodesSha256CodeNodeLookup map[string]types.CodeNode, updatedPageNodes []types.ASTNode, updatedPageNodesSha256CodeNodeLookup map[string]types.CodeNode, newPageNodes []types.ASTNode, existingNodes []types.CodeNode, projectCounts types.ProjectCounts, pageId string) ([]types.CodeNode, types.ProjectCounts) {
	aggregateCodeNodeChanges := make([]types.CodeNode, 0)
	existingHashCountMap := make(map[string]int)
	for _, existingNode := range existingNodes {
		existingHashCountMap[existingNode.Code]++
	}

	// Call all the 'Handle' functions in sequence
	unchangedCodeNodes, unchangedCount, newCount, removedCount := HandleUnchangedPageNodes(existingHashCountMap, unchangedPageNodes, unchangedPageNodesSha256CodeNodeLookup, pageId)
	projectCounts.ExistingCodeNodesCount += unchangedCount
	projectCounts.NewCodeNodesCount += newCount
	projectCounts.RemovedCodeNodesCount += removedCount

	updatedCodeNodes := HandleUpdatedPageNodes(updatedPageNodes, updatedPageNodesSha256CodeNodeLookup)
	projectCounts.UpdatedCodeNodesCount += len(updatedCodeNodes)
	newCodeNodes := HandleNewPageNodes(newPageNodes)
	projectCounts.NewCodeNodesCount += len(newCodeNodes)
	removedCodeNodesUpdatedForRemoval := HandleRemovedCodeNodes(removedCodeNodes)
	projectCounts.RemovedCodeNodesCount += len(removedCodeNodesUpdatedForRemoval)

	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, unchangedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, updatedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, newCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, removedCodeNodesUpdatedForRemoval...)
	return aggregateCodeNodeChanges, projectCounts
}

func HandleUnchangedPageNodes(existingHashCountMap map[string]int, unchangedIncomingPageNodes []types.ASTNode, unchangedPageNodesSha256CodeNodeLookup map[string]types.CodeNode, pageId string) ([]types.CodeNode, int, int, int) {
	codeNodesMatchingIncomingNodes := make([]types.CodeNode, 0)
	unchangedIncomingHashCountMap := make(map[string]int)
	unchangedCount := 0
	removedCount := 0
	newCount := 0
	totalCount := len(unchangedIncomingPageNodes)
	for _, node := range unchangedIncomingPageNodes {
		hash := MakeSha256HashForCode(node.Value)
		unchangedIncomingHashCountMap[hash]++
	}
	for hash, count := range unchangedIncomingHashCountMap {
		// If the number of times the hash is in the incoming count equals the number of times the hash is already on the page,
		// we just need this many unchanged nodes unmodified
		if existingHashCountMap[hash] == count {
			node := unchangedPageNodesSha256CodeNodeLookup[hash]
			for i := 0; i < count; i++ {
				codeNodesMatchingIncomingNodes = append(codeNodesMatchingIncomingNodes, node)
				unchangedCount++
			}
		} else if existingHashCountMap[hash] > count {
			// Some have been removed from the page. Mark the relevant number of nodes as removed.
			numberToRemove := existingHashCountMap[hash] - count

			node := unchangedPageNodesSha256CodeNodeLookup[hash]
			for i := 0; i < numberToRemove; i++ {
				removedNode := node
				removedNode.IsRemoved = true
				removedNode.DateRemoved = time.Now()
				codeNodesMatchingIncomingNodes = append(codeNodesMatchingIncomingNodes, removedNode)
				removedCount++
			}
			for i := 0; i < count; i++ {
				codeNodesMatchingIncomingNodes = append(codeNodesMatchingIncomingNodes, node)
				unchangedCount++
			}
		} else {
			// Some have been added to the page. Create the relevant number of new nodes.
			numberToAdd := existingHashCountMap[hash] - count
			node := unchangedPageNodesSha256CodeNodeLookup[hash]
			for i := 0; i < numberToAdd; i++ {
				newNode := node
				newNode.DateUpdated = time.Now()
				codeNodesMatchingIncomingNodes = append(codeNodesMatchingIncomingNodes, newNode)
				newCount++
			}
			for i := 0; i < count; i++ {
				codeNodesMatchingIncomingNodes = append(codeNodesMatchingIncomingNodes, node)
				unchangedCount++
			}
		}
	}
	sum := unchangedCount + removedCount + newCount
	if totalCount != sum {
		log.Printf("ISSUE: page %s, in HandleUnchangedPageNodes, unchangedCount %d, removedCount %d, newCount %d, sum %d, does not equal total incoming unchanged node count %d\n", pageId, unchangedCount, removedCount, newCount, sum, totalCount)
	}
	return codeNodesMatchingIncomingNodes, unchangedCount, newCount, removedCount
}

func HandleUpdatedPageNodes(updatedPageNodes []types.ASTNode, incomingSha256ToCodeNodesMap map[string]types.CodeNode) []types.CodeNode {
	codeNodeUpdates := make([]types.CodeNode, 0)
	for _, incomingNode := range updatedPageNodes {
		whiteSpaceTrimmedString := strings.TrimSpace(incomingNode.Value)
		hash := MakeSha256HashForCode(whiteSpaceTrimmedString)
		codeNode := incomingSha256ToCodeNodesMap[hash]
		codeNode.Code = whiteSpaceTrimmedString
		codeNode.SHA256Hash = hash
		codeNode.DateUpdated = time.Now()
		codeNodeUpdates = append(codeNodeUpdates, codeNode)
	}
	return codeNodeUpdates
}

func HandleNewPageNodes(newIncomingPageNodes []types.ASTNode) []types.CodeNode {
	newNodes := make([]types.CodeNode, 0)
	for _, incomingNode := range newIncomingPageNodes {
		newNode := MakeCodeNodeFromSnootyAST(incomingNode)
		newNodes = append(newNodes, newNode)
	}
	return newNodes
}

func HandleRemovedCodeNodes(removedCodeNodes []types.CodeNode) []types.CodeNode {
	updatedRemovedNodes := make([]types.CodeNode, 0)
	for _, node := range removedCodeNodes {
		node.IsRemoved = true
		node.DateUpdated = time.Now()
		updatedRemovedNodes = append(updatedRemovedNodes, node)
	}
	return updatedRemovedNodes
}

func FindRemovedNodes(existingNodeHashMap map[string]types.CodeNode, unchangedBucket []types.ASTNode, updatedBucket []types.ASTNode, newBucket []types.ASTNode) []types.CodeNode {
	unchangedHashBool := make(map[string]bool)
	removedNodes := make([]types.CodeNode, 0)
	for _, node := range unchangedBucket {
		hash := MakeSha256HashForCode(node.Value)
		unchangedHashBool[hash] = true
	}
	updatedHashBool := make(map[string]bool)
	for _, node := range updatedBucket {
		hash := MakeSha256HashForCode(node.Value)
		updatedHashBool[hash] = true
	}
	newHashBool := make(map[string]bool)
	for _, node := range newBucket {
		hash := MakeSha256HashForCode(node.Value)
		newHashBool[hash] = true
	}
	for hash, node := range existingNodeHashMap {
		found := false
		if unchangedHashBool[hash] {
			found = true
		} else if updatedHashBool[hash] {
			found = true
		} else if newHashBool[hash] {
			found = true
		}
		if !found {
			removedNodes = append(removedNodes, node)
		}
	}
	return removedNodes
}
