package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteSnippetReport(snippets []SnippetInfo, projectName string) {
	fmt.Println("Writing snippet report")
	snippetJsonData, marshallingErr := json.MarshalIndent(snippets, "", "  ")
	if marshallingErr != nil {
		fmt.Println("Error marshalling JSON:", marshallingErr)
		return
	}
	reportOutputDir := BaseReportOutputDir + projectName
	mkdirErr := os.MkdirAll(reportOutputDir, 0755)
	if mkdirErr != nil {
		fmt.Println("Error creating directory: ", mkdirErr)
		return
	}
	snippetDetailsFilepath := BaseReportOutputDir + projectName + "/snippets.json"
	writeReportErr := os.WriteFile(snippetDetailsFilepath, snippetJsonData, 0644)
	if writeReportErr != nil {
		fmt.Println("Error writing JSON to file:", writeReportErr)
		return
	}
	fmt.Println("Snippet report successfully written to", snippetDetailsFilepath)
}

func CalculateAccuracyPercentages(totalCodeCount int, llmCategorizedCount int, stringMatchedCount int, isDriversProject bool) float64 {
	if totalCodeCount == 0 {
		fmt.Println("Total code count is zero, cannot perform calculations.")
		return 0
	}
	// Calculate the percentage contribution of stringMatchedCount and llmCategorizedCount
	stringMatchedPercentage := (float64(stringMatchedCount) / float64(totalCodeCount)) * 100
	llmCategorizedPercentage := (float64(llmCategorizedCount) / float64(totalCodeCount)) * 100
	// Calculate accuracy estimate
	stringMatchAccuracy := float64(stringMatchedCount) // 100% accuracy
	var llmCategorizedAccuracy float64
	if isDriversProject {
		llmCategorizedAccuracy = float64(llmCategorizedCount) * 0.80 // 80% accuracy
	} else {
		llmCategorizedAccuracy = float64(llmCategorizedCount) * 0.65 // 65% accuracy
	}
	// Combined accuracy calculation
	totalAccuracyEstimate := (stringMatchAccuracy + llmCategorizedAccuracy) / float64(totalCodeCount) * 100
	// Print the results
	fmt.Printf("String Matched Percentage: %.2f%%\n", stringMatchedPercentage)
	fmt.Printf("LLM Categorized Percentage: %.2f%%\n", llmCategorizedPercentage)
	fmt.Printf("Overall Accuracy Estimate: %.2f%%\n", totalAccuracyEstimate)
	return totalAccuracyEstimate
}

func WriteCategoryCountsReport(totalCodeBlocks int, counts map[string]map[string]int, llmCategorised int, stringMatched int, projectName string, isDriversProject bool) {
	categorySums := GetCategorySums(counts)
	accuracyEstimate := CalculateAccuracyPercentages(totalCodeBlocks, llmCategorised, stringMatched, isDriversProject)
	catDetails := CategorizationDetails{
		LLMCategorizedCount: llmCategorised,
		StringMatchedCount:  stringMatched,
		AccuracyEstimate:    accuracyEstimate,
	}
	repoReport := RepoReport{
		TotalCodeBlocks:        totalCodeBlocks,
		CategorizationDetails:  catDetails,
		CategoryLanguageCounts: categorySums,
	}
	repoData, jsonMarshallingErr := json.MarshalIndent(repoReport, "", "  ")

	if jsonMarshallingErr != nil {
		fmt.Println("Error marshalling JSON:", jsonMarshallingErr)
		return
	}
	fmt.Println("Writing category and language counts report")
	filePath := BaseReportOutputDir + projectName + "/language_category_counts.json"
	writeReportErr := os.WriteFile(filePath, repoData, 0644)
	if writeReportErr != nil {
		fmt.Println("Error writing JSON to file: ", writeReportErr)
		return
	}
	fmt.Println("Category and language counts report successfully written to", filePath)
}
