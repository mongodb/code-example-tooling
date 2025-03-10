package compare_code_examples

import (
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
)

// FindRemovedNodes takes a map with a hash of SHA256es for existing nodes, and slices that represent all the other bucket
// types - unchanged examples, updated examples, and new examples. We compare the SHA256 hashes of the existing code
// examples against SHA256 hashes from the incoming buckets. If the SHA256 hash doesn't match any of the incoming examples,
// the example has been removed from the page. Append it to an array of removed nodes and hand it back to the call
// site for handling in the DB.
func FindRemovedNodes(existingNodeHashMap map[string]types.CodeNode, unchangedBucket []types.ASTNode, updatedBucket []types.ASTNode, newBucket []types.ASTNode) []types.CodeNode {
	unchangedHashBool := make(map[string]bool)
	removedNodes := make([]types.CodeNode, 0)
	for _, node := range unchangedBucket {
		hash := snooty.MakeSha256HashForCode(node.Value)
		unchangedHashBool[hash] = true
	}
	updatedHashBool := make(map[string]bool)
	for _, node := range updatedBucket {
		hash := snooty.MakeSha256HashForCode(node.Value)
		updatedHashBool[hash] = true
	}
	newHashBool := make(map[string]bool)
	for _, node := range newBucket {
		hash := snooty.MakeSha256HashForCode(node.Value)
		newHashBool[hash] = true
	}
	for hash, node := range existingNodeHashMap {
		found := false
		if unchangedHashBool[hash] {
			found = true
		} else if updatedHashBool[hash] {
			found = true
		} else if newHashBool[hash] {
			found = true
		}
		if !found {
			removedNodes = append(removedNodes, node)
		}
	}
	return removedNodes
}
