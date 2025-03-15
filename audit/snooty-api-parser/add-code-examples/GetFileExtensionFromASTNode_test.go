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
		{Bash, args{GetAstCodeNodeForLangForTesting(Bash)}, BashExtension},
		{C, args{GetAstCodeNodeForLangForTesting(C)}, CExtension},
		{CPP, args{GetAstCodeNodeForLangForTesting(CPP)}, CPPExtension},
		{CSharp, args{GetAstCodeNodeForLangForTesting(CSharp)}, CSharpExtension},
		{Go, args{GetAstCodeNodeForLangForTesting(Go)}, GoExtension},
		{Java, args{GetAstCodeNodeForLangForTesting(Java)}, JavaExtension},
		{JavaScript, args{GetAstCodeNodeForLangForTesting(JavaScript)}, JavaScriptExtension},
		{JSON, args{GetAstCodeNodeForLangForTesting(JSON)}, JSONExtension},
		{Kotlin, args{GetAstCodeNodeForLangForTesting(Kotlin)}, KotlinExtension},
		{PHP, args{GetAstCodeNodeForLangForTesting(PHP)}, PHPExtension},
		{Python, args{GetAstCodeNodeForLangForTesting(Python)}, PythonExtension},
		{Ruby, args{GetAstCodeNodeForLangForTesting(Ruby)}, RubyExtension},
		{Rust, args{GetAstCodeNodeForLangForTesting(Rust)}, RustExtension},
		{Scala, args{GetAstCodeNodeForLangForTesting(Scala)}, ScalaExtension},
		{Shell, args{GetAstCodeNodeForLangForTesting(Shell)}, ShellExtension},
		{Swift, args{GetAstCodeNodeForLangForTesting(Swift)}, SwiftExtension},
		{Text, args{GetAstCodeNodeForLangForTesting(Text)}, TextExtension},
		{TypeScript, args{GetAstCodeNodeForLangForTesting(TypeScript)}, TypeScriptExtension},
		{Undefined, args{GetAstCodeNodeForLangForTesting(Undefined)}, UndefinedExtension},
		{XML, args{GetAstCodeNodeForLangForTesting(XML)}, XMLExtension},
		{YAML, args{GetAstCodeNodeForLangForTesting(YAML)}, YAMLExtension},
		{"Empty string", args{GetAstCodeNodeForLangForTesting("")}, UndefinedExtension},
		{"console", args{GetAstCodeNodeForLangForTesting("console")}, ShellExtension},
		{"cs", args{GetAstCodeNodeForLangForTesting("cs")}, CSharpExtension},
		{"golang", args{GetAstCodeNodeForLangForTesting("golang")}, GoExtension},
		{"http", args{GetAstCodeNodeForLangForTesting("http")}, TextExtension},
		{"ini", args{GetAstCodeNodeForLangForTesting("ini")}, TextExtension},
		{"js", args{GetAstCodeNodeForLangForTesting("js")}, JavaScriptExtension},
		{"none", args{GetAstCodeNodeForLangForTesting("none")}, UndefinedExtension},
		{"sh", args{GetAstCodeNodeForLangForTesting("sh")}, ShellExtension},
		{"json\\n :copyable: false", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: false")}, JSONExtension},
		{"json\\n :copyable: true", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: true")}, JSONExtension},
		{"Some other non-normalized lang", args{GetAstCodeNodeForLangForTesting("Some other non-normalized lang")}, UndefinedExtension},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtensionFromASTNode(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetFileExtensionFromASTNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
