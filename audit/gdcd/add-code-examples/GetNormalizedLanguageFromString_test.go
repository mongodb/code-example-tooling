package add_code_examples

import "testing"

func TestGetNormalizedLanguageFromString(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{Bash, args{Bash}, Bash},
		{C, args{C}, C},
		{CPP, args{CPP}, CPP},
		{CSharp, args{CSharp}, CSharp},
		{Go, args{Go}, Go},
		{Java, args{Java}, Java},
		{JavaScript, args{JavaScript}, JavaScript},
		{JSON, args{JSON}, JSON},
		{Kotlin, args{Kotlin}, Kotlin},
		{PHP, args{PHP}, PHP},
		{Python, args{Python}, Python},
		{Ruby, args{Ruby}, Ruby},
		{Rust, args{Rust}, Rust},
		{Scala, args{Scala}, Scala},
		{Shell, args{Shell}, Shell},
		{Swift, args{Swift}, Swift},
		{Text, args{Text}, Text},
		{TypeScript, args{TypeScript}, TypeScript},
		{Undefined, args{Undefined}, Undefined},
		{XML, args{XML}, XML},
		{YAML, args{YAML}, YAML},
		{"Empty string", args{""}, Undefined},
		{"console", args{"console"}, Shell},
		{"cs", args{"cs"}, CSharp},
		{"golang", args{"golang"}, Go},
		{"http", args{"http"}, Text},
		{"ini", args{"ini"}, Text},
		{"js", args{"js"}, JavaScript},
		{"none", args{"none"}, Undefined},
		{"sh", args{"sh"}, Shell},
		{"json\\n :copyable: false", args{"json\\n :copyable: false"}, JSON},
		{"json\\n :copyable: true", args{"json\\n :copyable: true"}, JSON},
		{"Some other non-normalized lang", args{"Some other non-normalized lang"}, Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNormalizedLanguageFromString(tt.args.language); got != tt.want {
				t.Errorf("GetNormalizedLanguageFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
