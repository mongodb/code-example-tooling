package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// DependencyFile represents a found dependency management file
type DependencyFile struct {
	Path     string
	Type     FileType
	Dir      string
	Filename string
}

// FileType represents the type of dependency management file
type FileType string

const (
	PackageJSON    FileType = "package.json"
	PomXML         FileType = "pom.xml"
	RequirementsTxt FileType = "requirements.txt"
	GoMod          FileType = "go.mod"
	CsProj         FileType = ".csproj"
)

var dependencyFiles = map[string]FileType{
	"package.json":     PackageJSON,
	"pom.xml":          PomXML,
	"requirements.txt": RequirementsTxt,
	"go.mod":           GoMod,
}

// Scanner handles scanning for dependency files
type Scanner struct {
	startPath    string
	ignorePaths  []string
}

// New creates a new Scanner with default ignore patterns
func New(startPath string) *Scanner {
	return &Scanner{
		startPath: startPath,
		ignorePaths: []string{"node_modules", ".git", "vendor", "target", "dist", "build"},
	}
}

// NewWithIgnorePaths creates a new Scanner with custom ignore patterns
func NewWithIgnorePaths(startPath string, ignorePaths []string) *Scanner {
	// Always include common directories that should be ignored
	defaultIgnores := []string{"node_modules", ".git", "vendor", "target", "dist", "build"}

	// Merge default ignores with custom ones
	allIgnores := append(defaultIgnores, ignorePaths...)

	return &Scanner{
		startPath:   startPath,
		ignorePaths: allIgnores,
	}
}

// Scan finds all dependency management files starting from the given path
func (s *Scanner) Scan() ([]DependencyFile, error) {
	var depFiles []DependencyFile

	// Check if the start path exists
	info, err := os.Stat(s.startPath)
	if err != nil {
		return nil, err
	}

	// If it's a file, check if it's a dependency file
	if !info.IsDir() {
		if depFile := s.checkFile(s.startPath); depFile != nil {
			return []DependencyFile{*depFile}, nil
		}
		return depFiles, nil
	}

	// If it's a directory, walk through it
	err = filepath.Walk(s.startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Check if this directory should be ignored
			dirName := filepath.Base(path)
			for _, ignore := range s.ignorePaths {
				if dirName == ignore {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if depFile := s.checkFile(path); depFile != nil {
			depFiles = append(depFiles, *depFile)
		}

		return nil
	})

	return depFiles, err
}

// checkFile checks if a file is a dependency management file
func (s *Scanner) checkFile(path string) *DependencyFile {
	filename := filepath.Base(path)
	dir := filepath.Dir(path)

	// Check for exact matches
	if fileType, ok := dependencyFiles[filename]; ok {
		return &DependencyFile{
			Path:     path,
			Type:     fileType,
			Dir:      dir,
			Filename: filename,
		}
	}

	// Check for .csproj files
	if strings.HasSuffix(filename, ".csproj") {
		return &DependencyFile{
			Path:     path,
			Type:     CsProj,
			Dir:      dir,
			Filename: filename,
		}
	}

	return nil
}

// IsDependencyFile checks if the given path is a dependency management file
func IsDependencyFile(path string) bool {
	filename := filepath.Base(path)

	// Check for exact matches
	if _, ok := dependencyFiles[filename]; ok {
		return true
	}

	// Check for .csproj files
	return strings.HasSuffix(filename, ".csproj")
}

