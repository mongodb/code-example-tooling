package compare_code_examples

import (
	"common"
	"context"
	"gdcd/snooty"
	"gdcd/types"
	"github.com/tmc/langchaingo/llms/ollama"
)

// CompareExistingIncomingCodeExampleSlices takes []common.CodeNode, which represents the existing code example nodes from
// Atlas, and []types.ASTNode, which represents incoming code examples from the Snooty Data API. It also takes a types.ProjectReport
// to track various project changes and counts. This function compares the existing code examples with the incoming code examples
// to find unchanged, updated, new, and removed nodes. It appends these nodes into an updated []common.CodeNode slice,
// which it returns to the call site for making updates to Atlas. It also returns the updated types.ProjectReport.
// ASTNode state can be one of three things: new, unchanged, or updated.
// CodeNode state can be one of three things: unchanged, updated, or removed.
// This function attempts to assign a state to & appropriately handle every node.
func CompareExistingIncomingCodeExampleSlices(existingNodes []common.CodeNode, existingRemovedNodes []common.CodeNode, incomingNodes []types.ASTNode, report types.ProjectReport, pageId string, llm *ollama.LLM, ctx context.Context, isDriversProject bool) ([]common.CodeNode, types.ProjectReport) {
	// These are page nodes that are a partial match for nodes on the page. We assume they are making updates to an existing node.
	var updatedPageNodes []types.ASTNodeWrapper

	// These are incoming AST nodes that do not match any existing code nodes on the page. They are net new.
	var newPageNodes []types.ASTNodeWrapper

	// These are existing code nodes from the database that match incoming AST nodes from the Snooty Data API.
	// They are exact matches that are unchanged.
	var unchangedNodes []common.CodeNode

	// These are existing code nodes from the database, but are not coming in from the Snooty Data API. They must inherently
	// be removed from the page.
	var removedCodeNodes []common.CodeNode

	incomingCount := len(incomingNodes)

	// This will be a map of sha256 hashes for AST nodes coming in on the page from the Snooty Data API. The int
	// value represents the number of times the node's hash appears on the page.
	snootySha256Hashes := make(map[string]int)
	snootySha256ToAstNodeMap := make(map[string]types.ASTNode)

	// This will be a map of sha256 hashes for existing code nodes that are already in the database. The int value represents
	// the number of times the node's hash appears in the database. As we potentially match them with incoming AST nodes,
	// we will decrement the counter and/or remove the hash from the map. Incoming AST nodes should only
	// match 0 or 1 existing sha256 hashes, so we should eliminate them as potential matches once they have been matched.
	unmatchedSha256Hashes := make(map[string]int)

	// This map serves as a lookup table to easily find the code node that matches the given sha256 hash.
	unmatchedSha256ToCodeNodeMap := make(map[string]common.CodeNode)

	// This map serves as a lookup table to easily find the code node that matches the incoming sha256 hash in the
	// function to make the new array of code examples.
	incomingUpdatedSha256ToCodeNodeMap := make(map[string]common.CodeNode)

	// The same code example could theoretically appear more than once on a page. If a sha256 hash appears more than once
	// on a page, we increment the count for the existing hash. Build the map of hashes for the existing code nodes
	// in the database, and their counts. Also, create a lookup map to find the code node matching a given hash.
	for _, node := range existingNodes {
		unmatchedSha256Hashes[node.SHA256Hash]++
		unmatchedSha256ToCodeNodeMap[node.SHA256Hash] = node
	}

	// Create a SHA256 hash map for the incoming AST nodes for easy comparison with existing code nodes
	for _, node := range incomingNodes {
		// This makes a hash from the whitespace-trimmed AST node. We trim whitespace on AST nodes before adding
		// them to the DB, so this ensures an incoming node hash can match a whitespace-trimmed existing node hash.
		hash := snooty.MakeSha256HashForCode(node.Value)

		// Add the hash as an entry in the map, and increment its counter. If the hash does not already exist in the map,
		// this will create it. If it does already exist, this will just increment its counter.
		snootySha256Hashes[hash]++
		snootySha256ToAstNodeMap[hash] = node
	}

	// First, check for incoming AST nodes that are exact matches for existing code nodes. Consider both incoming and
	// existing nodes "unchanged" and remove them from the potential comparison candidates.
	for hash, count := range snootySha256Hashes {
		// Check to see if the incoming AST node hash is an exact match for an unmatched existing code node hash, and
		// the count is at least 1
		if unmatchedSha256Hashes[hash] >= 1 {
			// Get the matching code node
			unchangedCodeNode := unmatchedSha256ToCodeNodeMap[hash]

			if unchangedCodeNode.InstancesOnPage != 0 && unchangedCodeNode.InstancesOnPage != count {
				// If the unchanged code node does not match the count of number of times this hash appears, decrement
				// one instance from the counter since we are counting it as a "match" here. we don't just want to
				// delete it because it may also match an "update" later.
				unmatchedSha256Hashes[hash]--
			} else if unchangedCodeNode.InstancesOnPage == 0 && unmatchedSha256Hashes[hash] > 1 {
				// If `InstancesOnPage` is unitialized, we can't compare it with the hash count, so just decrement the hash count
				unmatchedSha256Hashes[hash]--
			} else {
				// If it _does_ match the number of times the hash appears, consider it unchanged. Delete it from the
				// unmatched hash list and map. Now that it has matched, we don't need to consider it as a possible
				// match for other nodes.
				delete(unmatchedSha256Hashes, hash)
				delete(unmatchedSha256ToCodeNodeMap, hash)
			}

			// Update the count to reflect how many times it currently appears on the page
			unchangedCodeNode.InstancesOnPage = count

			// Append it to the array of unchanged nodes. We use this to rebuild the array of code nodes we'll write to the DB.
			unchangedNodes = append(unchangedNodes, unchangedCodeNode)

			// Delete it from the incoming hash list and map.
			delete(snootySha256Hashes, hash)
			delete(snootySha256ToAstNodeMap, hash)
		}
	}

	// Now start checking whether the remaining incoming AST nodes are updates or net new examples.
	for hash, count := range snootySha256Hashes {
		astNode := snootySha256ToAstNodeMap[hash]
		nodePlusMetadata := types.ASTNodeWrapper{
			InstancesOnPage: count,
			Node:            astNode,
		}
		// Figure out whether the AST node is new or updated. If it matches an existing code node in the DB,
		// this function returns the existing code node along with the string "newExample" or "updated".
		newOrUpdated, existingNode := CodeNewOrUpdated(unmatchedSha256ToCodeNodeMap, astNode)
		if newOrUpdated == newExample {
			newPageNodes = append(newPageNodes, nodePlusMetadata)
		} else {
			if existingNode != nil {
				incomingUpdatedSha256ToCodeNodeMap[hash] = *existingNode

				// If the incoming AST node counts as an update for an existing code node, and that node's SHA256 hash
				// only exists once on the page, remove the node from the "eligible" nodes for comparison. Each incoming
				// AST node should match 0 or at most 1 existing code nodes. Once the nodes have been matched, the
				// existing code node should no longer be eligible for matching.
				if unmatchedSha256Hashes[existingNode.SHA256Hash] == 1 {
					delete(unmatchedSha256Hashes, existingNode.SHA256Hash)
					delete(unmatchedSha256ToCodeNodeMap, existingNode.SHA256Hash)
				} else {
					// If a sha256 hash appears more than once on a page, decrement one instance from the counter since
					// we are counting it as a "match" here
					unmatchedSha256Hashes[existingNode.SHA256Hash]--
				}
			}
			updatedPageNodes = append(updatedPageNodes, nodePlusMetadata)
		}
	}

	// If there are any unmatched existing code nodes after this process is complete, they must have been removed from the page.
	if len(unmatchedSha256Hashes) > 0 {
		for hash, _ := range unmatchedSha256Hashes {
			removedCodeNodes = append(removedCodeNodes, unmatchedSha256ToCodeNodeMap[hash])
		}
	}

	// Make the complete array of code nodes, which will overwrite the existing array. This array consists of: all
	// previously removed nodes, new removed nodes as of this run, unchanged nodes, updated nodes, and net new nodes.
	// This function also calls the func to update the report based on the counts.
	return MakeUpdatedCodeNodesArray(removedCodeNodes, existingRemovedNodes, unchangedNodes,
		updatedPageNodes, incomingUpdatedSha256ToCodeNodeMap, newPageNodes,
		incomingCount, report, pageId, llm, ctx, isDriversProject)
}
