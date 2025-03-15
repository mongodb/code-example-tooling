package utils

import (
	"log"
	"regexp"
	"strings"
)

func ExampleContainsString(contents string) (string, bool) {
	// These strings are typically included in usage examples
	aggregationExample := ".aggregate"
	mongoConnectionStringPrefix := "mongodb://"
	alternoConnectionStringPrefix := "mongodb+srv://"

	// These strings are typically included in return objects
	warningString := "warning"
	deprecatedString := "deprecated"
	idString := "_id"

	// These strings are typically included in non-MongoDB commands
	cmake := "cmake "

	// Some of the examples can be quite long. For the current case, we only care if `.aggregate` appears near the beginning of the example
	substringLengthToCheck := 50
	usageExampleSubstringsToEvaluate := []string{aggregationExample, mongoConnectionStringPrefix, alternoConnectionStringPrefix}
	returnObjectStringsToEvaluate := []string{warningString, deprecatedString, idString}
	nonMongoDBStringsToEvaluate := []string{cmake}

	if substringLengthToCheck < len(contents) {
		substring := contents[:substringLengthToCheck]
		for _, exampleString := range usageExampleSubstringsToEvaluate {
			if strings.Contains(substring, exampleString) {
				return UsageExample, true
			}
		}
		for _, exampleString := range returnObjectStringsToEvaluate {
			if strings.Contains(substring, exampleString) {
				return ExampleReturnObject, true
			}
		}
		for _, exampleString := range nonMongoDBStringsToEvaluate {
			if strings.Contains(substring, exampleString) {
				return NonMongoCommand, true
			}
		}
	} else {
		for _, exampleString := range usageExampleSubstringsToEvaluate {
			if strings.Contains(contents, exampleString) {
				return UsageExample, true
			}
		}
		for _, exampleString := range returnObjectStringsToEvaluate {
			if strings.Contains(contents, exampleString) {
				return ExampleReturnObject, true
			}
		}
		for _, exampleString := range nonMongoDBStringsToEvaluate {
			if strings.Contains(contents, exampleString) {
				return NonMongoCommand, true
			}
		}
	}

	/* 	This Regexp checks for '$' followed by 2 or more characters, followed by a colon
	i.e. '$gte:' or '$project:'
	AND the capture group (the part in parentheses) checks for a pair of angle brackets, which may span
	multiple lines. If the regexp matches, it's an aggregation example. If it contains one or more capture
	groups, it's an aggregation example containing something like '<placeholder>'. According to our definitions,
	we would consider an agg example with placeholders a "syntax example" - not a "usage example" - so
	the number of matches determines whether the example contains placeholders and is a syntax example. A single match
	means it does not have any capture groups and therefore does not contain placeholders, so it's a usage example.
	More than one match means it has one or more capture groups in addition to the single match, and that makes it
	a syntax example.
	*/
	aggPipeline := `(?s)\$[a-zA-Z]{2,}: ?(.*?<.+?>)?`
	re, err := regexp.Compile(aggPipeline)
	if err != nil {
		log.Fatal("Error compiling the regexp for the agg pipeline: ", err)
	}
	regExpMatches := re.FindStringSubmatch(contents)
	matchLength := len(regExpMatches)
	if matchLength > 1 {
		if regExpMatches[1] != "" {
			return SyntaxExample, true
		} else {
			return UsageExample, true
		}
	} else {
		return "Uncategorized", false
	}
}
