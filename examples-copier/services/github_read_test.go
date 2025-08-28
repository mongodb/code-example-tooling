package services_test

import (
	"encoding/base64"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/require"

	test "github.com/mongodb/code-example-tooling/code-copier/tests"
)

// Helper to b64-encode inline strings
func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

// Convenience to ensure SRC_BRANCH, REPO_OWNER, and REPO_NAME are set in the test environment
func ensureEnv(t *testing.T) (owner, repo string) {
	t.Helper()
	t.Setenv("SRC_BRANCH", "main")
	return test.EnvOwnerRepo(t)
}

// Convenience to register the same responder for both real and placeholder owner/repo
func stubContentsForBothOwners(path, contentB64 string, owner, repo string) {
	test.MockContentsEndpoint(owner, repo, path, contentB64)
	test.MockContentsEndpoint("REPO_OWNER", "REPO_NAME", path, contentB64)
}

func TestRetrieveAndParseConfigFile_Valid(t *testing.T) {
	_ = test.WithHTTPMock(t)
	owner, repo := ensureEnv(t)

	cfgJSON := `
[
  {
    "source_directory": "examples",
    "target_repo": "TargetRepoA",
    "target_branch": "main",
    "target_directory": "dest",
    "recursive_copy": true
  },
  {
    "source_directory": "v2/examples",
    "target_repo": "TargetRepoB",
    "target_branch": "release/2.0",
    "target_directory": "v2/dest",
    "recursive_copy": false
  }
]`
	// Reads configs.ConfigFile via the Contents API (base64-encoded payload)
	stubContentsForBothOwners(configs.ConfigFile, b64(cfgJSON), owner, repo)

	got, err := services.RetrieveAndParseConfigFile()
	require.NoError(t, err, "expected valid JSON to parse without error")
	require.Len(t, got, 2)

	// Spot-check some fields to ensure they're parsed correctly
	require.Equal(t, "examples", got[0].SourceDirectory)
	require.Equal(t, "TargetRepoA", got[0].TargetRepo)
	require.Equal(t, "main", got[0].TargetBranch)
	require.Equal(t, "dest", got[0].TargetDirectory)
	require.True(t, got[0].RecursiveCopy)

	require.Equal(t, "v2/examples", got[1].SourceDirectory)
	require.Equal(t, "TargetRepoB", got[1].TargetRepo)
	require.Equal(t, "release/2.0", got[1].TargetBranch)
	require.Equal(t, "v2/dest", got[1].TargetDirectory)
	require.False(t, got[1].RecursiveCopy)
}

func TestRetrieveAndParseConfigFile_InvalidJSON(t *testing.T) {
	_ = test.WithHTTPMock(t)
	owner, repo := ensureEnv(t)

	// Malformed JSON should trigger an error path in RetrieveAndParseConfigFile
	invalid := `{ "source_directory": "examples", ` // <-- truncated / invalid json

	stubContentsForBothOwners(configs.ConfigFile, b64(invalid), owner, repo)

	got, err := services.RetrieveAndParseConfigFile()
	require.Error(t, err, "invalid JSON must return an error")
	require.Nil(t, got)
}

func TestRetrieveFileContents_Success(t *testing.T) {
	_ = test.WithHTTPMock(t)
	owner, repo := ensureEnv(t)

	path := "examples/a.txt"
	payload := "hello"
	stubContentsForBothOwners(path, b64(payload), owner, repo)

	rc, err := services.RetrieveFileContents(path)
	require.NoError(t, err, "expected RetrieveFileContents to succeed")
	require.IsType(t, github.RepositoryContent{}, rc)
	require.Equal(t, path, rc.GetPath())
	require.NotNil(t, rc.Content)
	require.Contains(t, *rc.Content, b64(payload))
}

// Test that Retrieve and Parse round-trips with one entry
func TestRetrieveAndParseConfigFile_RoundTripMinimal(t *testing.T) {
	_ = test.WithHTTPMock(t)
	owner, repo := ensureEnv(t)

	min := types.ConfigFileType{
		{
			SourceDirectory: "examples",
			TargetRepo:      "TargetRepo",
			TargetBranch:    "main",
			TargetDirectory: "dest",
			RecursiveCopy:   true,
		},
	}
	// Note: using literals to avoid JSON marshaller here
	minJSON := `
[
  {
    "source_directory": "examples",
    "target_repo": "TargetRepo",
    "target_branch": "main",
    "target_directory": "dest",
    "recursive_copy": true
  }
]`
	stubContentsForBothOwners(configs.ConfigFile, b64(minJSON), owner, repo)

	got, err := services.RetrieveAndParseConfigFile()
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, min[0].SourceDirectory, got[0].SourceDirectory)
	require.Equal(t, min[0].TargetRepo, got[0].TargetRepo)
	require.Equal(t, min[0].TargetBranch, got[0].TargetBranch)
	require.Equal(t, min[0].TargetDirectory, got[0].TargetDirectory)
	require.Equal(t, min[0].RecursiveCopy, got[0].RecursiveCopy)
}
