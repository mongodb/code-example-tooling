package compare_code_examples

import (
	"common"
	"time"
)

// HandleRemovedCodeNodes takes a slice of []types.CodeNode, updates them to set them as removed, and hands it back to
// the call site. We append all the "Handle" function results to a slice, and overwrite the document in the DB with the
// updated code nodes. We don't just remove the nodes directly because we want to maintain a count of codes that we
// have removed - i.e. if we remove removed nodes, and add new nodes, the count stays the same and we can't track
// net new code examples.
func HandleRemovedCodeNodes(removedCodeNodes []common.CodeNode) ([]common.CodeNode, int) {
	updatedRemovedNodes := make([]common.CodeNode, 0)
	updatedRemovedCodeNodeCount := 0
	for _, node := range removedCodeNodes {
		node.IsRemoved = true
		node.DateRemoved = time.Now()
		if node.InstancesOnPage > 1 {
			updatedRemovedCodeNodeCount += node.InstancesOnPage
			node.InstancesOnPage = 0
		} else {
			updatedRemovedCodeNodeCount++
		}
		node.InstancesOnPage = 0
		updatedRemovedNodes = append(updatedRemovedNodes, node)
	}
	return updatedRemovedNodes, updatedRemovedCodeNodeCount
}
