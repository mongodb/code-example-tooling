package code_examples

import "strings"

// Language constants define canonical language names used throughout the tool.
// These are used for normalization and file extension mapping.
const (
	Bash       = "bash"
	C          = "c"
	CPP        = "cpp"
	CSharp     = "csharp"
	Console    = "console"
	Go         = "go"
	Java       = "java"
	JavaScript = "javascript"
	Kotlin     = "kotlin"
	PHP        = "php"
	PowerShell = "powershell"
	PS5        = "ps5"
	Python     = "python"
	Ruby       = "ruby"
	Rust       = "rust"
	Scala      = "scala"
	Shell      = "shell"
	Swift      = "swift"
	Text       = "text"
	TypeScript = "typescript"
	Undefined  = "undefined"
)

// File extension constants define the file extensions for each language.
// Used when generating output filenames for extracted code examples.
const (
	BashExtension       = ".sh"
	CExtension          = ".c"
	CPPExtension        = ".cpp"
	CSharpExtension     = ".cs"
	ConsoleExtension    = ".sh"
	GoExtension         = ".go"
	JavaExtension       = ".java"
	JavaScriptExtension = ".js"
	KotlinExtension     = ".kt"
	PHPExtension        = ".php"
	PowerShellExtension = ".ps1"
	PS5Extension        = ".ps1"
	PythonExtension     = ".py"
	RubyExtension       = ".rb"
	RustExtension       = ".rs"
	ScalaExtension      = ".scala"
	ShellExtension      = ".sh"
	SwiftExtension      = ".swift"
	TextExtension       = ".txt"
	TypeScriptExtension = ".ts"
	UndefinedExtension  = ".txt"
)

// GetFileExtensionFromLanguage returns the appropriate file extension for a given language.
//
// This function maps language identifiers to their corresponding file extensions.
// Handles various language name variants (e.g., "ts" -> ".ts", "c++" -> ".cpp", "golang" -> ".go").
// Returns ".txt" for unknown or undefined languages.
//
// Parameters:
//   - language: The language identifier (case-insensitive)
//
// Returns:
//   - string: The file extension including the leading dot (e.g., ".js", ".py")
func GetFileExtensionFromLanguage(language string) string {
	lang := strings.ToLower(strings.TrimSpace(language))

	langExtensionMap := map[string]string{
		Bash:       BashExtension,
		C:          CExtension,
		CPP:        CPPExtension,
		CSharp:     CSharpExtension,
		Console:    ConsoleExtension,
		Go:         GoExtension,
		Java:       JavaExtension,
		JavaScript: JavaScriptExtension,
		Kotlin:     KotlinExtension,
		PHP:        PHPExtension,
		PowerShell: PowerShellExtension,
		PS5:        PS5Extension,
		Python:     PythonExtension,
		Ruby:       RubyExtension,
		Rust:       RustExtension,
		Scala:      ScalaExtension,
		Shell:      ShellExtension,
		Swift:      SwiftExtension,
		Text:       TextExtension,
		TypeScript: TypeScriptExtension,
		Undefined:  UndefinedExtension,
		"c++":      CPPExtension,
		"c#":       CSharpExtension,
		"cs":       CSharpExtension,
		"golang":   GoExtension,
		"js":       JavaScriptExtension,
		"kt":       KotlinExtension,
		"py":       PythonExtension,
		"rb":       RubyExtension,
		"rs":       RustExtension,
		"sh":       ShellExtension,
		"ts":       TypeScriptExtension,
		"txt":      TextExtension,
		"ps1":      PowerShellExtension,
		"":         UndefinedExtension,
		"none":     UndefinedExtension,
	}

	if extension, exists := langExtensionMap[lang]; exists {
		return extension
	}

	return UndefinedExtension
}

// NormalizeLanguage normalizes a language string to a canonical form.
//
// This function converts various language name variants to their canonical forms:
//   - "ts" -> "typescript"
//   - "c++" -> "cpp"
//   - "golang" -> "go"
//   - "js" -> "javascript"
//   - etc.
//
// Parameters:
//   - language: The language identifier (case-insensitive)
//
// Returns:
//   - string: The normalized language name, or the original string if no normalization is defined
func NormalizeLanguage(language string) string {
	lang := strings.ToLower(strings.TrimSpace(language))

	normalizeMap := map[string]string{
		Bash:       Bash,
		C:          C,
		CPP:        CPP,
		CSharp:     CSharp,
		Console:    Console,
		Go:         Go,
		Java:       Java,
		JavaScript: JavaScript,
		Kotlin:     Kotlin,
		PHP:        PHP,
		PowerShell: PowerShell,
		PS5:        PS5,
		Python:     Python,
		Ruby:       Ruby,
		Rust:       Rust,
		Scala:      Scala,
		Shell:      Shell,
		Swift:      Swift,
		Text:       Text,
		TypeScript: TypeScript,
		"c++":      CPP,
		"c#":       CSharp,
		"cs":       CSharp,
		"golang":   Go,
		"js":       JavaScript,
		"kt":       Kotlin,
		"py":       Python,
		"rb":       Ruby,
		"rs":       Rust,
		"sh":       Shell,
		"ts":       TypeScript,
		"txt":      Text,
		"ps1":      PowerShell,
		"":         Undefined,
		"none":     Undefined,
	}

	if normalized, exists := normalizeMap[lang]; exists {
		return normalized
	}

	return lang
}
