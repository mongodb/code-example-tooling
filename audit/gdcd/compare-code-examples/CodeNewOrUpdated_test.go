package compare_code_examples

import (
	"common"
	"gdcd/compare-code-examples/data"
	"testing"
)

func TestChooseBucketForNewNode(t *testing.T) {
	codeNode, astNode := data.GetNewNodes()
	existingSha256HashToNodeLookup := make(map[string]common.CodeNode)
	existingSha256HashToNodeLookup[codeNode.SHA256Hash] = codeNode
	bucket, _ := CodeNewOrUpdated(existingSha256HashToNodeLookup, astNode)
	if bucket != newExample {
		t.Errorf("FAILED: got %s bucket, want %s", bucket, newExample)
	}
}

func TestChooseBucketForUpdatedNode(t *testing.T) {
	codeNode, astNode := data.GetUpdatedNodes()
	existingSha256HashToNodeLookup := make(map[string]common.CodeNode)
	existingSha256HashToNodeLookup[codeNode.SHA256Hash] = codeNode
	bucket, _ := CodeNewOrUpdated(existingSha256HashToNodeLookup, astNode)
	if bucket != updated {
		t.Errorf("FAILED: got %s bucket, want %s", bucket, updated)
	}
}
