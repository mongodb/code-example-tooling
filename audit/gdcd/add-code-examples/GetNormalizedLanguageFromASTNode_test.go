package add_code_examples

import (
	"common"
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
		{common.Bash, args{GetAstCodeNodeForLangForTesting(common.Bash)}, common.Bash},
		{common.C, args{GetAstCodeNodeForLangForTesting(common.C)}, common.C},
		{common.CPP, args{GetAstCodeNodeForLangForTesting(common.CPP)}, common.CPP},
		{common.CSharp, args{GetAstCodeNodeForLangForTesting(common.CSharp)}, common.CSharp},
		{common.Go, args{GetAstCodeNodeForLangForTesting(common.Go)}, common.Go},
		{common.Java, args{GetAstCodeNodeForLangForTesting(common.Java)}, common.Java},
		{common.JavaScript, args{GetAstCodeNodeForLangForTesting(common.JavaScript)}, common.JavaScript},
		{common.JSON, args{GetAstCodeNodeForLangForTesting(common.JSON)}, common.JSON},
		{common.Kotlin, args{GetAstCodeNodeForLangForTesting(common.Kotlin)}, common.Kotlin},
		{common.PHP, args{GetAstCodeNodeForLangForTesting(common.PHP)}, common.PHP},
		{common.Python, args{GetAstCodeNodeForLangForTesting(common.Python)}, common.Python},
		{common.Ruby, args{GetAstCodeNodeForLangForTesting(common.Ruby)}, common.Ruby},
		{common.Rust, args{GetAstCodeNodeForLangForTesting(common.Rust)}, common.Rust},
		{common.Scala, args{GetAstCodeNodeForLangForTesting(common.Scala)}, common.Scala},
		{common.Shell, args{GetAstCodeNodeForLangForTesting(common.Shell)}, common.Shell},
		{common.Swift, args{GetAstCodeNodeForLangForTesting(common.Swift)}, common.Swift},
		{common.Text, args{GetAstCodeNodeForLangForTesting(common.Text)}, common.Text},
		{common.TypeScript, args{GetAstCodeNodeForLangForTesting(common.TypeScript)}, common.TypeScript},
		{common.Undefined, args{GetAstCodeNodeForLangForTesting(common.Undefined)}, common.Undefined},
		{common.XML, args{GetAstCodeNodeForLangForTesting(common.XML)}, common.XML},
		{common.YAML, args{GetAstCodeNodeForLangForTesting(common.YAML)}, common.YAML},
		{"Empty string", args{GetAstCodeNodeForLangForTesting("")}, common.Undefined},
		{"console", args{GetAstCodeNodeForLangForTesting("console")}, common.Shell},
		{"cs", args{GetAstCodeNodeForLangForTesting("cs")}, common.CSharp},
		{"golang", args{GetAstCodeNodeForLangForTesting("golang")}, common.Go},
		{"http", args{GetAstCodeNodeForLangForTesting("http")}, common.Text},
		{"ini", args{GetAstCodeNodeForLangForTesting("ini")}, common.Text},
		{"js", args{GetAstCodeNodeForLangForTesting("js")}, common.JavaScript},
		{"none", args{GetAstCodeNodeForLangForTesting("none")}, common.Undefined},
		{"sh", args{GetAstCodeNodeForLangForTesting("sh")}, common.Shell},
		{"json\\n :copyable: false", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: false")}, common.JSON},
		{"json\\n :copyable: true", args{GetAstCodeNodeForLangForTesting("json\\n :copyable: true")}, common.JSON},
		{"Some other non-normalized lang", args{GetAstCodeNodeForLangForTesting("Some other non-normalized lang")}, common.Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNormalizedLanguageFromASTNode(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetNormalizedLanguageFromASTNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
