package add_code_examples

import (
	"common"
	"testing"
)

func TestGetNormalizedLanguageFromString(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{common.Bash, args{common.Bash}, common.Bash},
		{common.C, args{common.C}, common.C},
		{common.CPP, args{common.CPP}, common.CPP},
		{common.CSharp, args{common.CSharp}, common.CSharp},
		{common.Go, args{common.Go}, common.Go},
		{common.Java, args{common.Java}, common.Java},
		{common.JavaScript, args{common.JavaScript}, common.JavaScript},
		{common.JSON, args{common.JSON}, common.JSON},
		{common.Kotlin, args{common.Kotlin}, common.Kotlin},
		{common.PHP, args{common.PHP}, common.PHP},
		{common.Python, args{common.Python}, common.Python},
		{common.Ruby, args{common.Ruby}, common.Ruby},
		{common.Rust, args{common.Rust}, common.Rust},
		{common.Scala, args{common.Scala}, common.Scala},
		{common.Shell, args{common.Shell}, common.Shell},
		{common.Swift, args{common.Swift}, common.Swift},
		{common.Text, args{common.Text}, common.Text},
		{common.TypeScript, args{common.TypeScript}, common.TypeScript},
		{common.Undefined, args{common.Undefined}, common.Undefined},
		{common.XML, args{common.XML}, common.XML},
		{common.YAML, args{common.YAML}, common.YAML},
		{"Empty string", args{""}, common.Undefined},
		{"console", args{"console"}, common.Shell},
		{"cs", args{"cs"}, common.CSharp},
		{"golang", args{"golang"}, common.Go},
		{"http", args{"http"}, common.Text},
		{"ini", args{"ini"}, common.Text},
		{"js", args{"js"}, common.JavaScript},
		{"none", args{"none"}, common.Undefined},
		{"sh", args{"sh"}, common.Shell},
		{"json\\n :copyable: false", args{"json\\n :copyable: false"}, common.JSON},
		{"json\\n :copyable: true", args{"json\\n :copyable: true"}, common.JSON},
		{"Some other non-normalized lang", args{"Some other non-normalized lang"}, common.Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNormalizedLanguageFromString(tt.args.language); got != tt.want {
				t.Errorf("GetNormalizedLanguageFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
