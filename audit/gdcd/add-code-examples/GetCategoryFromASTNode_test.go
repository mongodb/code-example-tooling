package add_code_examples

import (
	"common"
	"gdcd/types"
	"testing"
)

func TestGetCategoryFromASTNode(t *testing.T) {
	type args struct {
		snootyNode types.ASTNode
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"syntax example", args{GetAstCodeNodeForCategoryForTesting("syntax example")}, common.SyntaxExample},
		{"syntaxexample", args{GetAstCodeNodeForCategoryForTesting("syntaxexample")}, common.SyntaxExample},
		{"syntax", args{GetAstCodeNodeForCategoryForTesting("syntax")}, common.SyntaxExample},
		{"Syntax Example", args{GetAstCodeNodeForCategoryForTesting("Syntax Example")}, common.SyntaxExample},
		{"usage example", args{GetAstCodeNodeForCategoryForTesting("usage example")}, common.UsageExample},
		{"usageexample", args{GetAstCodeNodeForCategoryForTesting("usageexample")}, common.UsageExample},
		{"usage", args{GetAstCodeNodeForCategoryForTesting("usage")}, common.UsageExample},
		{"example return object", args{GetAstCodeNodeForCategoryForTesting("example return object")}, common.ExampleReturnObject},
		{"return example", args{GetAstCodeNodeForCategoryForTesting("return example")}, common.ExampleReturnObject},
		{"return object", args{GetAstCodeNodeForCategoryForTesting("return object")}, common.ExampleReturnObject},
		{"example return", args{GetAstCodeNodeForCategoryForTesting("example return")}, common.ExampleReturnObject},
		{"examplereturnobject", args{GetAstCodeNodeForCategoryForTesting("examplereturnobject")}, common.ExampleReturnObject},
		{"return", args{GetAstCodeNodeForCategoryForTesting("return")}, common.ExampleReturnObject},
		{"example configuration object", args{GetAstCodeNodeForCategoryForTesting("example configuration object")}, common.ExampleConfigurationObject},
		{"configuration example", args{GetAstCodeNodeForCategoryForTesting("configuration example")}, common.ExampleConfigurationObject},
		{"configuration object", args{GetAstCodeNodeForCategoryForTesting("configuration object")}, common.ExampleConfigurationObject},
		{"example configuration", args{GetAstCodeNodeForCategoryForTesting("example configuration")}, common.ExampleConfigurationObject},
		{"exampleconfigurationobject", args{GetAstCodeNodeForCategoryForTesting("exampleconfigurationobject")}, common.ExampleConfigurationObject},
		{"configuration", args{GetAstCodeNodeForCategoryForTesting("configuration")}, common.ExampleConfigurationObject},
		{"non mongodb command", args{GetAstCodeNodeForCategoryForTesting("non mongodb command")}, common.NonMongoCommand},
		{"third party command", args{GetAstCodeNodeForCategoryForTesting("third party command")}, common.NonMongoCommand},
		{"some other category", args{GetAstCodeNodeForCategoryForTesting("some other category")}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCategoryFromASTNode(tt.args.snootyNode); got != tt.want {
				t.Errorf("GetCategoryFromASTNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
