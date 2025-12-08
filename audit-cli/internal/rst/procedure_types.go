package rst

import "regexp"

// ProcedureType represents the type of procedure implementation.
type ProcedureType string

const (
	// ProcedureDirective represents procedures using .. procedure:: directive
	ProcedureDirective ProcedureType = "procedure-directive"
	// OrderedList represents procedures using ordered lists
	OrderedList ProcedureType = "ordered-list"
)

// Procedure represents a parsed procedure from an RST file.
type Procedure struct {
	Type               ProcedureType       // Type of procedure (directive or ordered list)
	Title              string              // Title/heading above the procedure
	Options            map[string]string   // Directive options (for procedure directive)
	Steps              []Step              // Steps in the procedure
	LineNum            int                 // Line number where procedure starts (1-based)
	EndLineNum         int                 // Line number where procedure ends (1-based)
	HasSubSteps        bool                // Whether this procedure contains sub-procedures
	IsSubProcedure     bool                // Whether this is a sub-procedure within a step
	ComposableTutorial *ComposableTutorial // Composable tutorial wrapping this procedure (if any)
	TabSet             *TabSetInfo         // Tab set wrapping this procedure (if any)
	TabID              string              // The specific tab ID this procedure belongs to (if part of a tab set)
}

// TabSetInfo represents information about a tab set containing procedure variations.
// This is used for grouping procedures for analysis/reporting purposes.
type TabSetInfo struct {
	TabIDs     []string             // All tab IDs in the set (for grouping)
	Procedures map[string]Procedure // All procedures by tabid (for grouping)
}

// Step represents a single step in a procedure.
type Step struct {
	Title         string            // Step title (for .. step:: directive)
	Content       string            // Step content (raw RST)
	Options       map[string]string // Step options
	LineNum       int               // Line number where step starts
	Variations    []Variation       // Variations within this step (tabs or selected content)
	SubSteps      []Step            // DEPRECATED: Use SubProcedures instead. Kept for backward compatibility.
	SubProcedures []SubProcedure    // Multiple sub-procedures (each is an ordered list within this step)
}

// SubProcedure represents an ordered list within a step
type SubProcedure struct {
	Steps    []Step // The steps in this sub-procedure
	ListType string // "numbered" or "lettered" - the type of ordered list marker used
}

// Variation represents a content variation within a step.
type Variation struct {
	Type    VariationType     // Type of variation (tab or selected-content)
	Options []string          // Available options (tabids or selections)
	Content map[string]string // Content for each option
}

// VariationType represents the type of content variation.
type VariationType string

const (
	// TabVariation represents variations using .. tabs:: directive
	TabVariation VariationType = "tabs"
	// SelectedContentVariation represents variations using .. selected-content:: directive
	SelectedContentVariation VariationType = "selected-content"
)

// ComposableTutorial represents a composable tutorial structure.
type ComposableTutorial struct {
	Title                 string            // Title/heading above the composable tutorial
	Options               []string          // Available option names (e.g., ["interface", "language"])
	Defaults              []string          // Default selections (e.g., ["driver", "nodejs"])
	Selections            []string          // All unique selection combinations found
	GeneralContent        []string          // Content lines that apply to all selections
	LineNum               int               // Line number where tutorial starts
	FilePath              string            // Path to the source file (for resolving includes)
	Procedure             *Procedure        // The procedure within the composable tutorial
	SelectedContentBlocks []SelectedContent // All selected-content blocks (for extracting multiple procedures)
}

// TabSet represents a tabs directive containing procedures.
type TabSet struct {
	Title      string               // Title/heading above the tabs
	Tabs       map[string][]string  // Tab content by tabid (lines of RST)
	TabIDs     []string             // Ordered list of tab IDs
	Procedures map[string]Procedure // Parsed procedures by tabid
	LineNum    int                  // Line number where tabs start
	FilePath   string               // Path to the source file (for resolving includes)
}

// SelectedContent represents a selected-content block within a composable tutorial.
type SelectedContent struct {
	Selections []string // The selections for this content (e.g., ["driver", "nodejs"])
	Content    string   // The content for this selection
	LineNum    int      // Line number where this selected-content starts
}

// Regular expressions for parsing ordered lists
var (
	// Matches numbered lists: 1. or 1)
	numberedListRegex = regexp.MustCompile(`^(\s*)(\d+)[\.\)]\s+(.*)$`)
	// Matches lettered lists: a. or a) or A. or A)
	letteredListRegex = regexp.MustCompile(`^(\s*)([a-zA-Z])[\.\)]\s+(.*)$`)
	// Matches continuation marker: #. (used to continue an ordered list)
	continuationMarkerRegex = regexp.MustCompile(`^(\s*)#[\.\)]\s+(.*)$`)
)

// YAMLStep represents a step in a YAML steps file
type YAMLStep struct {
	Title   string      `yaml:"title"`
	StepNum int         `yaml:"stepnum"`
	Level   int         `yaml:"level"`
	Ref     string      `yaml:"ref"`
	Pre     string      `yaml:"pre"`
	Action  interface{} `yaml:"action"`
	Post    string      `yaml:"post"`
}
