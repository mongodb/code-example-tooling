package compare_code_examples

import "common"

func GetCodeNodeCount(codeNodes []common.CodeNode) int {
	count := 0
	for _, codeNode := range codeNodes {
		// If the `InstancesOnPage` field is initialized, it should have a count, and we add it to the count. If it's
		// not initialized, Go default initializes int values to 0, so its value should count as 0 here. In that case,
		// just increment the count by 1.
		if codeNode.InstancesOnPage != 0 {
			count += codeNode.InstancesOnPage
		} else {
			count++
		}
	}
	return count
}
