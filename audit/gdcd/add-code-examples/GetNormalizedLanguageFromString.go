package add_code_examples

func GetNormalizedLanguageFromString(language string) string {
	normalizeLanguagesMap := make(map[string]string)

	// Add the canonical languages and their values
	normalizeLanguagesMap[Bash] = Bash
	normalizeLanguagesMap[C] = C
	normalizeLanguagesMap[CPP] = CPP
	normalizeLanguagesMap[CSharp] = CSharp
	normalizeLanguagesMap[Go] = Go
	normalizeLanguagesMap[Java] = Java
	normalizeLanguagesMap[JavaScript] = JavaScript
	normalizeLanguagesMap[JSON] = JSON
	normalizeLanguagesMap[Kotlin] = Kotlin
	normalizeLanguagesMap[PHP] = PHP
	normalizeLanguagesMap[Python] = Python
	normalizeLanguagesMap[Ruby] = Ruby
	normalizeLanguagesMap[Rust] = Rust
	normalizeLanguagesMap[Scala] = Scala
	normalizeLanguagesMap[Shell] = Shell
	normalizeLanguagesMap[Swift] = Swift
	normalizeLanguagesMap[Text] = Text
	normalizeLanguagesMap[TypeScript] = TypeScript
	normalizeLanguagesMap[Undefined] = Undefined
	normalizeLanguagesMap[XML] = XML
	normalizeLanguagesMap[YAML] = YAML

	// Add variations and map to canonical values
	normalizeLanguagesMap[""] = Undefined
	normalizeLanguagesMap["console"] = Shell
	normalizeLanguagesMap["cs"] = CSharp
	normalizeLanguagesMap["golang"] = Go
	normalizeLanguagesMap["http"] = Text
	normalizeLanguagesMap["ini"] = Text
	normalizeLanguagesMap["js"] = JavaScript
	normalizeLanguagesMap["none"] = Undefined
	normalizeLanguagesMap["sh"] = Shell
	normalizeLanguagesMap["json\\n :copyable: false"] = JSON
	normalizeLanguagesMap["json\\n :copyable: true"] = JSON

	canonicalLanguage, exists := normalizeLanguagesMap[language]
	if exists {
		return canonicalLanguage
	} else {
		return Undefined
	}
}
