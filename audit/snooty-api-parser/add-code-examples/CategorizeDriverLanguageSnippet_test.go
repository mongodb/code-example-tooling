package add_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
	"testing"
)

func TestCategorizeDriverLanguageSnippet(t *testing.T) {
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
		// This function only returns two possible values - UsageExample or SyntaxExample
		{UsageExample, args{contents: GetCodeExampleForTesting(UsageExample, DriversMinusJs), llm: llm, ctx: ctx}, UsageExample},
		{SyntaxExample, args{contents: GetCodeExampleForTesting(SyntaxExample, DriversMinusJs), llm: llm, ctx: ctx}, SyntaxExample},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategorizeDriverLanguageSnippet(tt.args.contents, tt.args.llm, tt.args.ctx); got != tt.want {
				t.Errorf("CategorizeDriverLanguageSnippet() = got %v, want %v", got, tt.want)
			}
		})
	}
}
