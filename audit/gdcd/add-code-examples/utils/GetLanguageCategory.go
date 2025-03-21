package utils

import (
	"common"
)

// This is a redeclaration of the constants in add-code-examples.Constants. Redeclaring them here to avoid import cycle error.
// There's probably a better place to put these but brain can only so much brain right now so these are dups currently.
const (
	JsonLike       = "json-like"
	DriversMinusJs = "drivers-minus-js"
)

func GetLanguageCategory(lang string) string {
	jsonLike := []string{common.JSON, common.XML, common.YAML}
	driversLanguagesMinusJS := []string{common.C, common.CPP, common.CSharp, common.Go, common.Java, common.Kotlin, common.PHP, common.Python, common.Ruby, common.Rust, common.Scala, common.Swift, common.TypeScript}
	if SliceContainsString([]string{common.Bash, common.Shell}, lang) {
		return common.Shell
	} else if SliceContainsString(jsonLike, lang) {
		return JsonLike
	} else if SliceContainsString(driversLanguagesMinusJS, lang) {
		return DriversMinusJs
	} else if lang == common.JavaScript {
		return common.JavaScript
	} else if lang == common.Text {
		return common.Text
	} else if lang == common.Undefined {
		return common.Undefined
	} else {
		return ""
	}
}
