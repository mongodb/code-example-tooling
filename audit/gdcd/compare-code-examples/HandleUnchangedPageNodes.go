package compare_code_examples

import (
	"common"
	"gdcd/snooty"
	"gdcd/types"
	"time"
)

// HandleUnchangedPageNodes takes a map that counts the number of times the existing hash appears on the page, a slice
// of []types.ASTNode, and a lookup map that maps SHA256 hashes to types.CodeNode. If an example appears more than once
// on a page, we check the count of how many times it already exists on the page, compare it to the count of how many
// times it appears on the incoming page, and add or remove instances to the []common.CodeNode array to match the existing
// page count. We return the updated []common.CodeNode array. We append all the "Handle" function results to a slice,
// and overwrite the document in the DB with the updated code nodes.
func HandleUnchangedPageNodes(existingHashCountMap map[string]int, unchangedIncomingPageNodes []types.ASTNode, unchangedPageNodesSha256CodeNodeLookup map[string]common.CodeNode) []common.CodeNode {
	codeNodesMatchingIncomingNodes := make([]common.CodeNode, 0)
	unchangedIncomingHashCountMap := make(map[string]int)
	unchangedCount := 0
	removedCount := 0
	newCount := 0
	for _, node := range unchangedIncomingPageNodes {
		hash := snooty.MakeSha256HashForCode(node.Value)
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
	return codeNodesMatchingIncomingNodes
}
