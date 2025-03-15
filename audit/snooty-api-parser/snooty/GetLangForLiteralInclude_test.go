package snooty

import (
	add_code_examples "snooty-api-parser/add-code-examples"
	test_data "snooty-api-parser/snooty/test-data"
	"snooty-api-parser/types"
	"testing"
)

func TestGetLangForLiteralInclude(t *testing.T) {
	type args struct {
		snootyNode types.ASTNode
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Handles literalinclude with specified lang", args{test_data.MakeLiteralIncludeNodeForTesting(true, add_code_examples.C, false)}, add_code_examples.C},
		{"Handles literalinclude with empty string lang using filepath", args{test_data.MakeLiteralIncludeNodeForTesting(false, add_code_examples.C, true)}, add_code_examples.C},
		{"Uses child code node lang when literalinclude empty string lang and no filepath", args{test_data.MakeLiteralIncludeNodeForTesting(false, add_code_examples.C, false)}, add_code_examples.C},
		{"Lang should be undefined if no other conditions provide lang", args{test_data.MakeLiteralIncludeNodeForTesting(true, add_code_examples.Undefined, false)}, add_code_examples.Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLangForLiteralInclude(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetLangForLiteralInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}
