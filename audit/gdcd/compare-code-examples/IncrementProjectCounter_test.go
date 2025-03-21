package compare_code_examples

// TODO: Refactor tests after changing this func to take the project report instead of a project counter
//func TestIncrementProjectCounterCorrectlyUpdatesUnchangedCounts(t *testing.T) {
//	projectCounter := types.ProjectCounts{
//		IncomingCodeNodesCount:      0,
//		IncomingLiteralIncludeCount: 0,
//		IncomingIoCodeBlockCount:    0,
//		RemovedCodeNodesCount:       10,
//		UpdatedCodeNodesCount:       10,
//		UnchangedCodeNodesCount:     10,
//		NewCodeNodesCount:           10,
//		ExistingCodeNodesCount:      0,
//		ExistingLiteralIncludeCount: 0,
//		ExistingIoCodeBlockCount:    0,
//	}
//
//	unchangedCount := 5
//	updatedCount := 0
//	newCount := 0
//	removedCount := 0
//
//	updatedCounter := UpdateProjectReportForUpdatedCodeNodes(projectCounter, unchangedCount, updatedCount, newCount, removedCount)
//
//	wantUnchanged := 15
//	wantUpdated := 10
//	wantNew := 10
//	wantRemoved := 10
//
//	if updatedCounter.UnchangedCodeNodesCount != wantUnchanged {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.UnchangedCodeNodesCount, wantUnchanged)
//	}
//	if updatedCounter.UpdatedCodeNodesCount != wantUpdated {
//		t.Errorf("FAILED: got %d updated code nodes, want %d", updatedCounter.UpdatedCodeNodesCount, wantUpdated)
//	}
//	if updatedCounter.NewCodeNodesCount != wantNew {
//		t.Errorf("FAILED: got %d new code nodes, want %d", updatedCounter.NewCodeNodesCount, wantNew)
//	}
//	if updatedCounter.RemovedCodeNodesCount != wantRemoved {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.RemovedCodeNodesCount, wantRemoved)
//	}
//}
//
//func TestIncrementProjectCounterCorrectlyUpdatesUpdatedCounts(t *testing.T) {
//	projectCounter := types.ProjectCounts{
//		IncomingCodeNodesCount:      0,
//		IncomingLiteralIncludeCount: 0,
//		IncomingIoCodeBlockCount:    0,
//		RemovedCodeNodesCount:       10,
//		UpdatedCodeNodesCount:       10,
//		UnchangedCodeNodesCount:     10,
//		NewCodeNodesCount:           10,
//		ExistingCodeNodesCount:      0,
//		ExistingLiteralIncludeCount: 0,
//		ExistingIoCodeBlockCount:    0,
//	}
//
//	unchangedCount := 0
//	updatedCount := 5
//	newCount := 0
//	removedCount := 0
//
//	updatedCounter := UpdateProjectReportForUpdatedCodeNodes(projectCounter, unchangedCount, updatedCount, newCount, removedCount)
//
//	wantUnchanged := 10
//	wantUpdated := 15
//	wantNew := 10
//	wantRemoved := 10
//
//	if updatedCounter.UnchangedCodeNodesCount != wantUnchanged {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.UnchangedCodeNodesCount, wantUnchanged)
//	}
//	if updatedCounter.UpdatedCodeNodesCount != wantUpdated {
//		t.Errorf("FAILED: got %d updated code nodes, want %d", updatedCounter.UpdatedCodeNodesCount, wantUpdated)
//	}
//	if updatedCounter.NewCodeNodesCount != wantNew {
//		t.Errorf("FAILED: got %d new code nodes, want %d", updatedCounter.NewCodeNodesCount, wantNew)
//	}
//	if updatedCounter.RemovedCodeNodesCount != wantRemoved {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.RemovedCodeNodesCount, wantRemoved)
//	}
//}
//
//func TestIncrementProjectCounterCorrectlyUpdatesNewCounts(t *testing.T) {
//	projectCounter := types.ProjectCounts{
//		IncomingCodeNodesCount:      0,
//		IncomingLiteralIncludeCount: 0,
//		IncomingIoCodeBlockCount:    0,
//		RemovedCodeNodesCount:       10,
//		UpdatedCodeNodesCount:       10,
//		UnchangedCodeNodesCount:     10,
//		NewCodeNodesCount:           10,
//		ExistingCodeNodesCount:      0,
//		ExistingLiteralIncludeCount: 0,
//		ExistingIoCodeBlockCount:    0,
//	}
//
//	unchangedCount := 0
//	updatedCount := 0
//	newCount := 5
//	removedCount := 0
//
//	updatedCounter := UpdateProjectReportForUpdatedCodeNodes(projectCounter, unchangedCount, updatedCount, newCount, removedCount)
//
//	wantUnchanged := 10
//	wantUpdated := 10
//	wantNew := 15
//	wantRemoved := 10
//
//	if updatedCounter.UnchangedCodeNodesCount != wantUnchanged {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.UnchangedCodeNodesCount, wantUnchanged)
//	}
//	if updatedCounter.UpdatedCodeNodesCount != wantUpdated {
//		t.Errorf("FAILED: got %d updated code nodes, want %d", updatedCounter.UpdatedCodeNodesCount, wantUpdated)
//	}
//	if updatedCounter.NewCodeNodesCount != wantNew {
//		t.Errorf("FAILED: got %d new code nodes, want %d", updatedCounter.NewCodeNodesCount, wantNew)
//	}
//	if updatedCounter.RemovedCodeNodesCount != wantRemoved {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.RemovedCodeNodesCount, wantRemoved)
//	}
//}
//
//func TestIncrementProjectCounterCorrectlyUpdatesRemovedCounts(t *testing.T) {
//	projectCounter := types.ProjectCounts{
//		IncomingCodeNodesCount:      0,
//		IncomingLiteralIncludeCount: 0,
//		IncomingIoCodeBlockCount:    0,
//		RemovedCodeNodesCount:       10,
//		UpdatedCodeNodesCount:       10,
//		UnchangedCodeNodesCount:     10,
//		NewCodeNodesCount:           10,
//		ExistingCodeNodesCount:      0,
//		ExistingLiteralIncludeCount: 0,
//		ExistingIoCodeBlockCount:    0,
//	}
//
//	unchangedCount := 0
//	updatedCount := 0
//	newCount := 0
//	removedCount := 5
//
//	updatedCounter := UpdateProjectReportForUpdatedCodeNodes(projectCounter, unchangedCount, updatedCount, newCount, removedCount)
//
//	wantUnchanged := 10
//	wantUpdated := 10
//	wantNew := 10
//	wantRemoved := 15
//
//	if updatedCounter.UnchangedCodeNodesCount != wantUnchanged {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.UnchangedCodeNodesCount, wantUnchanged)
//	}
//	if updatedCounter.UpdatedCodeNodesCount != wantUpdated {
//		t.Errorf("FAILED: got %d updated code nodes, want %d", updatedCounter.UpdatedCodeNodesCount, wantUpdated)
//	}
//	if updatedCounter.NewCodeNodesCount != wantNew {
//		t.Errorf("FAILED: got %d new code nodes, want %d", updatedCounter.NewCodeNodesCount, wantNew)
//	}
//	if updatedCounter.RemovedCodeNodesCount != wantRemoved {
//		t.Errorf("FAILED: got %d unchanged code nodes, want %d", updatedCounter.RemovedCodeNodesCount, wantRemoved)
//	}
//}
