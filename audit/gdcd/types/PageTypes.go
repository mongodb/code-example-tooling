package types

import (
	"encoding/json"
	"errors"
)

type PageWrapper struct {
	Type string       `json:"type"`
	Data PageMetadata `json:"data"`
}

// Position represents the line position in the text.
type Position struct {
	Start PositionLine `json:"start"`
}

// PositionLine holds the line number in the position object.
type PositionLine struct {
	Line int `json:"line"`
}

// TextNode describes a textual element within the AST.
type TextNode struct {
	Type     string    `json:"type"`
	Position Position  `json:"position"`
	Value    string    `json:"value,omitempty"`
	Children []ASTNode `json:"children,omitempty"`
}

// AST represents the root of the abstract syntax tree for the page.
type AST struct {
	Type     string      `json:"type"`
	Position Position    `json:"position"`
	Children []ASTNode   `json:"children"`
	FileID   string      `json:"fileid"`
	Options  PageOptions `json:"options"`
}

// ASTNode accommodates various node types within the AST such as directives, sections, etc.
type ASTNode struct {
	Type           string                 `json:"type"`
	Position       Position               `json:"position"`
	Children       []ASTNode              `json:"children,omitempty"`
	Value          string                 `json:"value,omitempty"`
	Lang           string                 `json:"lang,omitempty"`
	Copyable       bool                   `json:"copyable,omitempty"`
	Entries        []ToctreeEntry         `json:"entries,omitempty"`
	EnumType       string                 `json:"enumtype,omitempty"`
	ID             string                 `json:"id,omitempty"`
	Domain         string                 `json:"domain,omitempty"`
	Name           string                 `json:"name,omitempty"`
	Argument       []TextNode             `json:"argument,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
	EmphasizeLines EmphasizeLines         `json:"emphasize_lines,omitempty"`
	LineNumbers    bool                   `json:"lineos,omitempty"`
}

// ToctreeEntry details entries contained within a toctree.
type ToctreeEntry struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// PageOptions holds various configuration settings for the page.
type PageOptions struct {
	Headings []Heading `json:"headings"`
}

// Heading encapsulates heading information, including depth and title.
type Heading struct {
	Depth int        `json:"depth"`
	ID    string     `json:"id"`
	Title []TextNode `json:"title"`
}

// PageMetadata conveys comprehensive metadata and content information about the page.
type PageMetadata struct {
	ID             string  `json:"_id"`
	GitHubUsername string  `json:"github_username,omitempty"`
	PageID         string  `json:"page_id"`
	AST            AST     `json:"ast"`
	BuildID        string  `json:"build_id"`
	CreatedAt      string  `json:"created_at"`
	Deleted        bool    `json:"deleted"`
	Filename       string  `json:"filename"`
	StaticAssets   []Asset `json:"static_assets"`
	UpdatedAt      string  `json:"updated_at"`
}

// EmphasizeLines custom type to hold explicit line numbers
type EmphasizeLines []int

// UnmarshalJSON implements custom unmarshalling logic for EmphasizeLines
func (e *EmphasizeLines) UnmarshalJSON(b []byte) error {
	var data [][]int
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	var lines []int
	for _, pair := range data {
		if len(pair) != 2 {
			return errors.New("each sub-array should contain exactly two elements representing a range")
		}
		for i := pair[0]; i <= pair[1]; i++ { // expand range
			lines = append(lines, i)
		}
	}
	*e = lines
	return nil
}

// Define the Asset struct that represents each static asset
type Asset struct {
	Checksum  string `json:"checksum"`
	Key       string `json:"key"`
	UpdatedAt string `json:"updated_at"`
}
