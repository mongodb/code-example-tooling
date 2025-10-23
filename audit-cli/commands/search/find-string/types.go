package find_string

// SearchResult contains the results of searching a single file.
//
// Used internally during the search operation to track results for each file.
type SearchResult struct {
	FilePath string // Path to the file that was searched
	Language string // Programming language (detected from file extension)
	Contains bool   // Whether the file contains the substring
}

// SearchReport contains statistics about the search operation.
//
// Tracks overall statistics for reporting to the user.
type SearchReport struct {
	FilesScanned       int            // Total number of files scanned
	FilesContaining    int            // Number of files containing the substring
	LanguageCounts     map[string]int // Count of files containing substring by language
	FilesWithSubstring []string       // List of file paths containing the substring
}

// NewSearchReport creates a new initialized SearchReport with empty maps and slices.
func NewSearchReport() *SearchReport {
	return &SearchReport{
		LanguageCounts:     make(map[string]int),
		FilesWithSubstring: make([]string, 0),
	}
}

// AddResult updates the report with a search result.
//
// This method should be called once for each file that is searched.
// It updates the statistics based on whether the file contains the substring.
func (r *SearchReport) AddResult(result SearchResult) {
	r.FilesScanned++

	if result.Contains {
		r.FilesContaining++
		r.FilesWithSubstring = append(r.FilesWithSubstring, result.FilePath)

		if result.Language != "" {
			r.LanguageCounts[result.Language]++
		}
	}
}
