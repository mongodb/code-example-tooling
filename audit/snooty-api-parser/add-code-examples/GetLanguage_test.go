package add_code_examples

import (
	"snooty-api-parser/types"
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
		{Bash, args{GetAstNodeForLangForTesting(Bash)}, Bash},
		{C, args{GetAstNodeForLangForTesting(C)}, C},
		{CPP, args{GetAstNodeForLangForTesting(CPP)}, CPP},
		{CSharp, args{GetAstNodeForLangForTesting(CSharp)}, CSharp},
		{Go, args{GetAstNodeForLangForTesting(Go)}, Go},
		{Java, args{GetAstNodeForLangForTesting(Java)}, Java},
		{JavaScript, args{GetAstNodeForLangForTesting(JavaScript)}, JavaScript},
		{JSON, args{GetAstNodeForLangForTesting(JSON)}, JSON},
		{Kotlin, args{GetAstNodeForLangForTesting(Kotlin)}, Kotlin},
		{PHP, args{GetAstNodeForLangForTesting(PHP)}, PHP},
		{Python, args{GetAstNodeForLangForTesting(Python)}, Python},
		{Ruby, args{GetAstNodeForLangForTesting(Ruby)}, Ruby},
		{Rust, args{GetAstNodeForLangForTesting(Rust)}, Rust},
		{Scala, args{GetAstNodeForLangForTesting(Scala)}, Scala},
		{Shell, args{GetAstNodeForLangForTesting(Shell)}, Shell},
		{Swift, args{GetAstNodeForLangForTesting(Swift)}, Swift},
		{Text, args{GetAstNodeForLangForTesting(Text)}, Text},
		{TypeScript, args{GetAstNodeForLangForTesting(TypeScript)}, TypeScript},
		{Undefined, args{GetAstNodeForLangForTesting(Undefined)}, Undefined},
		{XML, args{GetAstNodeForLangForTesting(XML)}, XML},
		{YAML, args{GetAstNodeForLangForTesting(YAML)}, YAML},
		{"Empty string", args{GetAstNodeForLangForTesting("")}, Undefined},
		{"console", args{GetAstNodeForLangForTesting("console")}, Shell},
		{"cs", args{GetAstNodeForLangForTesting("cs")}, CSharp},
		{"golang", args{GetAstNodeForLangForTesting("golang")}, Go},
		{"http", args{GetAstNodeForLangForTesting("http")}, Text},
		{"ini", args{GetAstNodeForLangForTesting("ini")}, Text},
		{"js", args{GetAstNodeForLangForTesting("js")}, JavaScript},
		{"none", args{GetAstNodeForLangForTesting("none")}, Undefined},
		{"sh", args{GetAstNodeForLangForTesting("sh")}, Shell},
		{"json\\n :copyable: false", args{GetAstNodeForLangForTesting("json\\n :copyable: false")}, JSON},
		{"json\\n :copyable: true", args{GetAstNodeForLangForTesting("json\\n :copyable: true")}, JSON},
		{"Some other non-normalized lang", args{GetAstNodeForLangForTesting("Some other non-normalized lang")}, Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLanguage(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}
