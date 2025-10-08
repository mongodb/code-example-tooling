package services

import (
	"sync"

	"github.com/mongodb/code-example-tooling/code-copier/types"
)

// FileStateService manages the state of files to upload and deprecate
type FileStateService interface {
	GetFilesToUpload() map[types.UploadKey]types.UploadFileContent
	GetFilesToDeprecate() map[string]types.DeprecatedFileEntry
	AddFileToUpload(key types.UploadKey, content types.UploadFileContent)
	AddFileToDeprecate(file string, entry types.DeprecatedFileEntry)
	ClearFilesToUpload()
	ClearFilesToDeprecate()
}

// DefaultFileStateService implements FileStateService with thread-safe operations
type DefaultFileStateService struct {
	mu               sync.RWMutex
	filesToUpload    map[types.UploadKey]types.UploadFileContent
	filesToDeprecate map[string]types.DeprecatedFileEntry
}

// NewFileStateService creates a new file state service
func NewFileStateService() FileStateService {
	return &DefaultFileStateService{
		filesToUpload:    make(map[types.UploadKey]types.UploadFileContent),
		filesToDeprecate: make(map[string]types.DeprecatedFileEntry),
	}
}

// GetFilesToUpload returns a copy of the files to upload map
func (fss *DefaultFileStateService) GetFilesToUpload() map[types.UploadKey]types.UploadFileContent {
	fss.mu.RLock()
	defer fss.mu.RUnlock()
	
	// Return a copy to prevent external modification
	result := make(map[types.UploadKey]types.UploadFileContent, len(fss.filesToUpload))
	for k, v := range fss.filesToUpload {
		result[k] = v
	}
	return result
}

// GetFilesToDeprecate returns a copy of the files to deprecate map
func (fss *DefaultFileStateService) GetFilesToDeprecate() map[string]types.DeprecatedFileEntry {
	fss.mu.RLock()
	defer fss.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]types.DeprecatedFileEntry, len(fss.filesToDeprecate))
	for k, v := range fss.filesToDeprecate {
		result[k] = v
	}
	return result
}

// AddFileToUpload adds or updates a file to upload
func (fss *DefaultFileStateService) AddFileToUpload(key types.UploadKey, content types.UploadFileContent) {
	fss.mu.Lock()
	defer fss.mu.Unlock()

	fss.filesToUpload[key] = content
}

// AddFileToDeprecate adds a file to the deprecation list
func (fss *DefaultFileStateService) AddFileToDeprecate(file string, entry types.DeprecatedFileEntry) {
	fss.mu.Lock()
	defer fss.mu.Unlock()

	fss.filesToDeprecate[file] = entry
}

// ClearFilesToUpload clears the files to upload map
func (fss *DefaultFileStateService) ClearFilesToUpload() {
	fss.mu.Lock()
	defer fss.mu.Unlock()
	
	fss.filesToUpload = make(map[types.UploadKey]types.UploadFileContent)
}

// ClearFilesToDeprecate clears the files to deprecate map
func (fss *DefaultFileStateService) ClearFilesToDeprecate() {
	fss.mu.Lock()
	defer fss.mu.Unlock()

	fss.filesToDeprecate = make(map[string]types.DeprecatedFileEntry)
}

