package utils

import (
	"fmt"
)

var primaryProgress int
var primaryTarget int
var projectName string
var secondaryProgress int
var secondaryTarget int
var currentCursorLine int

const (
	barWidth                    = 50
	incompleteProgressCharacter = "･"
	completedProgressCharacter  = "￭"
	primaryTargetLine           = 1
	secondaryTargetLine         = 2
	finishPositionLine          = 3
)

func moveCursorUp(lines int) {
	fmt.Printf("\033[%dF", lines) // ANSI escape code to move the cursor up 'lines' lines
}

func moveCursorDown(lines int) {
	fmt.Printf("\033[%dE", lines) // ANSI escape code to move the cursor down 'lines' lines
}

func SetUpProgressDisplay(totalProjects int, docsPages int, name string) {
	currentCursorLine = 1
	primaryProgress = 0
	primaryTarget = totalProjects
	secondaryProgress = 0
	secondaryTarget = docsPages
	projectName = name
	SetUpPrimaryProgressIndicator()
	SetUpSecondaryProgressIndicator()
}

func recursivelyMoveToCorrectLineForTarget(target int) {
	if currentCursorLine > target {
		moveCursorUp(1)
		currentCursorLine--
		recursivelyMoveToCorrectLineForTarget(target)
	} else if currentCursorLine < target {
		moveCursorDown(1)
		currentCursorLine++
		recursivelyMoveToCorrectLineForTarget(target)
	} else {
		// Do nothing because we're at the correct target line
	}
}

func UpdateSecondaryTarget() {
	if secondaryProgress < secondaryTarget {
		secondaryProgress++
		SetUpSecondaryProgressIndicator()
	}
}

func SetNewSecondaryTarget(docsPages int, name string) {
	secondaryProgress = 0
	secondaryTarget = docsPages
	projectName = name
	SetUpSecondaryProgressIndicator()
}

func UpdatePrimaryTarget() {
	if primaryProgress < primaryTarget {
		primaryProgress++
		SetUpPrimaryProgressIndicator()
	}
}

func SetUpPrimaryProgressIndicator() {
	primaryPercent := float64(primaryProgress) / float64(primaryTarget) * 100
	primaryNumHashes := int(float64(primaryProgress) / float64(primaryTarget) * float64(barWidth))
	primaryBar := fmt.Sprintf("[%s%s]", repeat(completedProgressCharacter, primaryNumHashes), repeat(incompleteProgressCharacter, barWidth-primaryNumHashes))
	message := "Projects progress: %s%s %.2f"
	PrintIndicator(message, "", primaryBar, primaryPercent, primaryTargetLine)
}

func SetUpSecondaryProgressIndicator() {
	secondaryPercent := float64(secondaryProgress) / float64(secondaryTarget) * 100
	secondaryNumHashes := int(float64(secondaryProgress) / float64(secondaryTarget) * float64(barWidth))
	secondaryBar := fmt.Sprintf("[%s%s]", repeat(completedProgressCharacter, secondaryNumHashes), repeat(incompleteProgressCharacter, barWidth-secondaryNumHashes))
	message := "Pages in %s progress: %s %.2f"
	PrintIndicator(message, projectName, secondaryBar, secondaryPercent, secondaryTargetLine)
}

func PrintIndicator(message string, maybeProjectName string, indicatorBar string, progressPercent float64, targetLine int) {
	indicator := fmt.Sprintf(message, maybeProjectName, indicatorBar, progressPercent)
	recursivelyMoveToCorrectLineForTarget(targetLine)
	fmt.Printf("\033[2K\033[0G")
	fmt.Printf(indicator)
}

func FinishPrintingProgressIndicators() {
	recursivelyMoveToCorrectLineForTarget(finishPositionLine)
}

func repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return s + repeat(s, count-1)
}
