package compare_code_examples

import (
	"common"
	"gdcd/snooty"
	"gdcd/types"
	"strings"
	"time"
)

// HandleUpdatedPageNodes takes a slice of updated []types.ASTNode and a lookup map that maps incoming SHA256 hashes to
// the existing common.CodeNode that they matched in the CodeNewOrUpdated function. For each updated ASTNode, we look
// up the matching code node, update the Code field text, add the new SHA256Hash, and append an updated date. We return
// the updated []common.CodeNode array. We append all the "Handle" function results to a slice, and overwrite the
// document in the DB with the updated code nodes.
func HandleUpdatedPageNodes(updatedPageNodes []types.ASTNodeWrapper, incomingSha256ToCodeNodesMap map[string]common.CodeNode) ([]common.CodeNode, int) {
	codeNodeUpdates := make([]common.CodeNode, 0)
	updatedCodeNodeCount := 0
	for _, incomingNode := range updatedPageNodes {
		whiteSpaceTrimmedString := strings.TrimSpace(incomingNode.Node.Value)
		hash := snooty.MakeSha256HashForCode(whiteSpaceTrimmedString)
		codeNode := incomingSha256ToCodeNodesMap[hash]
		codeNode.Code = whiteSpaceTrimmedString
		codeNode.SHA256Hash = hash
		codeNode.DateUpdated = time.Now()
		if incomingNode.InstancesOnPage != 0 {
			codeNode.InstancesOnPage = incomingNode.InstancesOnPage
			updatedCodeNodeCount += codeNode.InstancesOnPage
		} else {
			updatedCodeNodeCount++
		}
		codeNodeUpdates = append(codeNodeUpdates, codeNode)
	}
	return codeNodeUpdates, updatedCodeNodeCount
}
