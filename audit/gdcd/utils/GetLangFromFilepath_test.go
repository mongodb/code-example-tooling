package utils

import (
	"common"
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
		{common.Bash, args{"filename.sh"}, common.Shell},
		{common.C, args{"filename.c"}, common.C},
		{common.CPP, args{"filename.cpp"}, common.CPP},
		{common.CSharp, args{"filename.cs"}, common.CSharp},
		{common.Go, args{"filename.go"}, common.Go},
		{common.Java, args{"filename.java"}, common.Java},
		{common.JavaScript, args{"filename.js"}, common.JavaScript},
		{common.JSON, args{"filename.json"}, common.JSON},
		{common.Kotlin, args{"filename.kt"}, common.Kotlin},
		{common.PHP, args{"filename.php"}, common.PHP},
		{common.Python, args{"filename.py"}, common.Python},
		{common.Ruby, args{"filename.rb"}, common.Ruby},
		{common.Rust, args{"filename.rs"}, common.Rust},
		{common.Scala, args{"filename.scala"}, common.Scala},
		{common.Shell, args{"filename.sh"}, common.Shell},
		{common.Swift, args{"filename.swift"}, common.Swift},
		{common.Text, args{"filename.txt"}, common.Text},
		{common.TypeScript, args{"filename.ts"}, common.TypeScript},
		{common.XML, args{"filename.xml"}, common.XML},
		{common.YAML, args{"filename.yaml"}, common.YAML},
		{"Extension not present in our switch should hit default case", args{"filename.ini"}, common.Undefined},
		{"Invalid filepath should hit default case", args{"filename-that-has-no-extension"}, common.Undefined},
		{"Empty string filepath should hit default case", args{""}, common.Undefined},
		{"Filepath should properly handle multi-segment filepath", args{"/dir/other-dir/filename.c"}, common.C},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLangFromFilepath(tt.args.filepath); got != tt.want {
				t.Errorf("GetLangFromFilepath() = %v, want %v", got, tt.want)
			}
		})
	}
}
