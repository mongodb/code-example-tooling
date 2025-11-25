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

func TestGetProcedureVariations(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Find the procedure with tabs (should have 3 variations)
	var tabProcedure *Procedure
	for i := range procedures {
		if procedures[i].Title == "Procedure with Tabs" {
			tabProcedure = &procedures[i]
			break
		}
	}

	if tabProcedure == nil {
		t.Fatal("Could not find 'Procedure with Tabs'")
	}

	variations := GetProcedureVariations(*tabProcedure)
	if len(variations) != 3 {
		t.Errorf("Expected 3 variations for tabbed procedure, got %d: %v", len(variations), variations)
	}

	// Verify variations contain expected tabids
	expectedTabids := map[string]bool{"shell": true, "nodejs": true, "python": true}
	for _, variation := range variations {
		if !expectedTabids[variation] {
			t.Errorf("Unexpected variation: %s", variation)
		}
	}

	t.Logf("Found %d variations: %v", len(variations), variations)
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

func TestParseProceduresWithExpandIncludes(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-with-includes.rst"

	// With expanding includes - should detect selected-content blocks in included files
	proceduresExpand, err := ParseProceduresWithOptions(testFile, true)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions with expand failed: %v", err)
	}

	// Should find 1 unique procedure
	if len(proceduresExpand) != 1 {
		t.Errorf("With expand: expected 1 procedure, got %d", len(proceduresExpand))
	}

	// Should detect 3 variations from the selected-content blocks in the included files
	if len(proceduresExpand) > 0 {
		variations := GetProcedureVariations(proceduresExpand[0])
		if len(variations) != 3 {
			t.Errorf("With expand: expected 3 variations, got %d: %v", len(variations), variations)
		}

		// Verify expected selections
		expectedSelections := map[string]bool{
			"driver, nodejs":  true,
			"driver, python":  true,
			"atlas-cli, none": true,
		}
		for _, variation := range variations {
			if !expectedSelections[variation] {
				t.Errorf("Unexpected variation: %s", variation)
			}
		}
	}

	t.Logf("With expand: %d procedures with %d variations",
		len(proceduresExpand), len(GetProcedureVariations(proceduresExpand[0])))
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

func TestParseComposableTutorial(t *testing.T) {
	testFile := "../../testdata/input-files/source/procedure-test.rst"

	procedures, err := ParseProceduresWithOptions(testFile, false)
	if err != nil {
		t.Fatalf("ParseProceduresWithOptions failed: %v", err)
	}

	// Find the composable tutorial
	var composableProc *Procedure
	for i := range procedures {
		if procedures[i].Title == "Composable Tutorial Example" {
			composableProc = &procedures[i]
			break
		}
	}

	if composableProc == nil {
		t.Fatal("Could not find 'Composable Tutorial Example'")
	}

	// Verify it has variations
	variations := GetProcedureVariations(*composableProc)
	if len(variations) != 3 {
		t.Errorf("Expected 3 variations, got %d: %v", len(variations), variations)
	}

	// Verify expected selections
	expectedSelections := map[string]bool{
		"driver, nodejs":  true,
		"driver, python":  true,
		"atlas-cli, none": true,
	}

	for _, variation := range variations {
		if !expectedSelections[variation] {
			t.Errorf("Unexpected variation: %s", variation)
		}
	}

	t.Logf("Composable tutorial parsed correctly with %d variations", len(variations))
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

