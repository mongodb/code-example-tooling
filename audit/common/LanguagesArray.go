package common

// LanguagesArray is a custom type to handle unmarshalling languages.
type LanguagesArray []map[string]LanguageCounts

// ToMap converts the LanguagesArray to a map[string]LanguageMetrics.
func (languages LanguagesArray) ToMap() map[string]LanguageCounts {
	result := make(map[string]LanguageCounts)
	for _, languageEntry := range languages {
		for lang, metrics := range languageEntry {
			result[lang] = metrics
		}
	}
	return result
}
