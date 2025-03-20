package compare_code_examples

import (
	"gdcd/compare-code-examples/data"
	"gdcd/types"
	"testing"
)

func TestChooseBucketForNewNode(t *testing.T) {
	codeNode, astNode := data.GetNewNodes()
	existingSha256Hashes := make(map[string]int)
	existingSha256Hashes[codeNode.SHA256Hash] = 1
	bucket, _ := ChooseBucketForNode([]types.CodeNode{codeNode}, existingSha256Hashes, astNode)
	if bucket != newExample {
		t.Errorf("FAILED: got %s bucket, want %s", bucket, newExample)
	}
}

func TestChooseBucketForUpdatedNode(t *testing.T) {
	codeNode, astNode := data.GetUpdatedNodes()
	existingSha256Hashes := make(map[string]int)
	existingSha256Hashes[codeNode.SHA256Hash] = 1
	bucket, _ := ChooseBucketForNode([]types.CodeNode{codeNode}, existingSha256Hashes, astNode)
	if bucket != updated {
		t.Errorf("FAILED: got %s bucket, want %s", bucket, updated)
	}
}

func TestChooseBucketForUnchangedNode(t *testing.T) {
	codeNode, astNode := data.GetUnchangedNodes()
	existingSha256Hashes := make(map[string]int)
	existingSha256Hashes[codeNode.SHA256Hash] = 1
	bucket, _ := ChooseBucketForNode([]types.CodeNode{codeNode}, existingSha256Hashes, astNode)
	if bucket != unchanged {
		t.Errorf("FAILED: got %s bucket, want %s", bucket, unchanged)
	}
}

// Note: there is no test for the "removed" case because that condition doesn't call the ChooseBucketForNode function.
// Instead, that condition is handled by the FindRemovedNodes function
