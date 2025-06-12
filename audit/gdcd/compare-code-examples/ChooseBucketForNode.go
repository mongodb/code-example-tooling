package compare_code_examples

import (
	"common"
	"gdcd/snooty"
	"gdcd/types"
	"strings"
)

const (
	updated               = "updated"
	newExample            = "new"
	unchanged             = "unchanged"
	percentChangeAccepted = float64(30)
)

// ChooseBucketForNode takes the map of Sha256 hashes and nodes that are already in Atlas, and compares the incoming ASTNode
// against the existing hashes and nodes to figure out if it is a new code example, an existing code example that is
// unchanged, or an updated code example. If the SHA256 hash is an exact match for a code example already on the page,
// it is unchanged. If the code example text is within the matching percentage we accept, we consider it "updated" -
// otherwise, we consider it "new." If it's unchanged or updated, we also hand back the existing node.
func ChooseBucketForNode(existingSha256Hashes map[string]int, existingSha256ToCodeNodeMap map[string]common.CodeNode, node types.ASTNode) (string, *common.CodeNode) {
	whitespaceTrimmedString := strings.TrimSpace(node.Value)
	hash := snooty.MakeSha256HashForCode(whitespaceTrimmedString)
	_, exists := existingSha256Hashes[hash]
	if exists {
		matchingNode := existingSha256ToCodeNodeMap[hash]
		return unchanged, &matchingNode
	} else {
		// Could be updated, could be new
		for _, existingNode := range existingSha256ToCodeNodeMap {
			isUpdated := DiffCodeExamples(existingNode.Code, whitespaceTrimmedString, percentChangeAccepted)
			if isUpdated {
				return updated, &existingNode
			}
		}
		return newExample, nil
	}
}
