package snooty

import (
	add_code_examples "gdcd/add-code-examples"
	"gdcd/types"
)

func MakeIoCodeBlockForTesting(includeInputLang bool, includeChildCodeNodeLang bool, language string, includeFilepath bool, includeInputDirective bool, includeChildCodeNode bool, inputNotInFirstPosition bool, childCodeNodeNotInFirstPosition bool) types.ASTNode {
	emptyDirective := types.ASTNode{
		Type:           "directive",
		Position:       types.Position{},
		Children:       nil,
		Value:          "Random",
		Lang:           "",
		Copyable:       false,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: nil,
		LineNumbers:    false,
	}
	childCodeNode := types.ASTNode{
		Type:           "code",
		Position:       types.Position{},
		Children:       nil,
		Value:          "some code here",
		Lang:           "",
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
	if includeChildCodeNodeLang {
		childCodeNode.Lang = language
	}
	fileExtension := add_code_examples.GetFileExtensionFromStringLang(language)
	filepath := "filename" + fileExtension
	inputDirectiveArg := types.TextNode{
		Type:     "text",
		Position: types.Position{},
		Value:    filepath,
		Children: nil,
	}
	inputDirective := types.ASTNode{
		Type:           "directive",
		Position:       types.Position{},
		Children:       []types.ASTNode{},
		Value:          "",
		Lang:           "",
		Copyable:       false,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "input",
		Argument:       []types.TextNode{},
		Options:        nil,
		EmphasizeLines: nil,
		LineNumbers:    false,
	}
	if includeChildCodeNode {
		if childCodeNodeNotInFirstPosition {
			childCodeNode.Children = []types.ASTNode{emptyDirective, childCodeNode}
		} else {
			childCodeNode.Children = []types.ASTNode{childCodeNode}
		}
	}

	if includeFilepath {
		inputDirective.Argument = []types.TextNode{inputDirectiveArg}
	}
	if includeInputLang {
		options := make(map[string]interface{})
		options["language"] = language
	}
	ioCodeBlockNode := types.ASTNode{
		Type:           "directive",
		Position:       types.Position{},
		Children:       []types.ASTNode{},
		Value:          "",
		Lang:           "",
		Copyable:       false,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "io-code-block",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: nil,
	}
	if includeInputDirective {
		if inputNotInFirstPosition {
			ioCodeBlockNode.Children = []types.ASTNode{emptyDirective, inputDirective}
		} else {
			ioCodeBlockNode.Children = []types.ASTNode{inputDirective}
		}
	}
	return ioCodeBlockNode
}
