package snooty

import (
	"common"
	test_data "gdcd/snooty/test-data"
	"gdcd/types"
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
		{"Handles iocodeblock with input lang", args{test_data.MakeIoCodeBlockForTesting(true, true, common.C, true, true, true, false, false)}, common.C},
		{"Handles iocodeblock with no input lang using child code node lang", args{test_data.MakeIoCodeBlockForTesting(false, true, common.C, true, true, true, false, false)}, common.C},
		{"Handles iocodeblock with no input lang and no child code node lang using filepath", args{test_data.MakeIoCodeBlockForTesting(false, false, common.C, true, true, true, false, false)}, common.C},
		{"Lang should be undefined if no other conditions provide lang", args{test_data.MakeIoCodeBlockForTesting(false, false, common.C, false, true, true, false, false)}, common.Undefined},
		{"Handles no input directive", args{test_data.MakeIoCodeBlockForTesting(true, true, common.C, true, false, true, false, false)}, common.Undefined},
		{"Handles no child code node directive", args{test_data.MakeIoCodeBlockForTesting(false, true, common.C, true, true, false, false, false)}, common.C},
		{"Handles input directive not in first position", args{test_data.MakeIoCodeBlockForTesting(true, true, common.C, true, true, true, true, false)}, common.C},
		{"Handles child code node directive not in first position", args{test_data.MakeIoCodeBlockForTesting(false, true, common.C, true, true, true, false, true)}, common.C},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLangForIoCodeBlock(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetLangForIoCodeBlock() = got %v, want %v", got, tt.want)
			}
		})
	}
}
