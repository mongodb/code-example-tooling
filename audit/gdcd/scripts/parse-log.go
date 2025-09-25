package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// PageEvent represents a page creation or removal event
type PageEvent struct {
	Action       string // "removed" or "created"
	PageID       string
	Project      string
	CodeExamples int
	AppliedUsage int
}

// MovedPage represents a page that was moved from one location to another
type MovedPage struct {
	FromID       string
	ToID         string
	Project      string
	CodeExamples int
}

// AppliedUsageExample represents new applied usage examples on truly new pages
type AppliedUsageExample struct {
	PageID  string
	Project string
	Count   int
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run parse-log.go <log-file-path>")
		fmt.Println("Example: go run parse-log.go ../logs/2025-09-24-18-01-30-app.log")
		os.Exit(1)
	}

	logFile := os.Args[1]

	file, err := os.Open(logFile)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Regular expressions for parsing log lines
	projectChangesRegex := regexp.MustCompile(`Project changes for (.+)`)
	pageRemovedRegex := regexp.MustCompile(`Page removed: Page ID: (.+)`)
	pageCreatedRegex := regexp.MustCompile(`Page created: Page ID: (.+)`)
	codeExampleRemovedRegex := regexp.MustCompile(`Code example removed: Page ID: (.+), (\d+) code examples removed`)
	codeExampleCreatedRegex := regexp.MustCompile(`Code example created: Page ID: (.+), (\d+) new code examples added`)
	appliedUsageRegex := regexp.MustCompile(`Applied usage example added: Page ID: (.+), (\d+) new applied usage examples added`)

	removedPages := make(map[string]PageEvent)
	createdPages := make(map[string]PageEvent)
	appliedUsageMap := make(map[string]AppliedUsageExample)

	currentProject := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse project changes line to track current project
		if matches := projectChangesRegex.FindStringSubmatch(line); matches != nil {
			currentProject = matches[1]
			continue
		}

		// Skip processing if we don't have a current project
		if currentProject == "" {
			continue
		}

		// Parse page removed events
		if matches := pageRemovedRegex.FindStringSubmatch(line); matches != nil {
			pageID := matches[1]
			key := currentProject + "|" + pageID
			removedPages[key] = PageEvent{
				Action:  "removed",
				PageID:  pageID,
				Project: currentProject,
			}
		}

		// Parse page created events
		if matches := pageCreatedRegex.FindStringSubmatch(line); matches != nil {
			pageID := matches[1]
			key := currentProject + "|" + pageID
			createdPages[key] = PageEvent{
				Action:  "created",
				PageID:  pageID,
				Project: currentProject,
			}
		}

		// Parse code example removed events
		if matches := codeExampleRemovedRegex.FindStringSubmatch(line); matches != nil {
			pageID := matches[1]
			count, _ := strconv.Atoi(matches[2])
			key := currentProject + "|" + pageID
			if page, exists := removedPages[key]; exists {
				page.CodeExamples = count
				removedPages[key] = page
			}
		}

		// Parse code example created events
		if matches := codeExampleCreatedRegex.FindStringSubmatch(line); matches != nil {
			pageID := matches[1]
			count, _ := strconv.Atoi(matches[2])
			key := currentProject + "|" + pageID
			if page, exists := createdPages[key]; exists {
				page.CodeExamples = count
				createdPages[key] = page
			}
		}

		// Parse applied usage example events
		if matches := appliedUsageRegex.FindStringSubmatch(line); matches != nil {
			pageID := matches[1]
			count, _ := strconv.Atoi(matches[2])
			key := currentProject + "|" + pageID
			appliedUsageMap[key] = AppliedUsageExample{
				PageID:  pageID,
				Project: currentProject,
				Count:   count,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Identify moved pages
	movedPages := []MovedPage{}
	trulyRemovedPages := []PageEvent{}
	trulyCreatedPages := []PageEvent{}

	// Create a copy of removed pages to track which ones we've processed
	unprocessedRemoved := make(map[string]PageEvent)
	for k, v := range removedPages {
		unprocessedRemoved[k] = v
	}

	// Check each created page to see if it matches a removed page within the same project
	for _, createdPage := range createdPages {
		matchFound := false

		for removedKey, removedPage := range unprocessedRemoved {
			// Only consider matches within the same project
			if removedPage.Project == createdPage.Project &&
				isPageMoved(removedPage.PageID, createdPage.PageID, removedPage.CodeExamples, createdPage.CodeExamples) {
				// This is a moved page
				movedPages = append(movedPages, MovedPage{
					FromID:       removedPage.PageID,
					ToID:         createdPage.PageID,
					Project:      createdPage.Project,
					CodeExamples: createdPage.CodeExamples,
				})

				// Remove from unprocessed
				delete(unprocessedRemoved, removedKey)
				matchFound = true
				break
			}
		}

		if !matchFound {
			// This is a truly new page
			trulyCreatedPages = append(trulyCreatedPages, createdPage)
		}
	}

	// Remaining unprocessed removed pages are truly removed
	for _, removedPage := range unprocessedRemoved {
		trulyRemovedPages = append(trulyRemovedPages, removedPage)
	}

	// Print results
	printResults(movedPages, trulyCreatedPages, trulyRemovedPages, appliedUsageMap)
}

// isPageMoved checks if a removed page and created page represent the same page that was moved
func isPageMoved(removedID, createdID string, removedCodeExamples, createdCodeExamples int) bool {
	// Both conditions must be true:
	// 1. Same number of code examples
	// 2. At least one segment of the page ID is the same

	if removedCodeExamples != createdCodeExamples {
		return false
	}

	removedSegments := strings.Split(removedID, "|")
	createdSegments := strings.Split(createdID, "|")

	// Check if any segment matches
	for _, removedSegment := range removedSegments {
		for _, createdSegment := range createdSegments {
			if removedSegment == createdSegment {
				return true
			}
		}
	}

	return false
}

func printResults(movedPages []MovedPage, trulyCreatedPages []PageEvent, trulyRemovedPages []PageEvent, appliedUsageMap map[string]AppliedUsageExample) {
	fmt.Println("=== MOVED PAGES ===")
	if len(movedPages) == 0 {
		fmt.Println("No moved pages found.")
	} else {
		for _, moved := range movedPages {
			fmt.Printf("MOVED [%s]: %s -> %s (%d code examples)\n", moved.Project, moved.FromID, moved.ToID, moved.CodeExamples)
		}
	}

	/* If a page doesn't meet our criteria for being a "moved page", it's a "maybe new" or "maybe removed" page. It may
	 * be a completely renamed existing page where no segment of the ID matches i.e. `crud|update` being renamed
	 * `write|upsert` - or a moved page with some matching segment element but with a different number of code examples.
	 * If so, it would not meet our criteria for being a "moved" page. Compare these pages with the "maybe removed pages"
	 * to determine if they're truly new or removed.
	 */
	fmt.Println("\n=== MAYBE NEW PAGES ===")
	if len(trulyCreatedPages) == 0 {
		fmt.Println("No maybe new pages found.")
	} else {
		for _, created := range trulyCreatedPages {
			fmt.Printf("NEW [%s]: %s (%d total code examples)\n", created.Project, created.PageID, created.CodeExamples)
		}
	}

	fmt.Println("\n=== MAYBE REMOVED PAGES ===")
	if len(trulyRemovedPages) == 0 {
		fmt.Println("No maybe removed pages found.")
	} else {
		for _, removed := range trulyRemovedPages {
			fmt.Printf("REMOVED [%s]: %s (%d total code examples)\n", removed.Project, removed.PageID, removed.CodeExamples)
		}
	}

	// If a page is maybe new, we want to check if it has any new applied usage examples. If so, we want to report those.
	fmt.Println("\n=== NEW APPLIED USAGE EXAMPLES ===")

	// Filter applied usage examples to only include maybe new pages
	trulyNewAppliedUsage := []AppliedUsageExample{}
	totalNewAppliedUsage := 0

	// Create a set of moved page destination keys for quick lookup
	movedPageDestinations := make(map[string]bool)
	for _, moved := range movedPages {
		key := moved.Project + "|" + moved.ToID
		movedPageDestinations[key] = true
	}

	for key, usage := range appliedUsageMap {
		// Only include if this page is maybe new (not moved)
		if !movedPageDestinations[key] {
			trulyNewAppliedUsage = append(trulyNewAppliedUsage, usage)
			totalNewAppliedUsage += usage.Count
		}
	}

	if len(trulyNewAppliedUsage) == 0 {
		fmt.Println("No new applied usage examples on maybe new pages found.")
	} else {
		for _, usage := range trulyNewAppliedUsage {
			fmt.Printf("APPLIED USAGE [%s]: %s (%d applied usage examples)\n", usage.Project, usage.PageID, usage.Count)
		}
		fmt.Printf("\nTotal new applied usage examples: %d\n", totalNewAppliedUsage)
	}
}
