package snooty

import (
	"reflect"
	"snooty-api-parser/types"
	"testing"
)

func TestGetMetaKeywords(t *testing.T) {
	type args struct {
		nodes []types.ASTNode
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Finds and correctly splits keywords", args{LoadASTNodeTestDataFromFile(t, "page-with-code-nodes.json")}, []string{"client", "ssl", "tls", "localhost"}},
		{"Handles meta directive with no keywords", args{LoadASTNodeTestDataFromFile(t, "page-with-meta-no-keywords.json")}, nil},
		{"Handles no meta directive", args{LoadASTNodeTestDataFromFile(t, "page-with-no-meta.json")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMetaKeywords(tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetaKeywords() = %v, want %v", got, tt.want)
			}
		})
	}
}
