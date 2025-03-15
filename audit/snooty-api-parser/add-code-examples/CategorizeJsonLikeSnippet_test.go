package add_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
	"testing"
)

func TestCategorizeJsonLikeSnippet(t *testing.T) {
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
		// This function only returns two possible values - ExampleConfigurationObject or ExampleReturnObject
		{ExampleConfigurationObject, args{contents: GetCodeExampleForTesting(ExampleConfigurationObject, JsonLike), llm: llm, ctx: ctx}, ExampleConfigurationObject},
		{ExampleReturnObject, args{contents: GetCodeExampleForTesting(ExampleReturnObject, JsonLike), llm: llm, ctx: ctx}, ExampleReturnObject},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategorizeJsonLikeSnippet(tt.args.contents, tt.args.llm, tt.args.ctx); got != tt.want {
				t.Errorf("CategorizeJsonLikeSnippet() = got %v, want %v", got, tt.want)
			}
		})
	}
}
