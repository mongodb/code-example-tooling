package add_code_examples

import (
	"snooty-api-parser/types"
)

const (
	BASH       = "bash"
	C          = "c"
	CPP        = "cpp"
	CSHARP     = "csharp"
	Go         = "go"
	Java       = "java"
	JavaScript = "javascript"
	JSON       = "json"
	Kotlin     = "kotlin"
	PHP        = "php"
	Python     = "python"
	Ruby       = "ruby"
	Rust       = "rust"
	Scala      = "scala"
	Shell      = "shell"
	Swift      = "swift"
	Text       = "text"
	TypeScript = "typescript"
	Undefined  = "undefined"
	XML        = "xml"
	YAML       = "yaml"
)

func GetLanguage(snootyNode types.ASTNode) string {
	normalizeLanguagesMap := make(map[string]string)

	// Add the canonical languages and their values
	normalizeLanguagesMap[BASH] = BASH
	normalizeLanguagesMap[C] = C
	normalizeLanguagesMap[CPP] = CPP
	normalizeLanguagesMap[CSHARP] = CSHARP
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
	normalizeLanguagesMap["cs"] = CSHARP
	normalizeLanguagesMap["golang"] = Go
	normalizeLanguagesMap["http"] = Text
	normalizeLanguagesMap["ini"] = Text
	normalizeLanguagesMap["js"] = JavaScript
	normalizeLanguagesMap["none"] = Undefined
	normalizeLanguagesMap["sh"] = Shell
	normalizeLanguagesMap["json\\n :copyable: false"] = JSON
	normalizeLanguagesMap["json\\n :copyable: true"] = JSON

	snootyLanguageValue := snootyNode.Lang
	canonicalLanguage, exists := normalizeLanguagesMap[snootyLanguageValue]
	if exists {
		return canonicalLanguage
	} else {
		return Undefined
	}
}
