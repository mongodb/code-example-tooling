package procedures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	variations, err := ParseFile(testFile, "", false)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Expected: 5 unique procedures (one file per unique procedure)
	// 1. Simple Procedure with Steps (1 unique)
	// 2. Procedure with Tabs (1 unique, appears in 3 selections)
	// 3. Composable Tutorial (1 unique, appears in 3 selections)
	// 4. Ordered List Procedure (1 unique)
	// 5. Procedure with Sub-steps (1 unique)
	if len(variations) != 5 {
		t.Errorf("Expected to find 5 unique procedures, but got %d", len(variations))
	}

	// Verify that procedures with multiple selections have them listed
	foundMultiSelection := false
	for _, v := range variations {
		if strings.Contains(v.VariationName, ";") {
			foundMultiSelection = true
			// Count selections
			selections := strings.Split(v.VariationName, "; ")
			if len(selections) < 2 {
				t.Errorf("Procedure with semicolon should have multiple selections, got: %s", v.VariationName)
			}
		}
	}
	if !foundMultiSelection {
		t.Error("Expected to find at least one procedure with multiple selections")
	}

	t.Logf("Found %d unique procedures", len(variations))
}

func TestParseFileWithFilter(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	variations, err := ParseFile(testFile, "python", false)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Should only get procedures that appear in the "python" selection
	// Expected: 1 procedure (the one with tabs that includes python)
	if len(variations) != 1 {
		t.Errorf("Expected 1 procedure matching 'python', got %d", len(variations))
	}

	// Verify the variation name contains "python"
	for _, v := range variations {
		if !strings.Contains(v.VariationName, "python") {
			t.Errorf("Expected variation to contain 'python', got: %s", v.VariationName)
		}
	}

	t.Logf("Found %d procedure(s) matching 'python'", len(variations))
}

func TestParseFileDeterministic(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	// Run parsing multiple times to ensure deterministic results
	var allVariations [][]ProcedureVariation
	for i := 0; i < 5; i++ {
		variations, err := ParseFile(testFile, "", false)
		if err != nil {
			t.Fatalf("ParseFile failed on iteration %d: %v", i, err)
		}
		allVariations = append(allVariations, variations)
	}

	// Verify all runs produce the same count
	for i := 1; i < len(allVariations); i++ {
		if len(allVariations[i]) != len(allVariations[0]) {
			t.Errorf("Iteration %d: found %d procedures, want %d (non-deterministic!)",
				i, len(allVariations[i]), len(allVariations[0]))
		}
	}

	// Verify filenames are consistent
	for i := 1; i < len(allVariations); i++ {
		for j := 0; j < len(allVariations[0]); j++ {
			if allVariations[i][j].OutputFile != allVariations[0][j].OutputFile {
				t.Errorf("Iteration %d, procedure %d: filename = %s, want %s (non-deterministic!)",
					i, j, allVariations[i][j].OutputFile, allVariations[0][j].OutputFile)
			}
		}
	}
}

func TestWriteVariation(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"
	outputDir := t.TempDir()

	variations, err := ParseFile(testFile, "", false)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(variations) == 0 {
		t.Fatal("No variations found to test writing")
	}

	// Write the first variation
	err = WriteVariation(variations[0], outputDir, false)
	if err != nil {
		t.Fatalf("WriteVariation failed: %v", err)
	}

	// Check that the file was created
	outputPath := filepath.Join(outputDir, variations[0].OutputFile)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected output file %s to exist, but it doesn't", outputPath)
	}

	// Verify the file has content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(content) == 0 {
		t.Error("Output file is empty")
	}

	t.Logf("Successfully wrote variation to %s (%d bytes)", outputPath, len(content))
}

func TestWriteAllVariations(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"
	outputDir := t.TempDir()

	variations, err := ParseFile(testFile, "", false)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	filesWritten, err := WriteAllVariations(variations, outputDir, false, false)
	if err != nil {
		t.Fatalf("WriteAllVariations failed: %v", err)
	}

	if filesWritten != len(variations) {
		t.Errorf("Expected to write %d files, but wrote %d", len(variations), filesWritten)
	}

	// Verify all files exist
	for _, v := range variations {
		outputPath := filepath.Join(outputDir, v.OutputFile)
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s to exist, but it doesn't", outputPath)
		}
	}

	t.Logf("Successfully wrote %d files", filesWritten)
}

func TestParseFileWithIncludes(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-with-includes.rst"

	// With expanding includes, should find 1 unique procedure appearing in 3 selections
	// The selected-content blocks are in the included files (install-deps.rst, connect.rst, operations.rst)
	variationsExpand, err := ParseFile(testFile, "", true)
	if err != nil {
		t.Fatalf("ParseFile with expand failed: %v", err)
	}

	if len(variationsExpand) != 1 {
		t.Errorf("Expected 1 unique procedure with expanding includes, got %d", len(variationsExpand))
	}

	// Verify the expanded version has multiple selections
	if len(variationsExpand) > 0 {
		// Should detect 3 selections from the included selected-content blocks
		selectionsExpand := strings.Split(variationsExpand[0].VariationName, "; ")
		if len(selectionsExpand) != 3 {
			t.Errorf("Expected 3 selections with expand, got %d: %v", len(selectionsExpand), selectionsExpand)
		}

		// Verify expected selections
		expectedSelections := map[string]bool{
			"driver, nodejs":  true,
			"driver, python":  true,
			"atlas-cli, none": true,
		}
		for _, sel := range selectionsExpand {
			if !expectedSelections[sel] {
				t.Errorf("Unexpected selection: %s", sel)
			}
		}
	}

	t.Logf("With expand: %d procedures with %d selections",
		len(variationsExpand),
		len(strings.Split(variationsExpand[0].VariationName, "; ")))
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Procedure", "simple-procedure"},
		{"Connect to Cluster", "connect-to-cluster"},
		{"driver, nodejs", "driver-nodejs"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"Special!@#Characters", "specialcharacters"},
		{"  Leading and Trailing  ", "leading-and-trailing"},
		{"UPPERCASE", "uppercase"},
		{"Mixed_Case-With.Dots", "mixed-casewithdots"}, // Dots are removed, not converted to hyphens
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	variations, err := ParseFile(testFile, "", false)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify all filenames are unique
	filenameMap := make(map[string]bool)
	for _, v := range variations {
		if filenameMap[v.OutputFile] {
			t.Errorf("Duplicate filename generated: %s", v.OutputFile)
		}
		filenameMap[v.OutputFile] = true

		// Verify filename format
		if !strings.HasSuffix(v.OutputFile, ".rst") {
			t.Errorf("Filename should end with .rst: %s", v.OutputFile)
		}

		// Verify filename contains a hash (6 characters before .rst)
		parts := strings.Split(v.OutputFile, "_")
		if len(parts) < 2 {
			t.Errorf("Filename should contain at least heading and hash: %s", v.OutputFile)
		}
	}

	t.Logf("Generated %d unique filenames", len(filenameMap))
}

func TestDryRun(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"
	outputDir := t.TempDir()

	variations, err := ParseFile(testFile, "", false)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Write with dry run enabled
	filesWritten, err := WriteAllVariations(variations, outputDir, true, false)
	if err != nil {
		t.Fatalf("WriteAllVariations with dry run failed: %v", err)
	}

	// Should report files that would be written
	if filesWritten != len(variations) {
		t.Errorf("Dry run should report %d files, but reported %d", len(variations), filesWritten)
	}

	// Verify no files were actually created
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("Dry run should not create files, but found %d files", len(entries))
	}

	t.Logf("Dry run correctly reported %d files without writing them", filesWritten)
}

func TestContentHash(t *testing.T) {
	testFile := "../../../testdata/input-files/source/procedure-test.rst"

	// Parse the file multiple times
	var hashes []string
	for i := 0; i < 3; i++ {
		variations, err := ParseFile(testFile, "", false)
		if err != nil {
			t.Fatalf("ParseFile failed on iteration %d: %v", i, err)
		}

		// Extract hash from first filename (last 6 chars before .rst)
		if len(variations) > 0 {
			filename := variations[0].OutputFile
			// Remove .rst extension
			nameWithoutExt := strings.TrimSuffix(filename, ".rst")
			// Get last part after last underscore (the hash)
			parts := strings.Split(nameWithoutExt, "_")
			hash := parts[len(parts)-1]
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

func TestParseTabsWithProcedures(t *testing.T) {
	testFile := "../../../testdata/input-files/source/tabs-with-procedures.rst"

	variations, err := ParseFile(testFile, "", false)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Expected: 3 unique procedures (one for each tab: macos, ubuntu, windows)
	if len(variations) != 3 {
		t.Errorf("Expected to find 3 unique procedures, but got %d", len(variations))
	}

	// Verify each procedure has only one selection (its specific tab)
	for _, v := range variations {
		if strings.Contains(v.VariationName, ";") {
			t.Errorf("Expected each procedure to have only one selection, got: %s", v.VariationName)
		}

		// Verify the selection is one of the expected tabs
		expectedTabs := map[string]bool{
			"macos":   true,
			"ubuntu":  true,
			"windows": true,
		}
		if !expectedTabs[v.VariationName] {
			t.Errorf("Unexpected variation name: %s", v.VariationName)
		}
	}

	// Verify all three tabs are present
	foundTabs := make(map[string]bool)
	for _, v := range variations {
		foundTabs[v.VariationName] = true
	}
	if len(foundTabs) != 3 {
		t.Errorf("Expected to find all 3 tabs, but got: %v", foundTabs)
	}

	t.Logf("Found %d unique procedures from tabs", len(variations))
}
