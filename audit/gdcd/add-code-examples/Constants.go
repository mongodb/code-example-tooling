package add_code_examples

const (
	// Programming languages

	Bash       = "bash"
	C          = "c"
	CPP        = "cpp"
	CSharp     = "csharp"
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

	// File extensions

	BashExtension       = ".sh"
	CExtension          = ".c"
	CPPExtension        = ".cpp"
	CSharpExtension     = ".cs"
	GoExtension         = ".go"
	JavaExtension       = ".java"
	JavaScriptExtension = ".js"
	JSONExtension       = ".json"
	KotlinExtension     = ".kt"
	PHPExtension        = ".php"
	PythonExtension     = ".py"
	RubyExtension       = ".rb"
	RustExtension       = ".rs"
	ScalaExtension      = ".scala"
	ShellExtension      = ".sh"
	SwiftExtension      = ".swift"
	TextExtension       = ".txt"
	TypeScriptExtension = ".ts"
	UndefinedExtension  = ".txt"
	XMLExtension        = ".xml"
	YAMLExtension       = ".yaml"

	// Programming language categories

	JsonLike       = "json-like"
	DriversMinusJs = "drivers-minus-js"

	// Code example categories

	SyntaxExample              = "Syntax example"
	NonMongoCommand            = "Non-MongoDB command"
	ExampleReturnObject        = "Example return object"
	ExampleConfigurationObject = "Example configuration object"
	UsageExample               = "Usage example"

	// Other

	MODEL = "qwen2.5-coder"
)

var CanonicalLanguages = []string{Bash, C, CPP,
	CSharp, Go, Java, JavaScript,
	JSON, Kotlin, PHP, Python,
	Ruby, Rust, Scala, Shell,
	Swift, Text, TypeScript, Undefined, XML, YAML,
}
