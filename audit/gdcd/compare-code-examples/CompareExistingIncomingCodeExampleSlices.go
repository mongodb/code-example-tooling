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
func CompareExistingIncomingCodeExampleSlices(existingNodes []common.CodeNode, existingRemovedNodes []common.CodeNode, incomingNodes []types.ASTNode, report types.ProjectReport, pageId string, llm *ollama.LLM, ctx context.Context, isDriversProject bool) ([]common.CodeNode, types.ProjectReport) {
	// These are page nodes that are a partial match for nodes on the page. We assume they are making updates to an existing node.
	var updatedPageNodes []types.ASTNode

	// These are page nodes that do not match any existing nodes on the page. They are net new.
	var newPageNodes []types.ASTNode

	// These are code examples that already exist on the page, and match incoming examples from the Snooty Data API.
	// They are exact matches that are unchanged.
	var unchangedNodes []common.CodeNode

	// These are page nodes that exist in the database, but are not coming in from the Snooty Data API. They must inherently
	// be removed from the page.
	var removedPageNodes []common.CodeNode

	incomingCount := len(incomingNodes)

	// This will be a map of sha256 hashes for code examples coming in on the page from the Snooty Data API. The int
	// value represents the number of times the hash appears on the page.
	snootySha256Hashes := make(map[string]int)
	snootySha256ToAstNodeMap := make(map[string]types.ASTNode)

	// This will be a map of sha256 hashes for code examples that are already in the database. The int value represents
	// the number of times the hash appears in the database. As we potentially match them with incoming AST nodes,
	// we will decrement the counter and/or remove the hash from the map. Nodes coming in from the page should only
	// match 0 or 1 existing sha256 hashes, so we should eliminate them as potential matches once they have been matched.
	unmatchedSha256Hashes := make(map[string]int)

	// This map serves as a lookup table to easily find the code node that matches the given sha256 hash.
	unmatchedSha256ToCodeNodeMap := make(map[string]common.CodeNode)

	// This map serves as a lookup table to easily find the code node that matches the incoming sha256 hash in the
	// function to make the new array of code examples.
	incomingUpdatedSha256ToCodeNodeMap := make(map[string]common.CodeNode)

	// The same code example could theoretically appear more than once on a page. If a sha256 hash appears more than once
	// on a page, we increment the count for the existing hash. Build the map of hashes for the existing code examples
	// in the database, and their counts. Also, create a lookup map to find the code node matching a given hash.
	for _, node := range existingNodes {
		unmatchedSha256Hashes[node.SHA256Hash]++
		unmatchedSha256ToCodeNodeMap[node.SHA256Hash] = node
	}

	// Create a SHA256 hash map for the incoming nodes for easy comparison with existing nodes
	for _, node := range incomingNodes {
		// This makes a hash from the whitespace-trimmed code example. We trim whitespace on code examples before adding
		// them to the DB, so this ensures an incoming example hash can match a whitespace-trimmed existing example match.
		hash := snooty.MakeSha256HashForCode(node.Value)

		// Add the hash as an entry in the map, and increment its counter. If the hash does not already exist in the map,
		// this will create it. If it does already exist, this will just increment its counter.
		snootySha256Hashes[hash]++
		snootySha256ToAstNodeMap[hash] = node
	}

	// First, check for incoming examples that are exact matches for existing examples. Consider it "unchanged" and
	// remove it from the potential comparison candidates.
	for hash, count := range snootySha256Hashes {
		// Check to see if the incoming code example hash is an exact match for an unmatched hash, and the count is at least 1
		if unmatchedSha256Hashes[hash] >= 1 {
			// Get the matching code node and append it to the array of unchanged nodes. We use this to construct the
			// new array of nodes on the page.
			unchangedCodeNode := unmatchedSha256ToCodeNodeMap[hash]
			unchangedNodes = append(unchangedNodes, unchangedCodeNode)
			// If this hash appears only once in the existing hashes, delete it from the hash list and map. We don't
			// want to consider it as a possible match for other examples.
			if unmatchedSha256Hashes[hash] == 1 {
				delete(unmatchedSha256Hashes, hash)
				delete(unmatchedSha256ToCodeNodeMap, hash)
			} else {
				// If a sha256 hash appears more than once on a page, decrement one instance from the counter since
				// we are counting it as a "match" here
				unmatchedSha256Hashes[hash]--
			}
			// If the hash exists only once in the incoming code examples hash map, delete it from the hash list and map.
			// Otherwise, decrement the count of times it appears.
			if count == 1 {
				delete(snootySha256Hashes, hash)
				delete(snootySha256ToAstNodeMap, hash)
			} else {
				snootySha256Hashes[hash]--
			}
		}
	}

	// Now start checking whether the remaining incoming examples are updates or net new examples.
	for hash, count := range snootySha256Hashes {
		astNode := snootySha256ToAstNodeMap[hash]
		// Figure out whether the code example is new, updated, or unchanged. If it matches an existing code example in the DB,
		// this function returns the existing code example along with the "bucket name".
		newOrUpdated, existingNode := CodeNewOrUpdated(unmatchedSha256ToCodeNodeMap, astNode)
		if newOrUpdated == newExample {
			newPageNodes = append(newPageNodes, astNode)
			// If the hash exists only once in the incoming code examples hash map, delete it from the hash list and map.
			// Otherwise, decrement the count of times it appears.
			if count == 1 {
				delete(snootySha256Hashes, hash)
				delete(snootySha256ToAstNodeMap, hash)
			} else {
				snootySha256Hashes[hash]--
			}
		} else {
			if existingNode != nil {
				incomingUpdatedSha256ToCodeNodeMap[hash] = *existingNode

				// If the incoming node counts as an update for an existing node, and that node's SHA256 hash only exists
				// once on the page, remove the node from the "eligible" nodes for comparison. Each incoming code example
				// should match 0 or at most 1 existing code examples. Once the code example has been matched, the
				// existing example should no longer be eligible for matching.
				if unmatchedSha256Hashes[existingNode.SHA256Hash] == 1 {
					delete(unmatchedSha256Hashes, existingNode.SHA256Hash)
					delete(unmatchedSha256ToCodeNodeMap, existingNode.SHA256Hash)
				} else {
					// If a sha256 hash appears more than once on a page, decrement one instance from the counter since
					// we are counting it as a "match" here
					unmatchedSha256Hashes[existingNode.SHA256Hash]--
				}
			}
			updatedPageNodes = append(updatedPageNodes, astNode)
		}
	}

	// If there are any unmatched existing code examples after this process is complete, they must have been removed from the page.
	if len(unmatchedSha256Hashes) > 0 {
		for hash, _ := range unmatchedSha256Hashes {
			removedPageNodes = append(removedPageNodes, unmatchedSha256ToCodeNodeMap[hash])
		}
	}

	// Make the complete array of code nodes, which will overwrite the existing array. This array consists of: all
	// previously removed nodes, new removed nodes as of this run, unchanged nodes, updated nodes, and net new nodes.
	// This function also calls the func to update the report based on the counts.
	codeNodesToReturn := make([]common.CodeNode, 0)
	codeNodesToReturn, report = MakeUpdatedCodeNodesArray(removedPageNodes, existingRemovedNodes, unchangedNodes,
		updatedPageNodes, incomingUpdatedSha256ToCodeNodeMap, newPageNodes,
		incomingCount, report, pageId, llm, ctx, isDriversProject)
	return codeNodesToReturn, report
}
