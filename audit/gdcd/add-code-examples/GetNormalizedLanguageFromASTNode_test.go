package add_code_examples

import (
	"gdcd/types"
	"testing"
)

func TestGetLanguage(t *testing.T) {

	type args struct {
		snootyNode types.ASTNode
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{Bash, args{GetAstCodeNodeForLangForTesting(Bash)}, Bash},
		{C, args{GetAstCodeNodeForLangForTesting(C)}, C},
		{CPP, args{GetAstCodeNodeForLangForTesting(CPP)}, CPP},
		{CSharp, args{GetAstCodeNodeForLangForTesting(CSharp)}, CSharp},
		{Go, args{GetAstCodeNodeForLangForTesting(Go)}, Go},
		{Java, args{GetAstCodeNodeForLangForTesting(Java)}, Java},
		{JavaScript, args{GetAstCodeNodeForLangForTesting(JavaScript)}, JavaScript},
		{JSON, args{GetAstCodeNodeForLangForTesting(JSON)}, JSON},
		{Kotlin, args{GetAstCodeNodeForLangForTesting(Kotlin)}, Kotlin},
		{PHP, args{GetAstCodeNodeForLangForTesting(PHP)}, PHP},
		{Python, args{GetAstCodeNodeForLangForTesting(Python)}, Python},
		{Ruby, args{GetAstCodeNodeForLangForTesting(Ruby)}, Ruby},
		{Rust, args{GetAstCodeNodeForLangForTesting(Rust)}, Rust},
		{Scala, args{GetAstCodeNodeForLangForTesting(Scala)}, Scala},
		{Shell, args{GetAstCodeNodeForLangForTesting(Shell)}, Shell},
		{Swift, args{GetAstCodeNodeForLangForTesting(Swift)}, Swift},
		{Text, args{GetAstCodeNodeForLangForTesting(Text)}, Text},
		{TypeScript, args{GetAstCodeNodeForLangForTesting(TypeScript)}, TypeScript},
		{Undefined, args{GetAstCodeNodeForLangForTesting(Undefined)}, Undefined},
		{XML, args{GetAstCodeNodeForLangForTesting(XML)}, XML},
		{YAML, args{GetAstCodeNodeForLangForTesting(YAML)}, YAML},
		{"Empty string", args{GetAstCodeNodeForLangForTesting("")}, Undefined},
		{"console", args{GetAstCodeNodeForLangForTesting("console")}, Shell},
		{"cs", args{GetAstCodeNodeForLangForTesting("cs")}, CSharp},
		{"golang", args{GetAstCodeNodeForLangForTesting("golang")}, Go},
		{"http", args{GetAstCodeNodeForLangForTesting("http")}, Text},
		{"ini", args{GetAstCodeNodeForLangForTesting("ini")}, Text},
		{"js", args{GetAstCodeNodeForLangForTesting("js")}, JavaScript},
		{"none", args{GetAstCodeNodeForLangForTesting("none")}, Undefined},
		{"sh", args{GetAstCodeNodeForLangForTesting("sh")}, Shell},
		{"json\\n :copyable: false", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: false")}, JSON},
		{"json\\n :copyable: true", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: true")}, JSON},
		{"Some other non-normalized lang", args{GetAstCodeNodeForLangForTesting("Some other non-normalized lang")}, Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNormalizedLanguageFromASTNode(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetNormalizedLanguageFromASTNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
