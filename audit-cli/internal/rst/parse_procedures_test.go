package rst

import (
	"path/filepath"
	"testing"
)

func TestParseProceduresWithOptions(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Expected: 5 unique procedures
	if len(procedures) != 5 {
		t.Errorf("Expected 5 procedures, got %d", len(procedures))
	}

	// Verify each procedure has steps
	for i, proc := range procedures {
		if len(proc.Steps) == 0 {
			t.Errorf("Procedure %d (%s) has no steps", i, proc.Title)
		}
	}

	t.Logf("Found %d procedures", len(procedures))
}

func TestParseProceduresDeterministic(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	// Parse multiple times to ensure deterministic results
	var allProcedures [][]Procedure
	for i := 0; i < 5; i++ {
		procedures, err := ParseProceduresWithOptions(testFile, false)
		if err != nil {
			t.Fatalf("ParseProceduresWithOptions failed on iteration %d: %v", i, err)
		}
		allProcedures = append(allProcedures, procedures)
	}

	// Verify all runs produce the same count
	for i := 1; i < len(allProcedures); i++ {
		if len(allProcedures[i]) != len(allProcedures[0]) {
			t.Errorf("Iteration %d: found %d procedures, want %d (non-deterministic!)",
				i, len(allProcedures[i]), len(allProcedures[0]))
		}
	}

	// Verify procedure titles are in the same order
	for i := 1; i < len(allProcedures); i++ {
		for j := 0; j < len(allProcedures[0]); j++ {
			if allProcedures[i][j].Title != allProcedures[0][j].Title {
				t.Errorf("Iteration %d, procedure %d: title = %s, want %s (non-deterministic!)",
					i, j, allProcedures[i][j].Title, allProcedures[0][j].Title)
			}
		}
	}
}

func TestComputeProcedureContentHash(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	// Parse the file multiple times
	var hashes []string
	for i := 0; i < 5; i++ {
		procedures, err := ParseProceduresWithOptions(testFile, false)
		if err != nil {
			t.Fatalf("ParseProceduresWithOptions failed on iteration %d: %v", i, err)
		}

		if len(procedures) > 0 {
			hash := computeProcedureContentHash(&procedures[0])
			hashes = append(hashes, hash)
		}
	}

	// Verify all hashes are identical (deterministic)
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Iteration %d: hash = %s, want %s (non-deterministic!)", i, hashes[i], hashes[0])
		}
	}

	t.Logf("Content hash is deterministic: %s", hashes[0])
}

func TestParseProcedureDirective(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Find a procedure directive (not ordered list)
	var procedureDirective *Procedure
	for i := range procedures {
		if procedures[i].Title == "Simple Procedure with Steps" {
			procedureDirective = &procedures[i]
			break
		}
	}

	if procedureDirective == nil {
		t.Fatal("Could not find 'Simple Procedure with Steps'")
	}

	// Verify it has the expected number of steps
	if len(procedureDirective.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(procedureDirective.Steps))
	}

	// Verify step titles
	expectedTitles := []string{
		"Create a database connection",
		"Insert a document",
		"Close the connection",
	}

	for i, expectedTitle := range expectedTitles {
		if i >= len(procedureDirective.Steps) {
			t.Errorf("Missing step %d", i)
			continue
		}
		if procedureDirective.Steps[i].Title != expectedTitle {
			t.Errorf("Step %d: title = %q, want %q", i, procedureDirective.Steps[i].Title, expectedTitle)
		}
	}

	t.Logf("Procedure directive parsed correctly with %d steps", len(procedureDirective.Steps))
}

func TestParseOrderedListProcedure(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Find the ordered list procedure
	var orderedListProc *Procedure
	for i := range procedures {
		if procedures[i].Title == "Ordered List Procedure" {
			orderedListProc = &procedures[i]
			break
		}
	}

	if orderedListProc == nil {
		t.Fatal("Could not find 'Ordered List Procedure'")
	}

	// Verify it has the expected number of steps
	if len(orderedListProc.Steps) != 4 {
		t.Errorf("Expected 4 steps, got %d", len(orderedListProc.Steps))
	}

	t.Logf("Ordered list procedure parsed correctly with %d steps", len(orderedListProc.Steps))
}

func TestAbsolutePath(t *testing.T) {
	// Test with relative path
	relPath := "../../testdata/input-files/source/procedure-test.rst"
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Parse with absolute path
	procedures, err := ParseProceduresWithOptions(absPath, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions with absolute path failed: %v", err)
	}

	if len(procedures) != 5 {
		t.Errorf("Expected 5 procedures with absolute path, got %d", len(procedures))
	}

	t.Logf("Successfully parsed with absolute path: %s", absPath)
}

func TestContinuationMarkers(t *testing.T) {
	testFile := "../../testdata/input-files/source/continuation-marker-test.rst"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	if len(procedures) != 2 {
		t.Fatalf("Expected 2 procedures, got %d", len(procedures))
	}

	// Test lettered list with continuation markers
	letteredProc := procedures[0]
	if letteredProc.Title != "Lettered List with Continuation" {
		t.Errorf("Expected title 'Lettered List with Continuation', got '%s'", letteredProc.Title)
	}

	if len(letteredProc.Steps) != 3 {
		t.Fatalf("Expected 3 steps in lettered list, got %d", len(letteredProc.Steps))
	}

	// Verify step titles (note: regular list items don't include the marker in the title)
	// Only continuation markers get the computed marker prepended
	expectedTitles := []string{"First step", "b. Second step", "c. Third step"}
	for i, step := range letteredProc.Steps {
		if step.Title != expectedTitles[i] {
			t.Errorf("Step %d: expected title '%s', got '%s'", i, expectedTitles[i], step.Title)
		}
	}

	// Test numbered list with continuation markers
	numberedProc := procedures[1]
	if numberedProc.Title != "Numbered List with Continuation" {
		t.Errorf("Expected title 'Numbered List with Continuation', got '%s'", numberedProc.Title)
	}

	if len(numberedProc.Steps) != 4 {
		t.Fatalf("Expected 4 steps in numbered list, got %d", len(numberedProc.Steps))
	}

	// Verify step titles (note: regular list items don't include the marker in the title)
	// Only continuation markers get the computed marker prepended
	expectedNumberedTitles := []string{"First step", "2. Second step", "3. Third step", "4. Fourth step"}
	for i, step := range numberedProc.Steps {
		if step.Title != expectedNumberedTitles[i] {
			t.Errorf("Step %d: expected title '%s', got '%s'", i, expectedNumberedTitles[i], step.Title)
		}
	}

	t.Logf("Continuation markers parsed correctly")
}

func TestHierarchicalProcedure(t *testing.T) {
	testFile := "../../testdata/input-files/source/rotate-key-sharded-cluster.txt"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Should parse as 1 procedure (not 10 separate procedures)
	if len(procedures) != 1 {
		t.Fatalf("Expected 1 procedure, got %d", len(procedures))
	}

	proc := procedures[0]
	if proc.Title != "Procedure" {
		t.Errorf("Expected title 'Procedure', got '%s'", proc.Title)
	}

	// Should have 4 top-level steps (the numbered headings)
	if len(proc.Steps) != 4 {
		t.Fatalf("Expected 4 steps, got %d", len(proc.Steps))
	}

	// Verify step titles match the numbered headings
	expectedStepTitles := []string{
		"1. Modify the Keyfile to Include Old and New Keys",
		"2. Restart Each Member",
		"3. Update Keyfile Content to the New Key Only",
		"4. Restart Each Member",
	}

	for i, step := range proc.Steps {
		if step.Title != expectedStepTitles[i] {
			t.Errorf("Step %d: expected title '%s', got '%s'", i, expectedStepTitles[i], step.Title)
		}
	}

	// Verify HasSubSteps is set
	if !proc.HasSubSteps {
		t.Error("Expected HasSubSteps to be true")
	}

	// Verify that step 2 has sub-steps (the ordered lists)
	step2 := proc.Steps[1]
	if len(step2.SubSteps) == 0 {
		t.Error("Expected step 2 to have sub-steps")
	}

	t.Logf("Hierarchical procedure parsed correctly with %d steps", len(proc.Steps))
}

func TestSubProcedureDetection(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Find the "Procedure with Sub-steps" procedure
	var subStepProc *Procedure
	for i := range procedures {
		if procedures[i].Title == "Procedure with Sub-steps" {
			subStepProc = &procedures[i]
			break
		}
	}

	if subStepProc == nil {
		t.Fatal("Could not find 'Procedure with Sub-steps'")
	}

	// Verify HasSubSteps is set
	if !subStepProc.HasSubSteps {
		t.Error("Expected HasSubSteps to be true for 'Procedure with Sub-steps'")
	}

	// Verify at least one step has sub-steps
	hasSubSteps := false
	for _, step := range subStepProc.Steps {
		if len(step.SubSteps) > 0 {
			hasSubSteps = true
			break
		}
	}

	if !hasSubSteps {
		t.Error("Expected at least one step to have sub-steps")
	}

	t.Logf("Sub-procedure detection working correctly")
}

func TestSubProcedureListTypes(t *testing.T) {
	testFile := "../../testdata/input-files/source/rotate-key-sharded-cluster.txt"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Find the hierarchical procedure
	if len(procedures) == 0 {
		t.Fatal("Expected at least one procedure")
	}

	proc := procedures[0]

	// Verify it has steps with sub-procedures
	if len(proc.Steps) < 2 {
		t.Fatalf("Expected at least 2 steps, got %d", len(proc.Steps))
	}

	// Check step 2 (index 1) which should have sub-procedures
	step := proc.Steps[1]
	if len(step.SubProcedures) == 0 {
		t.Fatal("Expected step 2 to have sub-procedures")
	}

	// Verify all sub-procedures have the correct list type
	for i, subProc := range step.SubProcedures {
		if subProc.ListType != "lettered" {
			t.Errorf("Sub-procedure %d: expected list type 'lettered', got '%s'", i+1, subProc.ListType)
		}

		if len(subProc.Steps) == 0 {
			t.Errorf("Sub-procedure %d: expected at least one step", i+1)
		}

		// Verify steps are present
		t.Logf("Sub-procedure %d has %d steps with list type '%s'", i+1, len(subProc.Steps), subProc.ListType)
	}

	// Verify backward compatibility - SubSteps should still be populated
	if len(step.SubSteps) == 0 {
		t.Error("Expected SubSteps to be populated for backward compatibility")
	}

	// Count total steps across all sub-procedures
	totalSteps := 0
	for _, subProc := range step.SubProcedures {
		totalSteps += len(subProc.Steps)
	}

	// Verify SubSteps has the same total count
	if len(step.SubSteps) != totalSteps {
		t.Errorf("Expected SubSteps to have %d steps (flattened), got %d", totalSteps, len(step.SubSteps))
	}

	t.Logf("Sub-procedure list types tracked correctly: %d sub-procedures with %d total steps",
		len(step.SubProcedures), totalSteps)
}

func TestNumberedHeadingDetection(t *testing.T) {
	tests := []struct {
		heading  string
		expected bool
	}{
		{"1. First Step", true},
		{"2. Second Step", true},
		{"10. Tenth Step", true},
		{"123. Large Number", true},
		{"Step 1", false},
		{"1 First Step", false},
		{"a. Lettered Step", false},
		{"Procedure", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isNumberedHeading(tt.heading)
		if result != tt.expected {
			t.Errorf("isNumberedHeading(%q) = %v, want %v", tt.heading, result, tt.expected)
		}
	}
}
