package compare_code_examples

import "snooty-api-parser/types"

func IncrementProjectCounterForUpdatedCodeNodes(projectCounter types.ProjectCounts, unchangedCount int, updatedCount int, newCount int, removedCount int) types.ProjectCounts {
	projectCounter.UnchangedCodeNodesCount += unchangedCount
	projectCounter.UpdatedCodeNodesCount += updatedCount
	projectCounter.NewCodeNodesCount += newCount
	projectCounter.RemovedCodeNodesCount += removedCount
	return projectCounter
}
