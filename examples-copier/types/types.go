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

// **** END PR **** //
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
type Configs struct {
	SourceDirectory string `json:"source_directory"`
	TargetRepo      string `json:"target_repo"`
	TargetDirectory string `json:"target_directory"`
	TargetBranch    string `json:"target_branch"`
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
	RepoName   string `json:"repo_name"`
	BranchPath string `json:"branch_path"`
}

type UploadFileContent struct {
	TargetBranch string                     `json:"target_branch"`
	Content      []github.RepositoryContent `json:"content"`
}

type CreateFileRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha,omitempty"`
}
