package add_code_examples

import "common"

func GetNormalizedLanguageFromString(language string) string {
	normalizeLanguagesMap := make(map[string]string)

	// Add the canonical languages and their values
	normalizeLanguagesMap[common.Bash] = common.Bash
	normalizeLanguagesMap[common.C] = common.C
	normalizeLanguagesMap[common.CPP] = common.CPP
	normalizeLanguagesMap[common.CSharp] = common.CSharp
	normalizeLanguagesMap[common.Go] = common.Go
	normalizeLanguagesMap[common.Java] = common.Java
	normalizeLanguagesMap[common.JavaScript] = common.JavaScript
	normalizeLanguagesMap[common.JSON] = common.JSON
	normalizeLanguagesMap[common.Kotlin] = common.Kotlin
	normalizeLanguagesMap[common.PHP] = common.PHP
	normalizeLanguagesMap[common.Python] = common.Python
	normalizeLanguagesMap[common.Ruby] = common.Ruby
	normalizeLanguagesMap[common.Rust] = common.Rust
	normalizeLanguagesMap[common.Scala] = common.Scala
	normalizeLanguagesMap[common.Shell] = common.Shell
	normalizeLanguagesMap[common.Swift] = common.Swift
	normalizeLanguagesMap[common.Text] = common.Text
	normalizeLanguagesMap[common.TypeScript] = common.TypeScript
	normalizeLanguagesMap[common.Undefined] = common.Undefined
	normalizeLanguagesMap[common.XML] = common.XML
	normalizeLanguagesMap[common.YAML] = common.YAML

	// Add variations and map to canonical values
	normalizeLanguagesMap[""] = common.Undefined
	normalizeLanguagesMap["console"] = common.Shell
	normalizeLanguagesMap["cs"] = common.CSharp
	normalizeLanguagesMap["golang"] = common.Go
	normalizeLanguagesMap["http"] = common.Text
	normalizeLanguagesMap["ini"] = common.Text
	normalizeLanguagesMap["js"] = common.JavaScript
	normalizeLanguagesMap["none"] = common.Undefined
	normalizeLanguagesMap["sh"] = common.Shell
	normalizeLanguagesMap["json\\n :copyable: false"] = common.JSON
	normalizeLanguagesMap["json\\n :copyable: true"] = common.JSON

	canonicalLanguage, exists := normalizeLanguagesMap[language]
	if exists {
		return canonicalLanguage
	} else {
		return common.Undefined
	}
}
