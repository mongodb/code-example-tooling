package snooty

import (
	"snooty-api-parser/add-code-examples"
	"snooty-api-parser/snooty/test-data"
	"snooty-api-parser/types"
	"testing"
)

func TestGetLangForIoCodeBlock(t *testing.T) {
	type args struct {
		snootyNode types.ASTNode
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Handles iocodeblock with input lang", args{test_data.MakeIoCodeBlockForTesting(true, true, add_code_examples.C, true, true, true, false, false)}, add_code_examples.C},
		{"Handles iocodeblock with no input lang using child code node lang", args{test_data.MakeIoCodeBlockForTesting(false, true, add_code_examples.C, true, true, true, false, false)}, add_code_examples.C},
		{"Handles iocodeblock with no input lang and no child code node lang using filepath", args{test_data.MakeIoCodeBlockForTesting(false, false, add_code_examples.C, true, true, true, false, false)}, add_code_examples.C},
		{"Lang should be undefined if no other conditions provide lang", args{test_data.MakeIoCodeBlockForTesting(false, false, add_code_examples.C, false, true, true, false, false)}, add_code_examples.Undefined},
		{"Handles no input directive", args{test_data.MakeIoCodeBlockForTesting(true, true, add_code_examples.C, true, false, true, false, false)}, add_code_examples.Undefined},
		{"Handles no child code node directive", args{test_data.MakeIoCodeBlockForTesting(false, true, add_code_examples.C, true, true, false, false, false)}, add_code_examples.C},
		{"Handles input directive not in first position", args{test_data.MakeIoCodeBlockForTesting(true, true, add_code_examples.C, true, true, true, true, false)}, add_code_examples.C},
		{"Handles child code node directive not in first position", args{test_data.MakeIoCodeBlockForTesting(false, true, add_code_examples.C, true, true, true, false, true)}, add_code_examples.C},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLangForIoCodeBlock(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetLangForIoCodeBlock() = got %v, want %v", got, tt.want)
			}
		})
	}
}
