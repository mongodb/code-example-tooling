package add_code_examples

import (
	"snooty-api-parser/types"
	"testing"
)

func TestGetFileExtension(t *testing.T) {
	type args struct {
		snootyNode types.ASTNode
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{Bash, args{GetAstNodeForLangForTesting(Bash)}, BashExtension},
		{C, args{GetAstNodeForLangForTesting(C)}, CExtension},
		{CPP, args{GetAstNodeForLangForTesting(CPP)}, CPPExtension},
		{CSharp, args{GetAstNodeForLangForTesting(CSharp)}, CSharpExtension},
		{Go, args{GetAstNodeForLangForTesting(Go)}, GoExtension},
		{Java, args{GetAstNodeForLangForTesting(Java)}, JavaExtension},
		{JavaScript, args{GetAstNodeForLangForTesting(JavaScript)}, JavaScriptExtension},
		{JSON, args{GetAstNodeForLangForTesting(JSON)}, JSONExtension},
		{Kotlin, args{GetAstNodeForLangForTesting(Kotlin)}, KotlinExtension},
		{PHP, args{GetAstNodeForLangForTesting(PHP)}, PHPExtension},
		{Python, args{GetAstNodeForLangForTesting(Python)}, PythonExtension},
		{Ruby, args{GetAstNodeForLangForTesting(Ruby)}, RubyExtension},
		{Rust, args{GetAstNodeForLangForTesting(Rust)}, RustExtension},
		{Scala, args{GetAstNodeForLangForTesting(Scala)}, ScalaExtension},
		{Shell, args{GetAstNodeForLangForTesting(Shell)}, ShellExtension},
		{Swift, args{GetAstNodeForLangForTesting(Swift)}, SwiftExtension},
		{Text, args{GetAstNodeForLangForTesting(Text)}, TextExtension},
		{TypeScript, args{GetAstNodeForLangForTesting(TypeScript)}, TypeScriptExtension},
		{Undefined, args{GetAstNodeForLangForTesting(Undefined)}, UndefinedExtension},
		{XML, args{GetAstNodeForLangForTesting(XML)}, XMLExtension},
		{YAML, args{GetAstNodeForLangForTesting(YAML)}, YAMLExtension},
		{"Empty string", args{GetAstNodeForLangForTesting("")}, UndefinedExtension},
		{"console", args{GetAstNodeForLangForTesting("console")}, ShellExtension},
		{"cs", args{GetAstNodeForLangForTesting("cs")}, CSharpExtension},
		{"golang", args{GetAstNodeForLangForTesting("golang")}, GoExtension},
		{"http", args{GetAstNodeForLangForTesting("http")}, TextExtension},
		{"ini", args{GetAstNodeForLangForTesting("ini")}, TextExtension},
		{"js", args{GetAstNodeForLangForTesting("js")}, JavaScriptExtension},
		{"none", args{GetAstNodeForLangForTesting("none")}, UndefinedExtension},
		{"sh", args{GetAstNodeForLangForTesting("sh")}, ShellExtension},
		{"json\\n :copyable: false", args{GetAstNodeForLangForTesting("json\\n :copyable: false")}, JSONExtension},
		{"json\\n :copyable: true", args{GetAstNodeForLangForTesting("json\\n :copyable: true")}, JSONExtension},
		{"Some other non-normalized lang", args{GetAstNodeForLangForTesting("Some other non-normalized lang")}, UndefinedExtension},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtensionFromASTNode(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetFileExtensionFromASTNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
