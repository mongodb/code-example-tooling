package db

import "gdcd/types"

// GetCurrentRemovedAtlasCodeNodes takes the []types.CodeNode that already exist on the page, and separates them into
// two slices - one for code nodes that aren't marked IsRemoved and one for code nodes that have already been marked
// IsRemoved in a prior run. The ones that have not already been removed should be considered "current", and the ones
// that have previously been marked as removed should be appended to the array unchanged after processing is complete.
func GetCurrentRemovedAtlasCodeNodes(existingNodes []types.CodeNode) ([]types.CodeNode, []types.CodeNode) {
	currentCodeNodes := make([]types.CodeNode, 0)
	removedCodeNodes := make([]types.CodeNode, 0)
	for _, existingNode := range existingNodes {
		if existingNode.IsRemoved {
			removedCodeNodes = append(removedCodeNodes, existingNode)
		} else {
			currentCodeNodes = append(currentCodeNodes, existingNode)
		}
	}
	return currentCodeNodes, removedCodeNodes
}
