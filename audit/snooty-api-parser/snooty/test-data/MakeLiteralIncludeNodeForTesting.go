package test_data

import (
	add_code_examples "snooty-api-parser/add-code-examples"
	"snooty-api-parser/types"
)

func MakeLiteralIncludeNodeForTesting(includeLang bool, language string, includeFilepath bool) types.ASTNode {
	childCodeNode := types.ASTNode{
		Type:           "code",
		Position:       types.Position{},
		Children:       nil,
		Value:          "some code here",
		Lang:           language,
		Copyable:       true,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: nil,
	}
	literalIncludeNode := types.ASTNode{
		Type:           "directive",
		Position:       types.Position{},
		Children:       []types.ASTNode{childCodeNode},
		Value:          "",
		Lang:           "",
		Copyable:       false,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "literalinclude",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: nil,
	}
	if includeLang {
		literalIncludeNode.Lang = language
	}
	if includeFilepath {
		extension := add_code_examples.GetFileExtension(childCodeNode)
		value := "filename" + extension
		argument := types.TextNode{
			Type:     "text",
			Position: types.Position{},
			Value:    value,
			Children: nil,
		}
		literalIncludeNode.Argument = []types.TextNode{argument}
	}
	return literalIncludeNode
}
