package add_code_examples

import "common"

func GetFileExtensionFromStringLang(language string) string {
	langExtensionMap := make(map[string]string)

	// Add the canonical languages and their extensions
	langExtensionMap[common.Bash] = common.BashExtension
	langExtensionMap[common.C] = common.CExtension
	langExtensionMap[common.CPP] = common.CPPExtension
	langExtensionMap[common.CSharp] = common.CSharpExtension
	langExtensionMap[common.Go] = common.GoExtension
	langExtensionMap[common.Java] = common.JavaExtension
	langExtensionMap[common.JavaScript] = common.JavaScriptExtension
	langExtensionMap[common.JSON] = common.JSONExtension
	langExtensionMap[common.Kotlin] = common.KotlinExtension
	langExtensionMap[common.PHP] = common.PHPExtension
	langExtensionMap[common.Python] = common.PythonExtension
	langExtensionMap[common.Ruby] = common.RubyExtension
	langExtensionMap[common.Rust] = common.RustExtension
	langExtensionMap[common.Scala] = common.ScalaExtension
	langExtensionMap[common.Shell] = common.ShellExtension
	langExtensionMap[common.Swift] = common.SwiftExtension
	langExtensionMap[common.Text] = common.TextExtension
	langExtensionMap[common.TypeScript] = common.TypeScriptExtension
	langExtensionMap[common.Undefined] = common.UndefinedExtension
	langExtensionMap[common.XML] = common.XMLExtension
	langExtensionMap[common.YAML] = common.YAMLExtension

	// Add variations and map to canonical values
	langExtensionMap[""] = common.UndefinedExtension
	langExtensionMap["console"] = common.ShellExtension
	langExtensionMap["cs"] = common.CSharpExtension
	langExtensionMap["golang"] = common.GoExtension
	langExtensionMap["http"] = common.TextExtension
	langExtensionMap["ini"] = common.TextExtension
	langExtensionMap["js"] = common.JavaScriptExtension
	langExtensionMap["none"] = common.UndefinedExtension
	langExtensionMap["sh"] = common.ShellExtension
	langExtensionMap["json\\n :copyable: false"] = common.JSONExtension
	langExtensionMap["json\\n :copyable: true"] = common.JSONExtension

	extension, exists := langExtensionMap[language]
	if exists {
		return extension
	} else {
		return common.UndefinedExtension
	}
}
