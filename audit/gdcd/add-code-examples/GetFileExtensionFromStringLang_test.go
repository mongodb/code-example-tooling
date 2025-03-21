package add_code_examples

import (
	"common"
	"testing"
)

func TestGetFileExtensionFromStringLang(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{common.Bash, args{common.Bash}, common.BashExtension},
		{common.C, args{common.C}, common.CExtension},
		{common.CPP, args{common.CPP}, common.CPPExtension},
		{common.CSharp, args{common.CSharp}, common.CSharpExtension},
		{common.Go, args{common.Go}, common.GoExtension},
		{common.Java, args{common.Java}, common.JavaExtension},
		{common.JavaScript, args{common.JavaScript}, common.JavaScriptExtension},
		{common.JSON, args{common.JSON}, common.JSONExtension},
		{common.Kotlin, args{common.Kotlin}, common.KotlinExtension},
		{common.PHP, args{common.PHP}, common.PHPExtension},
		{common.Python, args{common.Python}, common.PythonExtension},
		{common.Ruby, args{common.Ruby}, common.RubyExtension},
		{common.Rust, args{common.Rust}, common.RustExtension},
		{common.Scala, args{common.Scala}, common.ScalaExtension},
		{common.Shell, args{common.Shell}, common.ShellExtension},
		{common.Swift, args{common.Swift}, common.SwiftExtension},
		{common.Text, args{common.Text}, common.TextExtension},
		{common.TypeScript, args{common.TypeScript}, common.TypeScriptExtension},
		{common.Undefined, args{common.Undefined}, common.UndefinedExtension},
		{common.XML, args{common.XML}, common.XMLExtension},
		{common.YAML, args{common.YAML}, common.YAMLExtension},
		{"Empty string", args{""}, common.UndefinedExtension},
		{"console", args{"console"}, common.ShellExtension},
		{"cs", args{"cs"}, common.CSharpExtension},
		{"golang", args{"golang"}, common.GoExtension},
		{"http", args{"http"}, common.TextExtension},
		{"ini", args{"ini"}, common.TextExtension},
		{"js", args{"js"}, common.JavaScriptExtension},
		{"none", args{"none"}, common.UndefinedExtension},
		{"sh", args{"sh"}, common.ShellExtension},
		{"json\\n :copyable: false", args{"json\\n :copyable: false"}, common.JSONExtension},
		{"json\\n :copyable: true", args{"json\\n :copyable: true"}, common.JSONExtension},
		{"Some other non-normalized lang", args{"Some other non-normalized lang"}, common.UndefinedExtension},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtensionFromStringLang(tt.args.language); got != tt.want {
				t.Errorf("GetFileExtensionFromStringLang() = %v, want %v", got, tt.want)
			}
		})
	}
}
