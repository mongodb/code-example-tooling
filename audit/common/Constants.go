package common

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

	// Code example categories

	SyntaxExample              = "Syntax example"
	NonMongoCommand            = "Non-MongoDB command"
	ExampleReturnObject        = "Example return object"
	ExampleConfigurationObject = "Example configuration object"
	UsageExample               = "Usage example"

	/*
		The constants below for Products, SubProducts, and Directories are used in
		the `productInfoMap` in the GetProductInfo.go file. The GetProductInfo
		func relies on these constants to return the appropriate `product` and/or
		`sub_product` name for their respective fields on a per-docs-page basis
		in the code examples DB.
	*/
	// Products

	Atlas                        = "Atlas"
	AtlasArchitecture            = "Atlas Architecture Center"
	BIConnector                  = "BI Connector"
	CloudManager                 = "Cloud Manager"
	Compass                      = "Compass"
	DBTools                      = "Database Tools"
	Django                       = "Django Integration"
	Drivers                      = "Drivers"
	EnterpriseKubernetesOperator = "Enterprise Kubernetes Operator"
	EFCoreProvider               = "Entity Framework Core Provider"
	KafkaConnector               = "Kafka Connector"
	MCPServer                    = "MongoDB MCP Server"
	MDBCLI                       = "MongoDB CLI"
	Mongosh                      = "MongoDB Shell"
	Mongosync                    = "Mongosync"
	OpsManager                   = "Ops Manager"
	RelationalMigrator           = "Relational Migrator"
	Server                       = "Server"
	SparkConnector               = "Spark Connector"

	// SubProducts

	AtlasCLI         = "Atlas CLI"
	AtlasOperator    = "Kubernetes Operator"
	Charts           = "Charts"
	DataFederation   = "Data Federation"
	OnlineArchive    = "Online Archive"
	Search           = "Search"
	StreamProcessing = "Stream Processing"
	Terraform        = "Terraform"
	Triggers         = "Triggers"
	VectorSearch     = "Vector Search"

	// Directories that map to specific sub-products

	DataFederationDir   = "data-federation"
	OnlineArchiveDir    = "online-archive"
	SearchDir           = "atlas-search"
	StreamProcessingDir = "atlas-stream-processing"
	TriggersDir         = "triggers"
	VectorSearchDir     = "atlas-vector-search"
	AiIntegrationsDir   = "ai-integrations"
)

var CanonicalLanguages = []string{Bash, C, CPP,
	CSharp, Go, Java, JavaScript,
	JSON, Kotlin, PHP, Python,
	Ruby, Rust, Scala, Shell,
	Swift, Text, TypeScript, Undefined, XML, YAML,
}
