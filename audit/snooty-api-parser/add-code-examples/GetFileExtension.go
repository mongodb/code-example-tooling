package add_code_examples

import "snooty-api-parser/types"

func GetFileExtension(snootyNode types.ASTNode) string {
	langExtensionMap := make(map[string]string)

	// Add the canonical languages and their extensions
	langExtensionMap[Bash] = BashExtension
	langExtensionMap[C] = CExtension
	langExtensionMap[CPP] = CPPExtension
	langExtensionMap[CSharp] = CSharpExtension
	langExtensionMap[Go] = GoExtension
	langExtensionMap[Java] = JavaExtension
	langExtensionMap[JavaScript] = JavaScriptExtension
	langExtensionMap[JSON] = JSONExtension
	langExtensionMap[Kotlin] = KotlinExtension
	langExtensionMap[PHP] = PHPExtension
	langExtensionMap[Python] = PythonExtension
	langExtensionMap[Ruby] = RubyExtension
	langExtensionMap[Rust] = RustExtension
	langExtensionMap[Scala] = ScalaExtension
	langExtensionMap[Shell] = ShellExtension
	langExtensionMap[Swift] = SwiftExtension
	langExtensionMap[Text] = TextExtension
	langExtensionMap[TypeScript] = TypeScriptExtension
	langExtensionMap[Undefined] = UndefinedExtension
	langExtensionMap[XML] = XMLExtension
	langExtensionMap[YAML] = YAMLExtension

	// Add variations and map to canonical values
	langExtensionMap[""] = UndefinedExtension
	langExtensionMap["console"] = ShellExtension
	langExtensionMap["cs"] = CSharpExtension
	langExtensionMap["golang"] = GoExtension
	langExtensionMap["http"] = TextExtension
	langExtensionMap["ini"] = TextExtension
	langExtensionMap["js"] = JavaScriptExtension
	langExtensionMap["none"] = UndefinedExtension
	langExtensionMap["sh"] = ShellExtension
	langExtensionMap["json\\n :copyable: false"] = JSONExtension
	langExtensionMap["json\\n :copyable: true"] = JSONExtension

	snootyLanguageValue := snootyNode.Lang
	extension, exists := langExtensionMap[snootyLanguageValue]
	if exists {
		return extension
	} else {
		return UndefinedExtension
	}
}
