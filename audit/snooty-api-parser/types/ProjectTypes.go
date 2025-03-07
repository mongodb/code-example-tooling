package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// TitleInfo holds the type, position, and value details.
type TitleInfo struct {
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Value    string   `json:"value"`
}

type ProjectMetadataWrapper struct {
	Type string          `json:"type"`
	Data ProjectMetadata `json:"data"`
}

// ProjectMetadata holds the entire data structure from the JSON.
type ProjectMetadata struct {
	ID                    string                 `json:"_id"`
	Project               string                 `json:"project"`
	Branch                string                 `json:"branch"`
	Title                 string                 `json:"title"`
	EOL                   bool                   `json:"eol"`
	Canonical             interface{}            `json:"canonical"` // Assuming nullable field
	SlugToTitle           map[string][]TitleInfo `json:"slugToTitle"`
	SlugToBreadcrumbLabel map[string]string      `json:"slugToBreadcrumbLabel"`
	Toctree               Toctree                `json:"toctree"`
	ToctreeOrder          []string               `json:"toctreeOrder"`
	ParentPaths           map[string][]string    `json:"parentPaths"`
	MultiPageTutorials    map[string]interface{} `json:"multiPageTutorials"` // Needs further clarification
	StaticFiles           map[string]string      `json:"static_files"`
	GithubUsername        string                 `json:"github_username"`
	BuildID               string                 `json:"build_id"`
	CreatedAt             string                 `json:"created_at"` // You might consider using time.Time
}

// Toctree represents the ToC tree structure.
type Toctree struct {
	Title    []TitleInfo   `json:"title"`
	Slug     string        `json:"slug"`
	Children []Toctree     `json:"children"`
	Options  ToctreeOption `json:"options"`
	URL      string        `json:"url,omitempty"` // Optional, based on JSON sample
}

// ToctreeOption holds the options for a Toctree item.
type ToctreeOption struct {
	Drawer bool `json:"drawer"`
}

type ProjectAsset struct {
	Type string    `json:"type"`
	Data FileAsset `json:"data"`
}

type FileAsset struct {
	Checksum  string   `json:"checksum"`
	AssetData string   `json:"assetData"`
	Filenames []string `json:"filenames"`
}

type TimestampData struct {
	Type string `json:"type"`
	Data int64  `json:"data"`
}

// Implement custom unmarshalling for TimestampData
func (t *TimestampData) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	t.Type = raw["type"].(string)
	// Handle the case where Data might be a string or float64
	switch data := raw["data"].(type) {
	case float64:
		t.Data = int64(data)
	case string:
		parsedData, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			return err
		}
		t.Data = parsedData
	default:
		return fmt.Errorf("unexpected type for data: %T", data)
	}
	return nil
}

type ProjectCounts struct {
	IncomingCodeNodesCount      int
	IncomingLiteralIncludeCount int
	IncomingIoCodeBlockCount    int
	RemovedCodeNodesCount       int
	UpdatedCodeNodesCount       int
	NewCodeNodesCount           int
	ExistingCodeNodesCount      int
	ExistingLiteralIncludeCount int
	ExistingIoCodeBlockCount    int
}
