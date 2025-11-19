package checker

import (
	"fmt"

	"dependency-manager/internal/scanner"
)

// UpdateMode defines how dependencies should be updated
type UpdateMode int

const (
	// DryRun only checks for updates without making changes
	DryRun UpdateMode = iota
	// UpdateFile updates the dependency file but doesn't install
	UpdateFile
	// FullUpdate updates the file and installs dependencies
	FullUpdate
)

// DependencyUpdate represents an available update for a dependency
type DependencyUpdate struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
	UpdateType     string // "major", "minor", "patch"
}

// CheckResult contains the results of checking a dependency file
type CheckResult struct {
	FilePath string
	FileType scanner.FileType
	Updates  []DependencyUpdate
	Error    error
}

// Checker interface defines methods for checking and updating dependencies
type Checker interface {
	// Check returns available updates for dependencies
	Check(filePath string, directOnly bool) ([]DependencyUpdate, error)

	// Update updates the dependency file based on the mode
	Update(filePath string, mode UpdateMode, directOnly bool) error

	// GetFileType returns the file type this checker handles
	GetFileType() scanner.FileType
}

// Registry holds all available checkers
type Registry struct {
	checkers map[scanner.FileType]Checker
}

// NewRegistry creates a new checker registry
func NewRegistry() *Registry {
	return &Registry{
		checkers: make(map[scanner.FileType]Checker),
	}
}

// Register adds a checker to the registry
func (r *Registry) Register(checker Checker) {
	r.checkers[checker.GetFileType()] = checker
}

// GetChecker returns the appropriate checker for a file type
func (r *Registry) GetChecker(fileType scanner.FileType) (Checker, error) {
	checker, ok := r.checkers[fileType]
	if !ok {
		return nil, fmt.Errorf("no checker registered for file type: %s", fileType)
	}
	return checker, nil
}

// CheckFile checks a single dependency file for updates
func (r *Registry) CheckFile(depFile scanner.DependencyFile, directOnly bool) CheckResult {
	checker, err := r.GetChecker(depFile.Type)
	if err != nil {
		return CheckResult{
			FilePath: depFile.Path,
			FileType: depFile.Type,
			Error:    err,
		}
	}

	updates, err := checker.Check(depFile.Path, directOnly)
	return CheckResult{
		FilePath: depFile.Path,
		FileType: depFile.Type,
		Updates:  updates,
		Error:    err,
	}
}

// UpdateFile updates a single dependency file
func (r *Registry) UpdateFile(depFile scanner.DependencyFile, mode UpdateMode, directOnly bool) error {
	checker, err := r.GetChecker(depFile.Type)
	if err != nil {
		return err
	}

	return checker.Update(depFile.Path, mode, directOnly)
}

