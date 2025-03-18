package utils

import (
	"fmt"
)

var primaryProgress int
var primaryTarget int
var projectName string
var secondaryProgress int
var secondaryTarget int

const (
	barWidth                    = 50
	lineMaxWidth                = 110
	incompleteProgressCharacter = "･"
	completedProgressCharacter  = "￭"
)

func padOutput(s string) string {
	return fmt.Sprintf("%-*s", lineMaxWidth, s)
}

func moveCursorUp(lines int) {
	fmt.Printf("\033[%dF", lines) // ANSI escape code to move the cursor up 'lines' lines
}

func moveCursorDown(lines int) {
	fmt.Printf("\033[%dB", lines) // ANSI escape code to move the cursor down 'lines' lines
}

func SetUpProgressDisplay(totalProjects int, docsPages int, name string) {
	primaryProgress = 0
	primaryTarget = totalProjects
	secondaryProgress = 0
	secondaryTarget = docsPages
	projectName = name
	PrintPrimaryProgressIndicator()
	PrintSecondaryProgressIndicator()
}

func UpdateSecondaryTarget() {
	if secondaryProgress < secondaryTarget {
		secondaryProgress++
		moveCursorUp(1)
		PrintSecondaryProgressIndicator()
	}
}

func SetNewSecondaryTarget(docsPages int, name string) {
	secondaryProgress = 0
	secondaryTarget = docsPages
	projectName = name
	PrintSecondaryProgressIndicator()
}

func UpdatePrimaryTarget() {
	if primaryProgress < primaryTarget {
		primaryProgress++
		moveCursorUp(2)
		PrintPrimaryProgressIndicator()
	}
}

func PrintPrimaryProgressIndicator() {
	primaryPercent := float64(primaryProgress) / float64(primaryTarget) * 100
	primaryNumHashes := int(float64(primaryProgress) / float64(primaryTarget) * float64(barWidth))
	primaryBar := fmt.Sprintf("[%s%s]", repeat(completedProgressCharacter, primaryNumHashes), repeat(incompleteProgressCharacter, barWidth-primaryNumHashes))
	indicator := fmt.Sprintf("Projects progress: %s %.2f%%", primaryBar, primaryPercent)
	fmt.Println(padOutput(indicator))
}

func PrintSecondaryProgressIndicator() {
	secondaryPercent := float64(secondaryProgress) / float64(secondaryTarget) * 100
	secondaryNumHashes := int(float64(secondaryProgress) / float64(secondaryTarget) * float64(barWidth))
	secondaryBar := fmt.Sprintf("[%s%s]", repeat(completedProgressCharacter, secondaryNumHashes), repeat(incompleteProgressCharacter, barWidth-secondaryNumHashes))
	indicator := fmt.Sprintf("Pages in %s progress: %s %.2f%%", projectName, secondaryBar, secondaryPercent)
	fmt.Println(padOutput(indicator))
}

func FinishPrintingProgressIndicators() {
	moveCursorDown(2)
}

func repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return s + repeat(s, count-1)
}
