package add_code_examples

import (
	"common"
	"context"
	"log"
	"testing"

	"github.com/tmc/langchaingo/llms/ollama"
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
		{common.UsageExample, args{contents: GetCodeExampleForTesting(common.UsageExample, DriversMinusJs), llm: llm, ctx: ctx}, common.UsageExample},
		{common.SyntaxExample, args{contents: GetCodeExampleForTesting(common.SyntaxExample, DriversMinusJs), llm: llm, ctx: ctx}, common.SyntaxExample},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := CategorizeDriverLanguageSnippet(tt.args.contents, tt.args.llm, tt.args.ctx); got != tt.want {
				t.Errorf("CategorizeDriverLanguageSnippet() = got %v, want %v", got, tt.want)
			}
		})
	}
}
