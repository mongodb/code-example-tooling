package add_code_examples

import (
	"common"
	"gdcd/types"
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
		{common.Bash, args{GetAstCodeNodeForLangForTesting(common.Bash)}, common.BashExtension},
		{common.C, args{GetAstCodeNodeForLangForTesting(common.C)}, common.CExtension},
		{common.CPP, args{GetAstCodeNodeForLangForTesting(common.CPP)}, common.CPPExtension},
		{common.CSharp, args{GetAstCodeNodeForLangForTesting(common.CSharp)}, common.CSharpExtension},
		{common.Go, args{GetAstCodeNodeForLangForTesting(common.Go)}, common.GoExtension},
		{common.Java, args{GetAstCodeNodeForLangForTesting(common.Java)}, common.JavaExtension},
		{common.JavaScript, args{GetAstCodeNodeForLangForTesting(common.JavaScript)}, common.JavaScriptExtension},
		{common.JSON, args{GetAstCodeNodeForLangForTesting(common.JSON)}, common.JSONExtension},
		{common.Kotlin, args{GetAstCodeNodeForLangForTesting(common.Kotlin)}, common.KotlinExtension},
		{common.PHP, args{GetAstCodeNodeForLangForTesting(common.PHP)}, common.PHPExtension},
		{common.Python, args{GetAstCodeNodeForLangForTesting(common.Python)}, common.PythonExtension},
		{common.Ruby, args{GetAstCodeNodeForLangForTesting(common.Ruby)}, common.RubyExtension},
		{common.Rust, args{GetAstCodeNodeForLangForTesting(common.Rust)}, common.RustExtension},
		{common.Scala, args{GetAstCodeNodeForLangForTesting(common.Scala)}, common.ScalaExtension},
		{common.Shell, args{GetAstCodeNodeForLangForTesting(common.Shell)}, common.ShellExtension},
		{common.Swift, args{GetAstCodeNodeForLangForTesting(common.Swift)}, common.SwiftExtension},
		{common.Text, args{GetAstCodeNodeForLangForTesting(common.Text)}, common.TextExtension},
		{common.TypeScript, args{GetAstCodeNodeForLangForTesting(common.TypeScript)}, common.TypeScriptExtension},
		{common.Undefined, args{GetAstCodeNodeForLangForTesting(common.Undefined)}, common.UndefinedExtension},
		{common.XML, args{GetAstCodeNodeForLangForTesting(common.XML)}, common.XMLExtension},
		{common.YAML, args{GetAstCodeNodeForLangForTesting(common.YAML)}, common.YAMLExtension},
		{"Empty string", args{GetAstCodeNodeForLangForTesting("")}, common.UndefinedExtension},
		{"console", args{GetAstCodeNodeForLangForTesting("console")}, common.ShellExtension},
		{"cs", args{GetAstCodeNodeForLangForTesting("cs")}, common.CSharpExtension},
		{"golang", args{GetAstCodeNodeForLangForTesting("golang")}, common.GoExtension},
		{"http", args{GetAstCodeNodeForLangForTesting("http")}, common.TextExtension},
		{"ini", args{GetAstCodeNodeForLangForTesting("ini")}, common.TextExtension},
		{"js", args{GetAstCodeNodeForLangForTesting("js")}, common.JavaScriptExtension},
		{"none", args{GetAstCodeNodeForLangForTesting("none")}, common.UndefinedExtension},
		{"sh", args{GetAstCodeNodeForLangForTesting("sh")}, common.ShellExtension},
		{"json\\n :copyable: false", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: false")}, common.JSONExtension},
		{"json\\n :copyable: true", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: true")}, common.JSONExtension},
		{"Some other non-normalized lang", args{GetAstCodeNodeForLangForTesting("Some other non-normalized lang")}, common.UndefinedExtension},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtensionFromASTNode(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetFileExtensionFromASTNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
