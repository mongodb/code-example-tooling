package add_code_examples

import "snooty-api-parser/types"

func GetAstCodeNodeForLangForTesting(lang string) types.ASTNode {
	return types.ASTNode{
		Type:           "code",
		Position:       types.Position{Start: types.PositionLine{Line: 51}},
		Children:       nil,
		Value:          "SomeValue",
		Lang:           lang,
		Copyable:       false,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: types.EmphasizeLines{14, 16},
	}
}
