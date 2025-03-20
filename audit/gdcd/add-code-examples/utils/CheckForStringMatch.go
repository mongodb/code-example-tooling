package utils

// CheckForStringMatch The bool we return from this func represents whether the string matching was successful.
// If the string match was successful, we don't need to move on to LLM matching.
func CheckForStringMatch(contents string, langCategory string) (string, bool) {
	// Prefix matching should be fastest as it only has to search the first N characters of a string to determine whether it's
	// a match. So first, try to match prefixes.
	category, hasPrefix := HasStringMatchPrefix(contents, langCategory)
	if hasPrefix {
		return category, hasPrefix
	} else {
		// If the prefix matching doesn't work, try the slower string matching.
		thisCategory, containsExampleString := ExampleContainsString(contents)
		if containsExampleString {
			return thisCategory, containsExampleString
		} else {
			return "Uncategorized", false
		}
	}
}
