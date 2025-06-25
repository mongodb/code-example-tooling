package types

// Branch represents each branch object in the JSON
type Branch struct {
	GitBranchName  string `json:"gitBranchName"`
	Active         bool   `json:"active"`
	FullUrl        string `json:"fullUrl"`
	Label          string `json:"label"`
	IsStableBranch bool   `json:"isStableBranch"`
	OfflineUrl     string `json:"offlineUrl,omitempty"` // Use omitempty to handle optional fields
}

// Search represents the search object nested within each data item
type Search struct {
	CategoryTitle string `json:"categoryTitle"`
}

// DocsProject represents each item in the "data" array
type DocsProject struct {
	DisplayName string   `json:"displayName"`
	RepoName    string   `json:"repoName"`
	Project     string   `json:"project"`
	Search      *Search  `json:"search,omitempty"` // Use pointer with omitempty for optional fields
	Branches    []Branch `json:"branches"`
}

// Response represents the top-level JSON structure
type Response struct {
	Data []DocsProject `json:"data"`
}

type ProjectDetails struct {
	ProjectName  string
	ActiveBranch string
	ProdUrl      string
}
