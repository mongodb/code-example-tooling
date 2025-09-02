package snooty

import (
	"common"
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
		{"Handles literalinclude with specified lang", args{MakeLiteralIncludeNodeForTesting(true, common.C, false)}, common.C},
		{"Handles literalinclude with empty string lang using filepath", args{MakeLiteralIncludeNodeForTesting(false, common.C, true)}, common.C},
		{"Uses child code node lang when literalinclude empty string lang and no filepath", args{MakeLiteralIncludeNodeForTesting(false, common.C, false)}, common.C},
		{"Lang should be undefined if no other conditions provide lang", args{MakeLiteralIncludeNodeForTesting(true, common.Undefined, false)}, common.Undefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLangForLiteralInclude(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetLangForLiteralInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}
