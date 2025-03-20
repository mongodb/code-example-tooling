package add_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
	"testing"
)

func TestCategorizeShellSnippet(t *testing.T) {
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
		{NonMongoCommand, args{contents: GetCodeExampleForTesting(NonMongoCommand, Shell), llm: llm, ctx: ctx}, NonMongoCommand},
		{SyntaxExample, args{contents: GetCodeExampleForTesting(SyntaxExample, Shell), llm: llm, ctx: ctx}, SyntaxExample},
		{ExampleConfigurationObject, args{contents: GetCodeExampleForTesting(ExampleConfigurationObject, Shell), llm: llm, ctx: ctx}, ExampleConfigurationObject},
		{ExampleReturnObject, args{contents: GetCodeExampleForTesting(ExampleReturnObject, Shell), llm: llm, ctx: ctx}, ExampleReturnObject},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategorizeShellSnippet(tt.args.contents, tt.args.llm, tt.args.ctx); got != tt.want {
				t.Errorf("CategorizeShellSnippet() = got %v, want %v", got, tt.want)
			}
		})
	}
}
