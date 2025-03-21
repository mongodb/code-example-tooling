package compare_code_examples

import (
	"common"
	"gdcd/compare-code-examples/data"
	"testing"
	"time"
)

func IsApproximatelyNow(t time.Time, tolerance time.Duration) bool {
	now := time.Now()
	return t.Before(now.Add(tolerance))
}

func TestHandleRemovedNodesCorrectlySetsRemovedValues(t *testing.T) {
	codeNode, _ := data.GetRemovedNodes()
	updatedRemovedNodes := HandleRemovedCodeNodes([]common.CodeNode{codeNode})
	removedNode := updatedRemovedNodes[0]
	tolerance := 2 * time.Second // Define tolerance of 2 seconds
	if !IsApproximatelyNow(removedNode.DateRemoved, tolerance) {
		t.Errorf("FAILED: removed node time is not approximately now")
	}
	if removedNode.IsRemoved == false {
		t.Errorf("FAILED: removed node is not marked removed")
	}
}
