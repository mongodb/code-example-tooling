package pathresolver

import (
	"path/filepath"
	"testing"
)

func TestFindSourceDirectory(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		wantContains string
		wantErr     bool
	}{
		{
			name:         "versioned project file",
			filePath:     "../../testdata/compare/product/v8.0/source/includes/example.rst",
			wantContains: "testdata/compare/product/v8.0/source",
			wantErr:      false,
		},
		{
			name:         "non-versioned project file",
			filePath:     "../../testdata/compare/product/manual/source/includes/example.rst",
			wantContains: "testdata/compare/product/manual/source",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindSourceDirectory(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindSourceDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Check that the path contains the expected substring
				if !filepath.IsAbs(got) {
					t.Errorf("FindSourceDirectory() returned relative path: %v", got)
				}
				if !filepath.HasPrefix(got, "/") {
					t.Errorf("FindSourceDirectory() returned non-absolute path: %v", got)
				}
				// Check that it ends with the expected path
				if !filepath.HasPrefix(got, "/") || !filepath.HasPrefix(filepath.Clean(got), "/") {
					t.Errorf("FindSourceDirectory() = %v, should be absolute", got)
				}
			}
		})
	}
}

func TestDetectProjectInfo(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		wantVersion string
		wantVersioned bool
		wantErr     bool
	}{
		{
			name:          "versioned project v8.0",
			filePath:      "../../testdata/compare/product/v8.0/source/includes/example.rst",
			wantVersion:   "v8.0",
			wantVersioned: true,
			wantErr:       false,
		},
		{
			name:          "versioned project manual",
			filePath:      "../../testdata/compare/product/manual/source/includes/example.rst",
			wantVersion:   "manual",
			wantVersioned: true,
			wantErr:       false,
		},
		{
			name:          "versioned project upcoming",
			filePath:      "../../testdata/compare/product/upcoming/source/includes/example.rst",
			wantVersion:   "upcoming",
			wantVersioned: true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectProjectInfo(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectProjectInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Version != tt.wantVersion {
					t.Errorf("DetectProjectInfo() Version = %v, want %v", got.Version, tt.wantVersion)
				}
				if got.IsVersioned != tt.wantVersioned {
					t.Errorf("DetectProjectInfo() IsVersioned = %v, want %v", got.IsVersioned, tt.wantVersioned)
				}
				if got.SourceDir == "" {
					t.Errorf("DetectProjectInfo() SourceDir is empty")
				}
				if got.ProductDir == "" {
					t.Errorf("DetectProjectInfo() ProductDir is empty")
				}
			}
		})
	}
}

func TestResolveVersionPaths(t *testing.T) {
	// Get absolute path to test data
	testFile := "../../testdata/compare/product/v8.0/source/includes/example.rst"
	absTestFile, _ := filepath.Abs(testFile)

	// Get product directory (parent of v8.0)
	sourceDir := filepath.Dir(absTestFile)                    // .../includes
	sourceDir = filepath.Dir(sourceDir)                       // .../source
	versionDir := filepath.Dir(sourceDir)                     // .../v8.0
	productDir := filepath.Dir(versionDir)                    // .../product

	versions := []string{"v8.0", "manual", "upcoming"}

	got, err := ResolveVersionPaths(absTestFile, productDir, versions)
	if err != nil {
		t.Fatalf("ResolveVersionPaths() error = %v", err)
	}

	if len(got) != 3 {
		t.Errorf("ResolveVersionPaths() returned %d paths, want 3", len(got))
	}

	// Check that each version path is constructed correctly
	for i, vp := range got {
		if vp.Version != versions[i] {
			t.Errorf("VersionPath[%d].Version = %v, want %v", i, vp.Version, versions[i])
		}
		expectedPath := filepath.Join(productDir, versions[i], "source", "includes", "example.rst")
		if vp.FilePath != expectedPath {
			t.Errorf("VersionPath[%d].FilePath = %v, want %v", i, vp.FilePath, expectedPath)
		}
	}
}

func TestExtractVersionFromPath(t *testing.T) {
	testFile := "../../testdata/compare/product/v8.0/source/includes/example.rst"
	absTestFile, _ := filepath.Abs(testFile)

	// Get product directory (parent of v8.0)
	sourceDir := filepath.Dir(absTestFile)                    // .../includes
	sourceDir = filepath.Dir(sourceDir)                       // .../source
	versionDir := filepath.Dir(sourceDir)                     // .../v8.0
	productDir := filepath.Dir(versionDir)                    // .../product

	got, err := ExtractVersionFromPath(absTestFile, productDir)
	if err != nil {
		t.Fatalf("ExtractVersionFromPath() error = %v", err)
	}

	want := "v8.0"
	if got != want {
		t.Errorf("ExtractVersionFromPath() = %v, want %v", got, want)
	}
}

func TestResolveRelativeToSource(t *testing.T) {
	sourceDir := "/path/to/manual/v8.0/source"
	
	tests := []struct {
		name         string
		relativePath string
		want         string
	}{
		{
			name:         "path with leading slash",
			relativePath: "/includes/file.rst",
			want:         "/path/to/manual/v8.0/source/includes/file.rst",
		},
		{
			name:         "path without leading slash",
			relativePath: "includes/file.rst",
			want:         "/path/to/manual/v8.0/source/includes/file.rst",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveRelativeToSource(sourceDir, tt.relativePath)
			if err != nil {
				t.Errorf("ResolveRelativeToSource() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveRelativeToSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

