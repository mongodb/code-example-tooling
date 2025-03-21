package compare_code_examples

import "github.com/sergi/go-diff/diffmatchpatch"

func DiffCodeExamples(original, newString string, percentChangeAccepted float64) bool {
	isTheSameString := false
	originalCount := len(original)
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, newString, false)
	totalChanges := 0
	deletedCharacterCount := 0
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffInsert {
			totalChanges += len(diff.Text)
		} else if diff.Type == diffmatchpatch.DiffDelete {
			deletedCharacterCount += len(diff.Text)
			if deletedCharacterCount == originalCount {
				return false
			} else {
				totalChanges -= len(diff.Text)
			}
		}
	}
	if totalChanges < 0 {
		totalChanges = totalChanges * -1
	}
	var changePercentage float64

	if totalChanges == 0 {
		changePercentage = (float64(deletedCharacterCount) / float64(originalCount)) * 100.0
	} else {
		changePercentage = (float64(totalChanges) / float64(originalCount)) * 100.0
	}
	if changePercentage < percentChangeAccepted {
		isTheSameString = true
	}
	return isTheSameString
}
