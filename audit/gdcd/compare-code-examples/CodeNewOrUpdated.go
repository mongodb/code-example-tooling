package compare_code_examples

import (
	"common"
	"gdcd/types"
	"strings"
)

const (
	updated               = "updated"
	newExample            = "new"
	percentChangeAccepted = float64(30)
)

// CodeNewOrUpdated takes the map of Sha256 hashes and code nodes that are already in Atlas, and compares the incoming ASTNode
// against the existing code nodes to figure out if it is a new code example or an existing code example that is updated.
// If the AST node text is within the matching percentage we accept, we consider the example "updated" -
// otherwise, we consider it "new." If it's updated, we also hand back the existing code node.
func CodeNewOrUpdated(existingSha256ToCodeNodeMap map[string]common.CodeNode, node types.ASTNode) (string, *common.CodeNode) {
	whitespaceTrimmedString := strings.TrimSpace(node.Value)
	// Could be updated, could be new
	for _, existingNode := range existingSha256ToCodeNodeMap {
		isUpdated := DiffCodeExamples(existingNode.Code, whitespaceTrimmedString, percentChangeAccepted)
		if isUpdated {
			return updated, &existingNode
		}
	}
	return newExample, nil
}
