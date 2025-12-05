package rst

import (
	"testing"
)

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
