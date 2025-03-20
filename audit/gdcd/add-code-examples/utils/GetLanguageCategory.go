package utils

const (
	Bash             = "bash"
	C                = "c"
	CPP              = "cpp"
	CSharp           = "csharp"
	Go               = "go"
	Java             = "java"
	JavaScript       = "javascript"
	JSON             = "json"
	Kotlin           = "kotlin"
	PHP              = "php"
	Python           = "python"
	Ruby             = "ruby"
	Rust             = "rust"
	Scala            = "scala"
	Shell            = "shell"
	Swift            = "swift"
	Text             = "text"
	TypeScript       = "typescript"
	Undefined        = "undefined"
	XML              = "xml"
	YAML             = "yaml"
	JSON_LIKE        = "json-like"
	DRIVERS_MINUS_JS = "drivers-minus-js"
)

func GetLanguageCategory(lang string) string {
	jsonLike := []string{JSON, XML, YAML}
	driversLanguagesMinusJS := []string{C, CPP, CSharp, Go, Java, Kotlin, PHP, Python, Ruby, Rust, Scala, Swift, TypeScript}
	if SliceContainsString([]string{Bash, Shell}, lang) {
		return Shell
	} else if SliceContainsString(jsonLike, lang) {
		return JSON_LIKE
	} else if SliceContainsString(driversLanguagesMinusJS, lang) {
		return DRIVERS_MINUS_JS
	} else if lang == JavaScript {
		return JavaScript
	} else if lang == Text {
		return Text
	} else if lang == Undefined {
		return Undefined
	} else {
		return ""
	}
}
