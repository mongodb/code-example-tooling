package main

const (
	C                = "c"
	CPP              = "cpp"
	CSHARP           = "csharp"
	GO               = "go"
	JAVA             = "java"
	JAVASCRIPT       = "javascript"
	JSON             = "json"
	KOTLIN           = "kotlin"
	PHP              = "php"
	PYTHON           = "python"
	RUBY             = "ruby"
	RUST             = "rust"
	SCALA            = "scala"
	SHELL            = "shell"
	SWIFT            = "swift"
	TEXT             = "text"
	TYPESCRIPT       = "typescript"
	XML              = "xml"
	YAML             = "yaml"
	DRIVERS_MINUS_JS = "drivers_minus_js"
	JSON_LIKE        = "json_like"
)

func GetLangFromExtension(ext string) string {
	extensionLangMap := map[string]string{
		".c":     C,
		".cpp":   CPP,
		".cs":    CSHARP,
		".go":    GO,
		".java":  JAVA,
		".js":    JAVASCRIPT,
		".json":  JSON,
		".kt":    KOTLIN,
		".php":   PHP,
		".py":    PYTHON,
		".rb":    RUBY,
		".rs":    RUST,
		".scala": SCALA,
		".sh":    SHELL,
		".swift": SWIFT,
		".txt":   TEXT,
		".ts":    TYPESCRIPT,
		".xml":   XML,
		".yaml":  YAML,
	}

	return extensionLangMap[ext]
}
