package utils

import (
	"gdcd/add-code-examples"
	"path"
)

func GetLangFromFilepath(filepath string) string {
	extension := ""
	switch extension = path.Ext(filepath); extension {
	// NOTE: this switch statement omits Bash, because the Bash extension is .sh - so we treat that language as Shell
	case add_code_examples.CExtension:
		return add_code_examples.C
	case add_code_examples.CPPExtension:
		return add_code_examples.CPP
	case add_code_examples.CSharpExtension:
		return add_code_examples.CSharp
	case add_code_examples.GoExtension:
		return add_code_examples.Go
	case add_code_examples.JavaExtension:
		return add_code_examples.Java
	case add_code_examples.JavaScriptExtension:
		return add_code_examples.JavaScript
	case add_code_examples.JSONExtension:
		return add_code_examples.JSON
	case add_code_examples.KotlinExtension:
		return add_code_examples.Kotlin
	case add_code_examples.PHPExtension:
		return add_code_examples.PHP
	case add_code_examples.PythonExtension:
		return add_code_examples.Python
	case add_code_examples.RubyExtension:
		return add_code_examples.Ruby
	case add_code_examples.RustExtension:
		return add_code_examples.Rust
	case add_code_examples.ScalaExtension:
		return add_code_examples.Scala
	case add_code_examples.ShellExtension:
		return add_code_examples.Shell
	case add_code_examples.SwiftExtension:
		return add_code_examples.Swift
	case add_code_examples.TextExtension:
		return add_code_examples.Text
	case add_code_examples.TypeScriptExtension:
		return add_code_examples.TypeScript
	case add_code_examples.XMLExtension:
		return add_code_examples.XML
	case add_code_examples.YAMLExtension:
		return add_code_examples.YAML
	default:
		return add_code_examples.Undefined
	}
}
