package snooty

import (
	"common"
	test_data "gdcd/snooty/test-data"
	"gdcd/types"
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
		{"Handles literalinclude with specified lang", args{test_data.MakeLiteralIncludeNodeForTesting(true, common.C, false)}, common.C},
		{"Handles literalinclude with empty string lang using filepath", args{test_data.MakeLiteralIncludeNodeForTesting(false, common.C, true)}, common.C},
		{"Uses child code node lang when literalinclude empty string lang and no filepath", args{test_data.MakeLiteralIncludeNodeForTesting(false, common.C, false)}, common.C},
		{"Lang should be undefined if no other conditions provide lang", args{test_data.MakeLiteralIncludeNodeForTesting(true, common.Undefined, false)}, common.Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLangForLiteralInclude(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetLangForLiteralInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}
