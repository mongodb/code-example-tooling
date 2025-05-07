package add_code_examples

import (
	"common"
	"context"
	"log"
	"testing"

	"github.com/tmc/langchaingo/llms/ollama"
)

func TestCategorizeTextSnippet(t *testing.T) {
	ctx := context.Background()
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	type args struct {
		contents string
		llm      *ollama.LLM
		ctx      context.Context
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{common.NonMongoCommand, args{contents: GetCodeExampleForTesting(common.NonMongoCommand, common.Text), llm: llm, ctx: ctx}, common.NonMongoCommand},
		{common.UsageExample, args{contents: GetCodeExampleForTesting(common.UsageExample, common.Text), llm: llm, ctx: ctx}, common.UsageExample},
		{common.SyntaxExample, args{contents: GetCodeExampleForTesting(common.SyntaxExample, common.Text), llm: llm, ctx: ctx}, common.SyntaxExample},
		{common.ExampleConfigurationObject, args{contents: GetCodeExampleForTesting(common.ExampleConfigurationObject, common.Text), llm: llm, ctx: ctx}, common.ExampleConfigurationObject},
		{common.ExampleReturnObject, args{contents: GetCodeExampleForTesting(common.ExampleReturnObject, common.Text), llm: llm, ctx: ctx}, common.ExampleReturnObject},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategorizeTextSnippet(tt.args.contents, tt.args.llm, tt.args.ctx); got != tt.want {
				t.Errorf("CategorizeTextSnippet() = got %v, want %v", got, tt.want)
			}
		})
	}
}
