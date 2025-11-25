package procedures

import (
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	report, err := AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// Expected: 5 unique procedures (grouped by heading + content hash)
	// 1. Simple Procedure with Steps (1 unique)
	// 2. Procedure with Tabs (1 unique, appears in 3 selections: shell, nodejs, python)
	// 3. Composable Tutorial (1 unique, appears in 3 selections: driver/nodejs, driver/python, atlas-cli/none)
	// 4. Ordered List Procedure (1 unique)
	// 5. Procedure with Sub-steps (1 unique)
	if report.TotalProcedures != 5 {
		t.Errorf("Expected to find 5 unique procedures, but got %d", report.TotalProcedures)
	}

	// Expected total appearances: 1 + 3 + 3 + 1 + 1 = 9
	if report.TotalVariations != 9 {
		t.Errorf("Expected to find 9 total procedure appearances, but got %d", report.TotalVariations)
	}

	// Verify implementation types
	if report.ProceduresByType["procedure-directive"] != 4 {
		t.Errorf("Expected 4 procedure-directive implementations, got %d", report.ProceduresByType["procedure-directive"])
	}
	if report.ProceduresByType["ordered-list"] != 1 {
		t.Errorf("Expected 1 ordered-list implementation, got %d", report.ProceduresByType["ordered-list"])
	}

	t.Logf("Found %d unique procedures with %d total appearances", report.TotalProcedures, report.TotalVariations)
}

func TestAnalyzeFileNonExistent(t *testing.T) {
	_, err := AnalyzeFile("nonexistent-file.rst")
	if err == nil {
		t.Error("Expected error for nonexistent file, but got none")
	}
}

func TestAnalyzeFileDeterministic(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	// Run analysis multiple times to ensure deterministic results
	var reports []*AnalysisReport
	for i := 0; i < 5; i++ {
		report, err := AnalyzeFile(testFile)
		if err != nil {
			t.Fatalf("AnalyzeFile failed on iteration %d: %v", i, err)
		}
		reports = append(reports, report)
	}

	// Verify all runs produce the same counts
	for i := 1; i < len(reports); i++ {
		if reports[i].TotalProcedures != reports[0].TotalProcedures {
			t.Errorf("Iteration %d: TotalProcedures = %d, want %d (non-deterministic!)",
				i, reports[i].TotalProcedures, reports[0].TotalProcedures)
		}
		if reports[i].TotalVariations != reports[0].TotalVariations {
			t.Errorf("Iteration %d: TotalVariations = %d, want %d (non-deterministic!)",
				i, reports[i].TotalVariations, reports[0].TotalVariations)
		}
	}
}

func TestAnalyzeFileWithExpandIncludes(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-with-includes.rst"

	// The analyze command always expands includes (it calls AnalyzeFile which uses expandIncludes=true)
	// This test verifies that include expansion works correctly for detecting variations
	// in selected-content blocks within included files.

	// With expanding includes (the default behavior)
	reportExpand, err := AnalyzeFileWithOptions(testFile, true)
	if err != nil {
		t.Fatalf("AnalyzeFile with expand failed: %v", err)
	}

	// Should find 1 unique procedure (the composable tutorial)
	if reportExpand.TotalProcedures != 1 {
		t.Errorf("With expand: expected 1 unique procedure, got %d", reportExpand.TotalProcedures)
	}

	// Should detect 3 variations (driver/nodejs, driver/python, atlas-cli/none)
	// from the selected-content blocks in the included files
	if reportExpand.TotalVariations != 3 {
		t.Errorf("With expand: expected 3 appearances, got %d", reportExpand.TotalVariations)
	}

	// Verify the procedure has the expected variations
	if len(reportExpand.Procedures) > 0 {
		proc := reportExpand.Procedures[0]
		if len(proc.Variations) != 3 {
			t.Errorf("Expected 3 variations, got %d: %v", len(proc.Variations), proc.Variations)
		}

		// Verify expected selections
		expectedSelections := map[string]bool{
			"driver, nodejs":  true,
			"driver, python":  true,
			"atlas-cli, none": true,
		}
		for _, variation := range proc.Variations {
			if !expectedSelections[variation] {
				t.Errorf("Unexpected variation: %s", variation)
			}
		}
	}

	t.Logf("With expand: %d procedures, %d appearances", reportExpand.TotalProcedures, reportExpand.TotalVariations)
}

func TestPrintReport(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	report, err := AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// Test with default options (just summary)
	options := OutputOptions{}
	PrintReport(report, options)

	// Test with ListSummary
	options = OutputOptions{
		ListSummary: true,
	}
	PrintReport(report, options)

	// Test with ListAll and all details
	options = OutputOptions{
		ListAll:        true,
		Implementation: true,
		SubProcedures:  true,
		StepCount:      true,
	}
	PrintReport(report, options)

	// This test just ensures PrintReport doesn't panic
	// In a real test, we might capture stdout and verify the output
}

func TestProcedureAnalysisDetails(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	report, err := AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// Verify we have the expected procedures
	expectedTitles := map[string]bool{
		"Simple Procedure with Steps": true,
		"Procedure with Tabs":          true,
		"Composable Tutorial Example":  true,
		"Ordered List Procedure":       true,
		"Procedure with Sub-steps":     true,
	}

	for _, proc := range report.Procedures {
		if !expectedTitles[proc.Procedure.Title] {
			t.Errorf("Unexpected procedure title: %s", proc.Procedure.Title)
		}

		// Verify step counts are reasonable
		if proc.StepCount == 0 {
			t.Errorf("Procedure '%s' has 0 steps", proc.Procedure.Title)
		}
	}
}

func TestAnalyzeTabsWithProcedures(t *testing.T) {
	testFile := "../../../testdata/input-files/source/tabs-with-procedures.rst"

	report, err := AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// Expected: 1 unique procedure (grouped as a tab set)
	// with 3 appearances (macos, ubuntu, windows)
	if report.TotalProcedures != 1 {
		t.Errorf("Expected to find 1 unique procedure, but got %d", report.TotalProcedures)
	}

	// Expected total appearances: 3 (one for each tab)
	if report.TotalVariations != 3 {
		t.Errorf("Expected to find 3 total procedure appearances, but got %d", report.TotalVariations)
	}

	// Verify the procedure has the expected variations
	if len(report.Procedures) > 0 {
		proc := report.Procedures[0]
		if len(proc.Variations) != 3 {
			t.Errorf("Expected 3 variations, got %d: %v", len(proc.Variations), proc.Variations)
		}

		// Verify expected tab IDs
		expectedTabs := map[string]bool{
			"macos":   true,
			"ubuntu":  true,
			"windows": true,
		}
		for _, variation := range proc.Variations {
			if !expectedTabs[variation] {
				t.Errorf("Unexpected variation: %s", variation)
			}
		}
	}

	t.Logf("Found %d unique procedures with %d total appearances", report.TotalProcedures, report.TotalVariations)
}

