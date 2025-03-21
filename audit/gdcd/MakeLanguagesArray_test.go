package main

// TODO: Refactor for existing languages structure
//func TestIoCodeBlockLanguagesCountCorrectly(t *testing.T) {
//	swiftIoCodeBlock := test_data.MakeIoCodeBlockForTesting(true, true, add_code_examples.Swift, true, true, true, false, false)
//	goIoCodeBlock := test_data.MakeIoCodeBlockForTesting(true, true, add_code_examples.Go, true, true, true, false, false)
//	anotherGoIoCodeBlock := test_data.MakeIoCodeBlockForTesting(true, true, add_code_examples.Go, true, true, true, false, false)
//	languagesArray := MakeLanguagesArray([]common.CodeNode{}, []types.ASTNode{}, []types.ASTNode{swiftIoCodeBlock, goIoCodeBlock, anotherGoIoCodeBlock})
//	gotSwiftCount := languagesArray[add_code_examples.Swift].IOCodeBlock
//	expectedSwiftCount := 1
//	if gotSwiftCount != expectedSwiftCount {
//		t.Errorf("MakeLanguagesArray() = for Swift io-code-block count, got %d, want %d", gotSwiftCount, expectedSwiftCount)
//	}
//	gotGoCount := languagesArray[add_code_examples.Go].IOCodeBlock
//	expectedGoCount := 2
//	if gotGoCount != expectedGoCount {
//		t.Errorf("MakeLanguagesArray() = for Go io-code-block count, got %d, want %d", gotGoCount, expectedGoCount)
//	}
//	gotLiteralIncludeSwiftCount := languagesArray[add_code_examples.Swift].LiteralIncludes
//	gotCodeNodeSwiftCount := languagesArray[add_code_examples.Swift].Total
//	expectedOtherNodeTypeCount := 0
//	if gotLiteralIncludeSwiftCount != expectedOtherNodeTypeCount {
//		t.Errorf("MakeLanguagesArray() = for Swift literalinclude count, got %d, want %d", gotLiteralIncludeSwiftCount, expectedOtherNodeTypeCount)
//	}
//	if gotCodeNodeSwiftCount != expectedOtherNodeTypeCount {
//		t.Errorf("MakeLanguagesArray() = for Swift code node count, got %d, want %d", gotCodeNodeSwiftCount, expectedOtherNodeTypeCount)
//	}
//}
//
//func TestLiteralIncludeNodeLanguagesCountCorrectly(t *testing.T) {
//	swiftLiteralInclude := test_data.MakeLiteralIncludeNodeForTesting(true, add_code_examples.Swift, true)
//	goLiteralInclude := test_data.MakeLiteralIncludeNodeForTesting(true, add_code_examples.Go, true)
//	anotherGoLiteralInclude := test_data.MakeLiteralIncludeNodeForTesting(true, add_code_examples.Go, true)
//	languagesArray := MakeLanguagesArray([]common.CodeNode{}, []types.ASTNode{swiftLiteralInclude, goLiteralInclude, anotherGoLiteralInclude}, []types.ASTNode{})
//	gotSwiftCount := languagesArray[add_code_examples.Swift].LiteralIncludes
//	expectedSwiftCount := 1
//	if gotSwiftCount != expectedSwiftCount {
//		t.Errorf("MakeLanguagesArray() = for Swift literalinclude count, got %d, want %d", gotSwiftCount, expectedSwiftCount)
//	}
//	gotGoCount := languagesArray[add_code_examples.Go].LiteralIncludes
//	expectedGoCount := 2
//	if gotGoCount != expectedGoCount {
//		t.Errorf("MakeLanguagesArray() = for Go literalinclude count, got %d, want %d", gotGoCount, expectedGoCount)
//	}
//	gotIoCodeBlockSwiftCount := languagesArray[add_code_examples.Swift].IOCodeBlock
//	gotCodeNodeSwiftCount := languagesArray[add_code_examples.Swift].Total
//	expectedOtherNodeTypeCount := 0
//	if gotIoCodeBlockSwiftCount != expectedOtherNodeTypeCount {
//		t.Errorf("MakeLanguagesArray() = for Swift io-code-block count, got %d, want %d", gotIoCodeBlockSwiftCount, expectedOtherNodeTypeCount)
//	}
//	if gotCodeNodeSwiftCount != expectedOtherNodeTypeCount {
//		t.Errorf("MakeLanguagesArray() = for Swift code node count, got %d, want %d", gotCodeNodeSwiftCount, expectedOtherNodeTypeCount)
//	}
//}
//
//func TestCodeNodeLanguagesCountCorrectly(t *testing.T) {
//	swiftCodeNode := test_data.MakeCodeNodeForTesting(add_code_examples.Swift, add_code_examples.SyntaxExample)
//	goCodeNode := test_data.MakeCodeNodeForTesting(add_code_examples.Go, add_code_examples.UsageExample)
//	anotherGoCodeNode := test_data.MakeCodeNodeForTesting(add_code_examples.Go, add_code_examples.SyntaxExample)
//	languagesArray := MakeLanguagesArray([]types.CodeNode{swiftCodeNode, goCodeNode, anotherGoCodeNode}, []types.ASTNode{}, []types.ASTNode{})
//	gotSwiftCount := languagesArray[add_code_examples.Swift].Total
//	expectedSwiftCount := 1
//	if gotSwiftCount != expectedSwiftCount {
//		t.Errorf("MakeLanguagesArray() = for Swift code node count, got %d, want %d", gotSwiftCount, expectedSwiftCount)
//	}
//	gotGoCount := languagesArray[add_code_examples.Go].Total
//	expectedGoCount := 2
//	if gotSwiftCount != expectedSwiftCount {
//		t.Errorf("MakeLanguagesArray() = for Go code node count, got %d, want %d", gotGoCount, expectedGoCount)
//	}
//	gotLiteralIncludeSwiftCount := languagesArray[add_code_examples.Swift].LiteralIncludes
//	gotIoCodeBlockSwiftCount := languagesArray[add_code_examples.Swift].IOCodeBlock
//	expectedOtherNodeTypeCount := 0
//	if gotLiteralIncludeSwiftCount != expectedOtherNodeTypeCount {
//		t.Errorf("MakeLanguagesArray() = for Swift literalinclude count, got %d, want %d", gotLiteralIncludeSwiftCount, expectedOtherNodeTypeCount)
//	}
//	if gotIoCodeBlockSwiftCount != expectedOtherNodeTypeCount {
//		t.Errorf("MakeLanguagesArray() = for Swift io-code-block count, got %d, want %d", gotIoCodeBlockSwiftCount, expectedOtherNodeTypeCount)
//	}
//}
