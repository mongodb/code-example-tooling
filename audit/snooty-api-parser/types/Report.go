package types

type ProjectCounts struct {
	NewPagesCount               int
	IncomingCodeNodesCount      int
	IncomingLiteralIncludeCount int
	IncomingIoCodeBlockCount    int
	RemovedCodeNodesCount       int
	UpdatedCodeNodesCount       int
	UnchangedCodeNodesCount     int
	NewCodeNodesCount           int
	ExistingCodeNodesCount      int
	ExistingLiteralIncludeCount int
	ExistingIoCodeBlockCount    int
	RemovedPagesCount           int
	TotalCurrentPageCount       int
}

// ChangeType represents the type of change.
type ChangeType int
type IssueType int

const (
	// Define the possible types of changes.
	PageCreated ChangeType = iota
	PageUpdated
	PageRemoved
	CodeExampleCreated
	CodeExampleUpdated
	CodeExampleRemoved
	CodeNodeCountChange
	LiteralIncludeCountChange
	IoCodeBlockCountChange
	ProjectSummaryCodeNodeCountChange
	ProjectSummaryPageCountChange
	AppliedUsageExampleAdded
)

const (
	// Define the possible types of issues
	PagesNotFoundIssue IssueType = iota
	CodeNodeCountIssue
	PageCountIssue
)

// Change represents a change happening to data.
type Change struct {
	Type ChangeType  // The type of change
	Data interface{} // The data associated with the change
}

type Issue struct {
	Type IssueType   // The type of change
	Data interface{} // The data associated with the change
}

// String returns a string representation of the ChangeType for easier readability.
func (ct ChangeType) String() string {
	return [...]string{"Page created", "Page updated", "Page removed", "Code example created", "Code example updated", "Code example removed", "Code node count change", "literalinclude count change", "io-code-block count change", "Project summary node count change", "Project summary page count change", "Applied usage example added"}[ct]
}

// String returns a string representation of the IssueType for easier readability.
func (it IssueType) String() string {
	return [...]string{"Pages not found", "Code node count issue", "Page count issue"}[it]
}

type ProjectReport struct {
	ProjectName string
	Changes     []Change
	Issues      []Issue
	Counter     ProjectCounts
}
