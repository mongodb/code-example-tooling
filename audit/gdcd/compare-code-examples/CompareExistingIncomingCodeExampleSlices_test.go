package compare_code_examples

import (
	"common"
	"context"
	"fmt"
	"gdcd/add-code-examples"
	"gdcd/compare-code-examples/data"
	"gdcd/types"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
	"testing"
)

func TestOneNewCodeExampleHandledCorrectly(t *testing.T) {
	existingNodes := []common.CodeNode{}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := data.GetNewASTNodes(1)
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 1 {
		t.Errorf("FAILED: Got %d, want 1 code nodes", len(nodes))
	}
	if updatedReport.Counter.NewCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 new code node", len(nodes))
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestOneUnchangedCodeExampleHandledCorrectly(t *testing.T) {
	existingNode, incomingNode := data.GetUnchangedNodes()
	existingNodes := []common.CodeNode{existingNode}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := []types.ASTNode{incomingNode}
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 1 {
		t.Errorf("FAILED: Got %d, want 1 code node", len(nodes))
	}
	if updatedReport.Counter.UnchangedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 unchanged code node", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestOneUpdatedCodeExampleHandledCorrectly(t *testing.T) {
	existingNode, incomingNode := data.GetUpdatedNodes()
	existingNodes := []common.CodeNode{existingNode}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := []types.ASTNode{incomingNode}
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 1 {
		t.Errorf("FAILED: Got %d, want 1 code node", len(nodes))
	}
	if updatedReport.Counter.UpdatedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 updated code node", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestOneRemovedCodeExampleHandledCorrectly(t *testing.T) {
	existingNode, _ := data.GetRemovedNodes()
	existingNodes := []common.CodeNode{existingNode}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := []types.ASTNode{}
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 1 {
		t.Errorf("FAILED: Got %d, want 1 code node", len(nodes))
	}
	if !nodes[0].IsRemoved {
		t.Errorf("FAILED: The code node should show as removed")
	}
	if updatedReport.Counter.RemovedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 removed code node", updatedReport.Counter.RemovedCodeNodesCount)
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestOneNewOneUpdatedCodeExampleHandledCorrectly(t *testing.T) {
	newNode := data.GetNewASTNodes(1)
	existingNode, updatedASTNode := data.GetUpdatedNodes()
	existingNodes := []common.CodeNode{existingNode}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := append(newNode, updatedASTNode)
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 2 {
		t.Errorf("FAILED: Got %d, want 2 code nodes", len(nodes))
	}
	hasWrongReportCount := false
	if updatedReport.Counter.UpdatedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 updated code node", updatedReport.Counter.UpdatedCodeNodesCount)
		hasWrongReportCount = true
	}
	if updatedReport.Counter.NewCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 new code node", updatedReport.Counter.NewCodeNodesCount)
		hasWrongReportCount = true
	}
	if hasWrongReportCount {
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestOneNewOneUnchangedCodeExampleHandledCorrectly(t *testing.T) {
	newNode := data.GetNewASTNodes(1)
	existingNode, incomingNode := data.GetUnchangedNodes()
	existingNodes := []common.CodeNode{existingNode}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := append(newNode, incomingNode)
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 2 {
		t.Errorf("FAILED: Got %d, want 2 code nodes", len(nodes))
	}
	hasWrongReportCount := false
	if updatedReport.Counter.UnchangedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, Report should show 1 unchanged code node", updatedReport.Counter.UnchangedCodeNodesCount)
		hasWrongReportCount = true
	}
	if updatedReport.Counter.NewCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 new code node", updatedReport.Counter.NewCodeNodesCount)
		hasWrongReportCount = true
	}
	if hasWrongReportCount {
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestOneNewOneRemovedCodeExampleHandledCorrectly(t *testing.T) {
	existingNode, _ := data.GetRemovedNodes()
	existingNodes := []common.CodeNode{existingNode}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := data.GetNewASTNodes(1)
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 2 {
		t.Errorf("FAILED: Got %d, want 2 code nodes", len(nodes))
	}
	hasWrongReportCount := false
	if updatedReport.Counter.RemovedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 removed code node", updatedReport.Counter.RemovedCodeNodesCount)
		hasWrongReportCount = true
	}
	if updatedReport.Counter.NewCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 new code node", updatedReport.Counter.NewCodeNodesCount)
		hasWrongReportCount = true
	}
	if hasWrongReportCount {
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestOneUpdatedOneUnchangedCodeExampleHandledCorrectly(t *testing.T) {
	existingNode1, incomingUnchangedNode := data.GetUnchangedNodes()
	existingNode2, incomingUpdatedNode := data.GetUpdatedNodes()
	existingNodes := []common.CodeNode{existingNode1, existingNode2}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := []types.ASTNode{incomingUnchangedNode, incomingUpdatedNode}
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)
	if len(nodes) != 2 {
		t.Errorf("FAILED: Got %d, want 2 code nodes", len(nodes))
	}
	hasWrongReportCount := false
	if updatedReport.Counter.UpdatedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 updated code node", updatedReport.Counter.UpdatedCodeNodesCount)
		hasWrongReportCount = true
	}
	if updatedReport.Counter.UnchangedCodeNodesCount != 1 {
		t.Errorf("FAILED: Got %d, report should show 1 unchanged code node", updatedReport.Counter.UnchangedCodeNodesCount)
		hasWrongReportCount = true
	}
	if hasWrongReportCount {
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestDuplicateUpdatedCodeExampleHandledCorrectly(t *testing.T) {
	existingNodes := []common.CodeNode{}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := []types.ASTNode{}
	existingNode, updatedASTNode := data.GetUpdatedNodes()
	existingNodes = append(existingNodes, existingNode)
	// Appending the same node twice to represent it as a duplicate on the page
	incomingNodes = append(incomingNodes, updatedASTNode)
	incomingNodes = append(incomingNodes, updatedASTNode)
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)

	// For this test, we only want to store the code node once in the array. But we want `InstancesOnPage` to show
	// that it is on the page twice. And we want the report to count it as two code examples.
	if len(nodes) != 1 {
		t.Errorf("FAILED: Got %d, want 1 code nodes", len(nodes))
	}
	if nodes[0].InstancesOnPage != 2 {
		t.Errorf("FAILED: Got %d, want 2 instances on page", nodes[0].InstancesOnPage)
	}
	if updatedReport.Counter.UpdatedCodeNodesCount != 2 {
		t.Errorf("FAILED: Got %d, report should show 2 updated code nodes", len(nodes))
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestDuplicateUnchangedCodeExampleHandledCorrectly(t *testing.T) {
	existingNodes := []common.CodeNode{}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := []types.ASTNode{}
	existingNode, incomingUnchangedNode := data.GetUnchangedNodes()
	existingNodes = append(existingNodes, existingNode)
	// Appending the same node twice to represent it as a duplicate on the page
	incomingNodes = append(incomingNodes, incomingUnchangedNode)
	incomingNodes = append(incomingNodes, incomingUnchangedNode)
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)

	// For this test, we only want to store the code node once in the array. But we want `InstancesOnPage` to show
	// that it is on the page twice. And we want the report to count it as two code examples.
	if len(nodes) != 1 {
		t.Errorf("FAILED: Got %d, want 1 code nodes", len(nodes))
	}
	if nodes[0].InstancesOnPage != 2 {
		t.Errorf("FAILED: Got %d, want 2 instances on page", nodes[0].InstancesOnPage)
	}
	if updatedReport.Counter.UnchangedCodeNodesCount != 2 {
		t.Errorf("FAILED: Got %d, report should show 2 unchanged code nodes", len(nodes))
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}

func TestDuplicateNewCodeExampleHandledCorrectly(t *testing.T) {
	existingNodes := []common.CodeNode{}
	existingRemovedNodes := []common.CodeNode{}
	incomingNodes := data.GetNewASTNodes(1)
	incomingNodes = append(incomingNodes, incomingNodes...)
	initialReport := types.ProjectReport{
		ProjectName: "test-project",
		Changes:     nil,
		Issues:      nil,
		Counter:     types.ProjectCounts{},
	}
	pageId := "some/page/url"
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	nodes, updatedReport := CompareExistingIncomingCodeExampleSlices(existingNodes, existingRemovedNodes, incomingNodes, initialReport, pageId, llm, ctx, true)

	// For this test, we only want to store the new code node once in the array. But we want `InstancesOnPage` to show
	// that it is on the page twice. And we want the report to count it as two code examples.
	if len(nodes) != 1 {
		t.Errorf("FAILED: Got %d, want 1 code nodes", len(nodes))
	}
	if nodes[0].InstancesOnPage != 2 {
		t.Errorf("FAILED: Got %d, want 2 instances on page", nodes[0].InstancesOnPage)
	}
	if updatedReport.Counter.NewCodeNodesCount != 2 {
		t.Errorf("FAILED: Got %d, report should show 2 new code nodes", len(nodes))
		fmt.Println("New code node count: ", updatedReport.Counter.NewCodeNodesCount)
		fmt.Println("Updated code node count: ", updatedReport.Counter.UpdatedCodeNodesCount)
		fmt.Println("Unchanged code node count: ", updatedReport.Counter.UnchangedCodeNodesCount)
		fmt.Println("Removed code node count: ", updatedReport.Counter.RemovedCodeNodesCount)
	}
}
