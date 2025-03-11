package add_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
	"testing"
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
		// TODO: Add test cases.
		{NonMongoCommand, args{contents: GetCodeExampleForTesting(NonMongoCommand, Text), llm: llm, ctx: ctx}, NonMongoCommand},
		{UsageExample, args{contents: GetCodeExampleForTesting(UsageExample, Text), llm: llm, ctx: ctx}, UsageExample},
		{SyntaxExample, args{contents: GetCodeExampleForTesting(SyntaxExample, Text), llm: llm, ctx: ctx}, SyntaxExample},
		{ExampleConfigurationObject, args{contents: GetCodeExampleForTesting(ExampleConfigurationObject, Text), llm: llm, ctx: ctx}, ExampleConfigurationObject},
		{ExampleReturnObject, args{contents: GetCodeExampleForTesting(ExampleReturnObject, Text), llm: llm, ctx: ctx}, ExampleReturnObject},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategorizeTextSnippet(tt.args.contents, tt.args.llm, tt.args.ctx); got != tt.want {
				t.Errorf("CategorizeTextSnippet() = got %v, want %v", got, tt.want)
			}
		})
	}
}
