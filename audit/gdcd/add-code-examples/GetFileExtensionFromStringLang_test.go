package add_code_examples

import "testing"

func TestGetFileExtensionFromStringLang(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{Bash, args{Bash}, BashExtension},
		{C, args{C}, CExtension},
		{CPP, args{CPP}, CPPExtension},
		{CSharp, args{CSharp}, CSharpExtension},
		{Go, args{Go}, GoExtension},
		{Java, args{Java}, JavaExtension},
		{JavaScript, args{JavaScript}, JavaScriptExtension},
		{JSON, args{JSON}, JSONExtension},
		{Kotlin, args{Kotlin}, KotlinExtension},
		{PHP, args{PHP}, PHPExtension},
		{Python, args{Python}, PythonExtension},
		{Ruby, args{Ruby}, RubyExtension},
		{Rust, args{Rust}, RustExtension},
		{Scala, args{Scala}, ScalaExtension},
		{Shell, args{Shell}, ShellExtension},
		{Swift, args{Swift}, SwiftExtension},
		{Text, args{Text}, TextExtension},
		{TypeScript, args{TypeScript}, TypeScriptExtension},
		{Undefined, args{Undefined}, UndefinedExtension},
		{XML, args{XML}, XMLExtension},
		{YAML, args{YAML}, YAMLExtension},
		{"Empty string", args{""}, UndefinedExtension},
		{"console", args{"console"}, ShellExtension},
		{"cs", args{"cs"}, CSharpExtension},
		{"golang", args{"golang"}, GoExtension},
		{"http", args{"http"}, TextExtension},
		{"ini", args{"ini"}, TextExtension},
		{"js", args{"js"}, JavaScriptExtension},
		{"none", args{"none"}, UndefinedExtension},
		{"sh", args{"sh"}, ShellExtension},
		{"json\\n :copyable: false", args{"json\\n :copyable: false"}, JSONExtension},
		{"json\\n :copyable: true", args{"json\\n :copyable: true"}, JSONExtension},
		{"Some other non-normalized lang", args{"Some other non-normalized lang"}, UndefinedExtension},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtensionFromStringLang(tt.args.language); got != tt.want {
				t.Errorf("GetFileExtensionFromStringLang() = %v, want %v", got, tt.want)
			}
		})
	}
}
