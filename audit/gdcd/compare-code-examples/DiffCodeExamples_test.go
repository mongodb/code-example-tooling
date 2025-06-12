package compare_code_examples

import (
	"log"
	"testing"
)

func TestDiffCodeExampleNoDifference(t *testing.T) {
	atlasExample := "db.collection.find(\"foo\")"
	snootyExample := "db.collection.find(\"foo\")"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if !isSameExample {
		log.Fatalf("Test should show the example as the same but it didn't.")
	}
}

func TestDiffCodeExampleOneCharacterDifference(t *testing.T) {
	atlasExample := "db.collection.find(\"foo\")"
	snootyExample := "db.collection.find(\"food\")"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if !isSameExample {
		log.Fatalf("Test should show the example as the same but it didn't.")
	}
}

func TestDiffCodeExample10PercentDifference(t *testing.T) {
	atlasExample := "1234567890"
	snootyExample := "1234567891"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if !isSameExample {
		log.Fatalf("Test should show the example as the same but it didn't.")
	}
}

func TestDiffCodeExample20PercentDifference(t *testing.T) {
	atlasExample := "1234567890"
	snootyExample := "1234567812"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if !isSameExample {
		log.Fatalf("Test should show the example as the same but it didn't.")
	}
}

func TestDiffCodeExample30PercentDifference(t *testing.T) {
	atlasExample := "1234567890"
	snootyExample := "1234567hsl"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if isSameExample {
		log.Fatalf("Test checks for less than 30 percent change, so 30 percent change should fail.")
	}
}

func TestDiffCodeExample40PercentDifference(t *testing.T) {
	atlasExample := "1234567890"
	snootyExample := "1234561234"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if isSameExample {
		log.Fatalf("Test checks for less than 30 percent change, so 40 percent change should fail.")
	}
}

func TestDiffCodeExample50PercentDifference(t *testing.T) {
	atlasExample := "1234567890"
	snootyExample := "1234512345"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if isSameExample {
		log.Fatalf("Test checks for less than 30 percent change, so 50 percent change should fail.")
	}
}

func TestDiffCodeExample100PercentDifference(t *testing.T) {
	atlasExample := "1234567890"
	snootyExample := "hello world"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if isSameExample {
		log.Fatalf("Test should not show the example as the same but it does.")
	}
}

func TestDiffCodeExampleEmptyStringSnooty(t *testing.T) {
	atlasExample := "1234567890"
	snootyExample := ""
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if isSameExample {
		log.Fatalf("Test should not show the example as the same but it does.")
	}
}

func TestDiffCodeExampleEmptyStringAtlas(t *testing.T) {
	atlasExample := ""
	snootyExample := "1234567890"
	isSameExample := DiffCodeExamples(atlasExample, snootyExample, percentChangeAccepted)
	if isSameExample {
		log.Fatalf("Test should not show the example as the same but it does.")
	}
}
