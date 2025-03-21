package utils

import (
	"common"
	"path"
)

func GetLangFromFilepath(filepath string) string {
	extension := ""
	switch extension = path.Ext(filepath); extension {
	// NOTE: this switch statement omits Bash, because the Bash extension is .sh - so we treat that language as Shell
	case common.CExtension:
		return common.C
	case common.CPPExtension:
		return common.CPP
	case common.CSharpExtension:
		return common.CSharp
	case common.GoExtension:
		return common.Go
	case common.JavaExtension:
		return common.Java
	case common.JavaScriptExtension:
		return common.JavaScript
	case common.JSONExtension:
		return common.JSON
	case common.KotlinExtension:
		return common.Kotlin
	case common.PHPExtension:
		return common.PHP
	case common.PythonExtension:
		return common.Python
	case common.RubyExtension:
		return common.Ruby
	case common.RustExtension:
		return common.Rust
	case common.ScalaExtension:
		return common.Scala
	case common.ShellExtension:
		return common.Shell
	case common.SwiftExtension:
		return common.Swift
	case common.TextExtension:
		return common.Text
	case common.TypeScriptExtension:
		return common.TypeScript
	case common.XMLExtension:
		return common.XML
	case common.YAMLExtension:
		return common.YAML
	default:
		return common.Undefined
	}
}
