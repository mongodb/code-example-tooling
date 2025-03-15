package compare_code_examples

import (
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
	"strings"
)

const (
	updated               = "updated"
	newExample            = "new"
	unchanged             = "unchanged"
	percentChangeAccepted = float64(50)
)

// ChooseBucketForNode takes the array of []types.CodeNode that are already in Atlas, and compares the incoming ASTNode
// against that array to figure out if it is a new code example, an existing code example that is unchanged, or an updated
// code example. If the SHA256 hash is an exact match for a code example already on the page, it is unchanged. If it is
// within the matching percentage we accept, we consider it "updated" - otherwise, we consider it "new."
// If it's an updated example, we also hand back the existing node so we can update it.
func ChooseBucketForNode(existingNodes []types.CodeNode, existingSha256Hashes map[string]int, node types.ASTNode) (string, *types.CodeNode) {
	whitespaceTrimmedString := strings.TrimSpace(node.Value)
	hash := snooty.MakeSha256HashForCode(whitespaceTrimmedString)
	_, exists := existingSha256Hashes[hash]
	if exists {
		return unchanged, nil
	} else {
		// Could be updated, could be new
		for _, existingNode := range existingNodes {
			isUpdated := DiffCodeExamples(existingNode.Code, whitespaceTrimmedString, percentChangeAccepted)
			if isUpdated {
				return updated, &existingNode
			}
		}
		return newExample, nil
	}
}
