package add_code_examples

import "gdcd/types"

func GetAstCodeNodeForCategoryForTesting(category string) types.ASTNode {
	return types.ASTNode{
		Type:           "code",
		Position:       types.Position{Start: types.PositionLine{Line: 51}},
		Children:       nil,
		Value:          "SomeValue",
		Lang:           "javascript",
		Copyable:       false,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: types.EmphasizeLines{14, 16},
		Category:       category,
	}
}
