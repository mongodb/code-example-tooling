// Package pages provides output formatting for page count results.
package pages

import (
	"fmt"
	"sort"
)

// PrintResults prints the counting results.
//
// If countByProject is true, prints a breakdown by project.
// If byVersion is true, prints a breakdown by project and version.
// Otherwise, prints only the total count.
//
// Parameters:
//   - result: The counting results
//   - countByProject: If true, show breakdown by project
//   - byVersion: If true, show breakdown by project and version
func PrintResults(result *CountResult, countByProject bool, byVersion bool) {
	if byVersion {
		printByVersion(result)
	} else if countByProject {
		printByProject(result)
	} else {
		printTotal(result)
	}
}

// printTotal prints only the total count as a single integer.
func printTotal(result *CountResult) {
	fmt.Println(result.TotalCount)
}

// printByProject prints a breakdown of counts by project.
func printByProject(result *CountResult) {
	if len(result.ProjectCounts) == 0 {
		fmt.Println("No pages found")
		return
	}

	// Get sorted list of project names
	var projectNames []string
	for name := range result.ProjectCounts {
		projectNames = append(projectNames, name)
	}
	sort.Strings(projectNames)

	// Print header
	fmt.Println("Page Counts by Project:")
	fmt.Println()

	// Print each project with its count
	for _, name := range projectNames {
		count := result.ProjectCounts[name]
		fmt.Printf("  %-30s %5d\n", name, count)
	}

	// Print total
	fmt.Println()
	fmt.Printf("Total: %d\n", result.TotalCount)
}

// printByVersion prints a breakdown of counts by project and version.
func printByVersion(result *CountResult) {
	if len(result.VersionCounts) == 0 {
		fmt.Println("No pages found")
		return
	}

	// Get sorted list of project names
	var projectNames []string
	for name := range result.VersionCounts {
		projectNames = append(projectNames, name)
	}
	sort.Strings(projectNames)

	// Print each project with its versions
	for _, projectName := range projectNames {
		versionCounts := result.VersionCounts[projectName]

		fmt.Printf("Project: %s\n", projectName)

		// Get sorted list of version names
		var versionNames []string
		for version := range versionCounts {
			versionNames = append(versionNames, version)
		}
		sort.Strings(versionNames)

		// Print each version with its count
		for _, versionName := range versionNames {
			count := versionCounts[versionName]
			displayName := versionName
			if displayName == "" {
				displayName = "(no version)"
			}
			fmt.Printf("  %-28s %5d\n", displayName, count)
		}

		fmt.Println()
	}

	// Print total
	fmt.Printf("Total: %d\n", result.TotalCount)
}

