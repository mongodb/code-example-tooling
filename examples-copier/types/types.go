package types

import (
	"github.com/google/go-github/v48/github"
	"github.com/shurcooL/githubv4"
)

// **** PR **** //

type PullRequestQuery struct {
	Repository struct {
		PullRequest struct {
			Files struct {
				Edges []struct {
					Node struct {
						Path       githubv4.String
						Additions  githubv4.Int
						Deletions  githubv4.Int
						ChangeType githubv4.String
					}
				}
			} `graphql:"files(first: 100)"`
		} `graphql:"pullRequest(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type ChangedFile struct {
	Path      string
	Additions int
	Deletions int
	Status    string
}

// **** CHANGED FILES **** //

type RepoFilesResponse struct {
	Repository Repository `graphql:"repository(owner: $owner, name: $name)"`
}
type Repository struct {
	Object *GitObject `graphql:"object(expression: \"HEAD:\")"`
}
type TreeFragment struct {
	Entries []*Entry `json:"entries,omitempty"`
}
type GitObject struct {
	TreeFragment `graphql:"... on Tree"`
}
type Entry struct {
	Name   string        `json:"name"`
	Type   string        `json:"type"`
	Mode   int           `json:"mode"`
	Object *BlobFragment `json:"object"`
}

type BlobFragment struct {
	Blob BlobObject `graphql:"... on Blob"`
}
type BlobObject struct {
	Text string `json:"text"`
}
type ConfigFileType []Configs

// Configs represents the configuration for file copying operations between repositories.
//
// Example usage:
//
//	config := Configs{
//	  SourceDirectory: "docs/api",
//	  TargetRepo: "company/public-docs",
//	  TargetBranch: "main",
//	  TargetDirectory: "reference",
//	  RecursiveCopy: true,
//	  CopierCommitStrategy: "pr",
//	  PRTitle: "Update API documentation",
//	  CommitMessage: "Sync API docs from internal repo",
//	  MergeWithoutReview: false,
//	}
type Configs struct {
	SourceDirectory      string `json:"source_directory"`
	TargetRepo           string `json:"target_repo"`
	TargetBranch         string `json:"target_branch"`
	TargetDirectory      string `json:"target_directory"`
	RecursiveCopy        bool   `json:"recursive_copy"`
	CopierCommitStrategy string `json:"copier_commit_strategy"`
	PRTitle              string `json:"pr_title"`
	CommitMessage        string `json:"commit_message"`
	MergeWithoutReview   bool   `json:"merge_without_review"`
}
type DeprecationFile []DeprecatedFileEntry
type DeprecatedFileEntry struct {
	FileName  string `json:"filename"`
	Repo      string `json:"repo"`
	Branch    string `json:"branch"`
	DeletedOn string `json:"deleted_on"`
}

// **** UPLOAD TYPES **** //

type UploadKey struct {
	RepoName       string `json:"repo_name"`
	BranchPath     string `json:"branch_path"`
	RuleName       string `json:"rule_name"`        // Include rule name to allow multiple rules targeting same repo/branch
	CommitStrategy string `json:"commit_strategy"`  // Include strategy to differentiate direct vs PR
}

type UploadFileContent struct {
	TargetBranch   string                     `json:"target_branch"`
	Content        []github.RepositoryContent `json:"content"`
	CommitStrategy CommitStrategy             `json:"commit_strategy,omitempty"`
	CommitMessage  string                     `json:"commit_message,omitempty"`
	PRTitle        string                     `json:"pr_title,omitempty"`
	PRBody         string                     `json:"pr_body,omitempty"`
	UsePRTemplate  bool                       `json:"use_pr_template,omitempty"`  // If true, fetch and merge PR template from target repo
	AutoMergePR    bool                       `json:"auto_merge_pr,omitempty"`
}

// CommitStrategy represents the strategy for committing changes
type CommitStrategy string

const (
	CommitStrategyDirect CommitStrategy = "direct"
	CommitStrategyPR     CommitStrategy = "pull_request"
)

type CreateFileRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha,omitempty"`
}
