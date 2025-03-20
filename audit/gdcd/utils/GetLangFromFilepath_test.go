package utils

import (
	"gdcd/add-code-examples"
	"testing"
)

func TestGetLangFromFilepath(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{add_code_examples.Bash, args{"filename.sh"}, add_code_examples.Shell},
		{add_code_examples.C, args{"filename.c"}, add_code_examples.C},
		{add_code_examples.CPP, args{"filename.cpp"}, add_code_examples.CPP},
		{add_code_examples.CSharp, args{"filename.cs"}, add_code_examples.CSharp},
		{add_code_examples.Go, args{"filename.go"}, add_code_examples.Go},
		{add_code_examples.Java, args{"filename.java"}, add_code_examples.Java},
		{add_code_examples.JavaScript, args{"filename.js"}, add_code_examples.JavaScript},
		{add_code_examples.JSON, args{"filename.json"}, add_code_examples.JSON},
		{add_code_examples.Kotlin, args{"filename.kt"}, add_code_examples.Kotlin},
		{add_code_examples.PHP, args{"filename.php"}, add_code_examples.PHP},
		{add_code_examples.Python, args{"filename.py"}, add_code_examples.Python},
		{add_code_examples.Ruby, args{"filename.rb"}, add_code_examples.Ruby},
		{add_code_examples.Rust, args{"filename.rs"}, add_code_examples.Rust},
		{add_code_examples.Scala, args{"filename.scala"}, add_code_examples.Scala},
		{add_code_examples.Shell, args{"filename.sh"}, add_code_examples.Shell},
		{add_code_examples.Swift, args{"filename.swift"}, add_code_examples.Swift},
		{add_code_examples.Text, args{"filename.txt"}, add_code_examples.Text},
		{add_code_examples.TypeScript, args{"filename.ts"}, add_code_examples.TypeScript},
		{add_code_examples.XML, args{"filename.xml"}, add_code_examples.XML},
		{add_code_examples.YAML, args{"filename.yaml"}, add_code_examples.YAML},
		{"Extension not present in our switch should hit default case", args{"filename.ini"}, add_code_examples.Undefined},
		{"Invalid filepath should hit default case", args{"filename-that-has-no-extension"}, add_code_examples.Undefined},
		{"Empty string filepath should hit default case", args{""}, add_code_examples.Undefined},
		{"Filepath should properly handle multi-segment filepath", args{"/dir/other-dir/filename.c"}, add_code_examples.C},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLangFromFilepath(tt.args.filepath); got != tt.want {
				t.Errorf("GetLangFromFilepath() = %v, want %v", got, tt.want)
			}
		})
	}
}
