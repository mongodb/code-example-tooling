package add_code_examples

import (
	"common"
)

func IsNewAppliedUsageExample(node common.CodeNode) bool {
	codeExampleCharacterCount := len([]rune(node.Code))
	if node.Category == common.UsageExample && codeExampleCharacterCount > 300 {
		return true
	}
	return false
}
