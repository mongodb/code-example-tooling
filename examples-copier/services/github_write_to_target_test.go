package services_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/jarcoal/httpmock"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/require"

	// test helpers (utils.go)
	test "github.com/mongodb/code-example-tooling/code-copier/tests"
)

func TestMain(m *testing.M) {
	// Minimal env so init() and any env readers are happy.
	os.Setenv(configs.RepoOwner, "my-org")
	os.Setenv(configs.RepoName, "target-repo")
	os.Setenv(configs.InstallationId, "12345")
	os.Setenv(configs.AppId, "1166559")
	os.Setenv(configs.AppClientId, "IvTestClientId")
	os.Setenv("SKIP_SECRET_MANAGER", "true")
	os.Setenv("SRC_BRANCH", "main")

	// Provide an RSA private key (both raw and b64) so ConfigurePermissions can parse.
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	os.Setenv("GITHUB_APP_PRIVATE_KEY", string(pemBytes))
	os.Setenv("GITHUB_APP_PRIVATE_KEY_B64", base64.StdEncoding.EncodeToString(pemBytes))

	code := m.Run()

	// Cleanup
	os.Unsetenv(configs.RepoOwner)
	os.Unsetenv(configs.RepoName)
	os.Unsetenv(configs.InstallationId)
	os.Unsetenv(configs.AppId)
	os.Unsetenv(configs.AppClientId)
	os.Unsetenv("SKIP_SECRET_MANAGER")
	os.Unsetenv("SRC_BRANCH")
	os.Unsetenv("GITHUB_APP_PRIVATE_KEY")
	os.Unsetenv("GITHUB_APP_PRIVATE_KEY_B64")

	os.Exit(code)
}

// LEGACY TESTS - These tests are for legacy code that was removed in commit a64726c
// The AddToRepoAndFilesMap and IterateFilesForCopy functions were removed as part of the
// migration to the new pattern-matching system. These tests are commented out but kept for reference.
//
// The new system uses pattern matching rules defined in YAML config files.
// See pattern_matcher_test.go for tests of the new system.

/*
func TestAddToRepoAndFilesMap_NewEntry(t *testing.T) {
	services.FilesToUpload = nil

	name := "example.txt"
	dummyFile := github.RepositoryContent{Name: &name}

	services.AddToRepoAndFilesMap("TargetRepo1", "main", dummyFile)

	require.NotNil(t, services.FilesToUpload, "FilesToUpload map should be initialized")
	key := types.UploadKey{RepoName: "TargetRepo1", BranchPath: "refs/heads/main", RuleName: "", CommitStrategy: ""}
	entry, exists := services.FilesToUpload[key]
	require.True(t, exists, "Entry for TargetRepo1/main should exist")
	require.Equal(t, "main", entry.TargetBranch)
	require.Len(t, entry.Content, 1)
	require.Equal(t, "example.txt", *entry.Content[0].Name)
}

func TestAddToRepoAndFilesMap_AppendEntry(t *testing.T) {
	services.FilesToUpload = make(map[types.UploadKey]types.UploadFileContent)
	key := types.UploadKey{RepoName: "TargetRepo1", BranchPath: "refs/heads/main", RuleName: "", CommitStrategy: ""}

	initialName := "first.txt"
	services.FilesToUpload[key] = types.UploadFileContent{
		TargetBranch: "main",
		Content:      []github.RepositoryContent{{Name: &initialName}},
	}

	newName := "second.txt"
	newFile := github.RepositoryContent{Name: &newName}
	services.AddToRepoAndFilesMap("TargetRepo1", "main", newFile)

	entry := services.FilesToUpload[key]
	require.Len(t, entry.Content, 2)
	require.ElementsMatch(t, []string{"first.txt", "second.txt"},
		[]string{*entry.Content[0].Name, *entry.Content[1].Name})
}

func TestAddToRepoAndFilesMap_NestedFiles(t *testing.T) {
	services.FilesToUpload = make(map[types.UploadKey]types.UploadFileContent)
	key := types.UploadKey{RepoName: "TargetRepo1", BranchPath: "refs/heads/main", RuleName: "", CommitStrategy: ""}

	initialName := "level1/first.txt"
	services.FilesToUpload[key] = types.UploadFileContent{
		TargetBranch: "main",
		Content:      []github.RepositoryContent{{Name: &initialName}},
	}

	newName := "level1/level2/level3/nested-second.txt"
	newFile := github.RepositoryContent{Name: &newName}
	services.AddToRepoAndFilesMap("TargetRepo1", "main", newFile)

	entry := services.FilesToUpload[key]
	require.Len(t, entry.Content, 2)
	require.ElementsMatch(t, []string{"level1/first.txt", "level1/level2/level3/nested-second.txt"},
		[]string{*entry.Content[0].Name, *entry.Content[1].Name})
}

func TestIterateFilesForCopy_Deletes(t *testing.T) {
	cfg := types.Configs{
		SourceDirectory: "src/examples",
		TargetRepo:      "TargetRepo1",
		TargetBranch:    "main",
		TargetDirectory: "dest/examples",
		RecursiveCopy:   false,
	}
	configFile := types.ConfigFileType{cfg}
	changed := []types.ChangedFile{{
		Path:   "src/examples/sample.txt",
		Status: "DELETED",
	}}

	services.FilesToUpload = nil
	services.FilesToDeprecate = nil

	err := services.IterateFilesForCopy(changed, configFile)
	require.NoError(t, err)

	targetPath := "dest/examples/sample.txt"
	require.Contains(t, services.FilesToDeprecate, targetPath)
	require.Equal(t, cfg, services.FilesToDeprecate[targetPath])
	require.Nil(t, services.FilesToUpload)
}

func TestIterateFilesForCopy_RecursiveVsNonRecursive(t *testing.T) {
	t.Setenv("SRC_BRANCH", "main")
	_ = test.WithHTTPMock(t)

	owner, repo := test.EnvOwnerRepo(t)

	// Simulate changes under the source directory
	changed := []types.ChangedFile{
		test.MakeChanged("ADDED", "examples/a.txt"),
		test.MakeChanged("MODIFIED", "examples/sub/b.txt"),
		test.MakeChanged("ADDED", "examples/sub/deeper/c.txt"),
	}

	// Helper to base64-encode small content blobs
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

	// Register responders for owner/repo
	for _, or := range [][2]string{{owner, repo}, {"REPO_OWNER", "REPO_NAME"}} {
		test.MockContentsEndpoint(or[0], or[1], "examples/a.txt", b64("A"))
		test.MockContentsEndpoint(or[0], or[1], "examples/sub/b.txt", b64("B"))
		test.MockContentsEndpoint(or[0], or[1], "examples/sub/deeper/c.txt", b64("C"))
	}

	// Same source; two configs exercising recursive vs non-recursive and different targets
	cases := []struct {
		name   string
		cfg    types.Configs
		expect []string // expected TARGET paths
	}{
		{
			name: "recursive=true copies all depths",
			cfg: types.Configs{
				SourceDirectory: "examples",
				TargetRepo:      "TargetRepoR",
				TargetBranch:    "main",
				TargetDirectory: "dest",
				RecursiveCopy:   true,
			},
			expect: []string{
				"dest/a.txt",
				"dest/sub/b.txt",
				"dest/sub/deeper/c.txt",
			},
		},
		{
			name: "recursive=false copies only root files",
			cfg: types.Configs{
				SourceDirectory: "examples",
				TargetRepo:      "TargetRepoNR",
				TargetBranch:    "main",
				TargetDirectory: "dest",
				RecursiveCopy:   false,
			},
			expect: []string{
				"dest/a.txt",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.ResetGlobals()
			err := services.IterateFilesForCopy(changed, types.ConfigFileType{tc.cfg})
			require.NoError(t, err)
			// Compares staged entries cfg.SourceDirectory -> cfg.TargetDirectory.
			test.AssertUploadedPathsFromConfig(t, tc.cfg, tc.expect)
		})
	}
}
*/

func TestAddFilesToTargetRepoBranch_Succeeds(t *testing.T) {
	_ = test.WithHTTPMock(t)

	owner, repo := test.EnvOwnerRepo(t)
	branch := "main"

	// Set up cached token for the org to bypass GitHub App auth
	test.SetupOrgToken(owner, "test-token")

	baseRefURL, commitsURL, updateRefURL := test.MockGitHubWriteEndpoints(owner, repo, branch)

	files := []github.RepositoryContent{
		{
			Name:    github.String("dir/example1.txt"),
			Path:    github.String("dir/example1.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 1"))),
		},
		{
			Name:    github.String("dir/example2.txt"),
			Path:    github.String("dir/example2.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 2"))),
		},
	}
	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{
		{RepoName: repo, BranchPath: "refs/heads/" + branch}: {
			TargetBranch: branch,
			Content:      files,
		},
	}

	services.AddFilesToTargetRepoBranch()

	info := httpmock.GetCallCountInfo()
	require.Equal(t, 1, info["GET "+baseRefURL])

	// POST /git/trees is registered via regex; sum by prefix
	treeCalls := 0
	for k, v := range info {
		if strings.HasPrefix(k, "POST https://api.github.com/repos/"+owner+"/"+repo+"/git/trees") {
			treeCalls += v
		}
	}
	require.Equal(t, 1, treeCalls)

	require.Equal(t, 1, info["POST "+commitsURL])
	require.Equal(t, 1, info["PATCH "+updateRefURL])

	services.FilesToUpload = nil
}

func TestAddFilesToTargetRepoBranch_ViaPR_Succeeds(t *testing.T) {
	_ = test.WithHTTPMock(t)
	t.Setenv("COPIER_COMMIT_STRATEGY", "pr")

	owner, repo := test.EnvOwnerRepo(t)
	baseBranch := "main"

	// Force fresh token; stub token endpoint then configure permissions.
	services.InstallationAccessToken = ""
	test.MockGitHubAppTokenEndpoint(os.Getenv(configs.InstallationId))
	services.ConfigurePermissions()

	// Set up cached token for the org to bypass GitHub App auth
	test.SetupOrgToken(owner, "test-token")

	// Base ref used to create temp branch
	httpmock.RegisterRegexpResponder("GET",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/ref/(?:refs/)?heads/`+baseBranch+`$`),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref": "refs/heads/" + baseBranch, "object": map[string]any{"sha": "baseSha"},
		}),
	)

	// Create temp branch
	createRefURL := test.MockCreateRef(owner, repo)

	// Temp branch: GET ref, POST tree, POST commit, PATCH ref
	tempHead := `copier/\d{8}-\d{6}`
	httpmock.RegisterRegexpResponder("GET",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/ref/(?:refs/)?heads/`+tempHead+`$`),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref": "refs/heads/copier/20250101-000000", "object": map[string]any{"sha": "baseSha"},
		}),
	)
	httpmock.RegisterRegexpResponder("POST",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/trees(\?.*)?$`),
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newTreeSha"}),
	)
	commitsURL := "https://api.github.com/repos/" + owner + "/" + repo + "/git/commits"
	httpmock.RegisterResponder("POST", commitsURL,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newCommitSha"}),
	)
	httpmock.RegisterRegexpResponder("PATCH",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/refs/heads/`+tempHead+`$`),
		httpmock.NewStringResponder(200, "{}"),
	)

	// PR create + merge; delete temp branch
	test.MockPullsAndMerge(owner, repo, 42)
	test.MockDeleteTempRef(owner, repo)

	// Stage files to baseBranch; service will write via temp branch → PR merge
	files := []github.RepositoryContent{
		{
			Name:    github.String("dir/example1.txt"),
			Path:    github.String("dir/example1.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 1"))),
		},
		{
			Name:    github.String("dir/example2.txt"),
			Path:    github.String("dir/example2.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 2"))),
		},
	}
	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{
		{RepoName: repo, BranchPath: "refs/heads/" + baseBranch}: {
			TargetBranch:   baseBranch,
			Content:        files,
			CommitStrategy: "pr",
			AutoMergePR:    true,
		},
	}

	services.AddFilesToTargetRepoBranch()

	// Assertions
	require.Equal(t, 1, test.CountByMethodAndURLRegexp("POST",
		regexp.MustCompile(`/app/installations/`+regexp.QuoteMeta(os.Getenv(configs.InstallationId))+`/access_tokens$`),
	))
	info := httpmock.GetCallCountInfo()
	require.Equal(t, 1, info["POST "+createRefURL])

	require.Equal(t, 1, test.CountByMethodAndURLRegexp("POST",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/pulls$`),
	))
	require.Equal(t, 1, test.CountByMethodAndURLRegexp("PUT",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/pulls/42/merge$`),
	))
	require.Equal(t, 1, info["POST "+commitsURL])

	require.GreaterOrEqual(t,
		test.CountByMethodAndURLRegexp("GET",
			regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/ref/(?:refs/)?heads/`+regexp.QuoteMeta(baseBranch)+`$`)),
		1,
	)
	require.GreaterOrEqual(t,
		test.CountByMethodAndURLRegexp("GET",
			regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/ref/(?:refs/)?heads/copier/\d{8}-\d{6}$`)),
		1,
	)
	require.GreaterOrEqual(t,
		test.CountByMethodAndURLRegexp("POST",
			regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/trees`)),
		1,
	)
	require.GreaterOrEqual(t,
		test.CountByMethodAndURLRegexp("PATCH",
			regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/refs/heads/copier/\d{8}-\d{6}$`)),
		1,
	)
	require.GreaterOrEqual(t,
		test.CountByMethodAndURLRegexp("DELETE",
			regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/refs/heads/copier/\d{8}-\d{6}$`)),
		1,
	)

	services.FilesToUpload = nil
}



// --- Added critical tests for merge conflicts and configuration/default priorities ---

func TestAddFiles_DirectConflict_NonFastForward(t *testing.T) {
	_ = test.WithHTTPMock(t)

	owner, repo := test.EnvOwnerRepo(t)
	branch := "main"

	// Set up cached token for the org to bypass GitHub App auth
	test.SetupOrgToken(owner, "test-token")

	// Mock standard direct write endpoints
	baseRefURL, commitsURL, updateRefURL := test.MockGitHubWriteEndpoints(owner, repo, branch)

	// Override UpdateRef to simulate 422 Unprocessable Entity (non-fast-forward)
	httpmock.RegisterResponder("PATCH", updateRefURL, httpmock.NewJsonResponderOrPanic(422, map[string]any{
		"message": "Update is not a fast forward",
	}))

	files := []github.RepositoryContent{
		{
			Name:    github.String("dir/example1.txt"),
			Path:    github.String("dir/example1.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 1"))),
		},
	}
	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{
		{RepoName: repo, BranchPath: "refs/heads/" + branch}: {
			TargetBranch: branch,
			Content:      files,
		},
	}

	// Run – should not panic; error is handled/logged internally.
	services.AddFilesToTargetRepoBranch()

	info := httpmock.GetCallCountInfo()
	require.Equal(t, 1, info["GET "+baseRefURL])
	require.Equal(t, 1, info["POST "+commitsURL])
	require.Equal(t, 1, info["PATCH "+updateRefURL])

	services.FilesToUpload = nil
}

func TestAddFiles_ViaPR_MergeConflict_Dirty_NotMerged(t *testing.T) {
	_ = test.WithHTTPMock(t)
	t.Setenv("COPIER_COMMIT_STRATEGY", "pr")

	owner, repo := test.EnvOwnerRepo(t)
	baseBranch := "main"

	// Fresh token path
	services.InstallationAccessToken = ""
	test.MockGitHubAppTokenEndpoint(os.Getenv(configs.InstallationId))
	services.ConfigurePermissions()

	// Set up cached token for the org to bypass GitHub App auth
	test.SetupOrgToken(owner, "test-token")

	// Base ref for creating temp branch
	httpmock.RegisterRegexpResponder("GET",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/ref/(?:refs/)?heads/`+baseBranch+`$`),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref": "refs/heads/" + baseBranch, "object": map[string]any{"sha": "baseSha"},
		}),
	)
	createRefURL := test.MockCreateRef(owner, repo)

	// Temp branch interactions
	tempHead := `copier/\d{8}-\d{6}`
	httpmock.RegisterRegexpResponder("GET",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/ref/(?:refs/)?heads/`+tempHead+`$`),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref": "refs/heads/copier/20250101-000000", "object": map[string]any{"sha": "baseSha"},
		}),
	)
	httpmock.RegisterRegexpResponder("POST",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/trees(\?.*)?$`),
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newTreeSha"}),
	)
	commitsURL := "https://api.github.com/repos/" + owner + "/" + repo + "/git/commits"
	httpmock.RegisterResponder("POST", commitsURL,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newCommitSha"}),
	)
	httpmock.RegisterRegexpResponder("PATCH",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/refs/heads/`+tempHead+`$`),
		httpmock.NewStringResponder(200, "{}"),
	)

	// PR create
	pr_number := 77
	httpmock.RegisterResponder("POST",
		"https://api.github.com/repos/"+owner+"/"+repo+"/pulls",
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"number": pr_number, "html_url": "https://github.com/"+owner+"/"+repo+"/pull/77"}),
	)
	// PR mergeability check returns dirty -> not mergeable
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/"+owner+"/"+repo+"/pulls/77",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{"mergeable": false, "mergeable_state": "dirty"}),
	)
	// Note: do NOT register PUT /merge to ensure it isn't called
	// Also do NOT register DELETE for temp ref; conflict path returns early before cleanup

	// Minimal file to write
	files := []github.RepositoryContent{{
		Name:    github.String("f.txt"),
		Path:    github.String("f.txt"),
		Content: github.String(base64.StdEncoding.EncodeToString([]byte("x"))),
	}}
	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{
		{RepoName: repo, BranchPath: "refs/heads/" + baseBranch}: {
			TargetBranch:   baseBranch,
			Content:        files,
			CommitStrategy: "pr",
		},
	}

	services.AddFilesToTargetRepoBranch()

	// Assertions
	info := httpmock.GetCallCountInfo()
	require.Equal(t, 1, info["POST "+createRefURL])
 require.Equal(t, 1, test.CountByMethodAndURLRegexp("POST",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/pulls$`)))
	// No merge call should have been made
 require.Equal(t, 0, test.CountByMethodAndURLRegexp("PUT",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/pulls/77/merge$`)))
	// No delete of temp ref because we returned early
 require.Equal(t, 0, test.CountByMethodAndURLRegexp("DELETE",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/refs/heads/copier/\d{8}-\d{6}$`)))

	services.FilesToUpload = nil
}

func TestPriority_Strategy_ConfigOverridesEnv_And_MessageFallbacks(t *testing.T) {
	_ = test.WithHTTPMock(t)

	owner, repo := test.EnvOwnerRepo(t)
	baseBranch := "main"

	// Env specifies PR, but config will override to direct
	t.Setenv("COPIER_COMMIT_STRATEGY", "pr")

	// Set up cached token for the org to bypass GitHub App auth
	test.SetupOrgToken(owner, "test-token")

	// Mocks for direct flow
	baseRefURL, commitsURL, updateRefURL := test.MockGitHubWriteEndpoints(owner, repo, baseBranch)

	// Intercept POST commit to assert commit message fallback when config empty but env default set
	wantMsg := "Env Default Commit Message"
	t.Setenv(configs.DefaultCommitMessage, wantMsg)

	// Replace commits responder with custom body assertion
	httpmock.RegisterResponder("POST", commitsURL, func(req *http.Request) (*http.Response, error) {
		defer req.Body.Close()
		b, _ := io.ReadAll(req.Body)
		if !strings.Contains(string(b), wantMsg) {
			t.Fatalf("commit body does not contain expected message: %s; body=%s", wantMsg, string(b))
		}
		return httpmock.NewJsonResponse(201, map[string]any{"sha": "newCommitSha"})
	})

	files := []github.RepositoryContent{{
		Name:    github.String("a.txt"),
		Path:    github.String("a.txt"),
		Content: github.String(base64.StdEncoding.EncodeToString([]byte("x"))),
	}}

	cfg := types.Configs{
		TargetRepo:           repo,
		TargetBranch:         baseBranch,
		CopierCommitStrategy: "direct", // overrides env "pr"
		// CommitMessage empty -> use env default
	}

	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{
		{RepoName: repo, BranchPath: "refs/heads/" + baseBranch, CommitStrategy: cfg.CopierCommitStrategy}: {TargetBranch: baseBranch, Content: files},
	}

	services.AddFilesToTargetRepoBranch() // No longer takes parameters - uses FilesToUpload map

	info := httpmock.GetCallCountInfo()
	require.Equal(t, 1, info["GET "+baseRefURL])
	require.Equal(t, 1, info["POST "+commitsURL])
	require.Equal(t, 1, info["PATCH "+updateRefURL])
	// No PR endpoints should be called
	require.Equal(t, 0, test.CountByMethodAndURLRegexp("POST", regexp.MustCompile(`/pulls$`)))

	services.FilesToUpload = nil
}

func TestPriority_PRTitleDefaultsToCommitMessage_And_NoAutoMergeWhenConfigPresent(t *testing.T) {
	_ = test.WithHTTPMock(t)
	t.Setenv("COPIER_COMMIT_STRATEGY", "pr")

	owner, repo := test.EnvOwnerRepo(t)
	baseBranch := "main"

	// Token setup
	services.InstallationAccessToken = ""
	test.MockGitHubAppTokenEndpoint(os.Getenv(configs.InstallationId))
	services.ConfigurePermissions()

	// Set up cached token for the org to bypass GitHub App auth
	test.SetupOrgToken(owner, "test-token")

	// Base ref and temp branch setup
	httpmock.RegisterRegexpResponder("GET",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/ref/(?:refs/)?heads/`+baseBranch+`$`),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{"ref": "refs/heads/" + baseBranch, "object": map[string]any{"sha": "baseSha"}}),
	)
	_ = test.MockCreateRef(owner, repo)
		tempHead := `copier/\d{8}-\d{6}`
	httpmock.RegisterRegexpResponder("GET",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/ref/(?:refs/)?heads/`+tempHead+`$`),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{"ref": "refs/heads/copier/20250101-000000", "object": map[string]any{"sha": "baseSha"}}),
	)
	httpmock.RegisterRegexpResponder("POST",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/trees(\?.*)?$`),
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "t"}),
	)
	commitsURL := "https://api.github.com/repos/" + owner + "/" + repo + "/git/commits"
	want := "Env Fallback Message"
	t.Setenv(configs.DefaultCommitMessage, want)
	httpmock.RegisterResponder("POST", commitsURL, func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		if !strings.Contains(string(b), want) {
			t.Fatalf("expected commit message %q, got body=%s", want, string(b))
		}
		return httpmock.NewJsonResponse(201, map[string]any{"sha": "c"})
	})
	httpmock.RegisterRegexpResponder("PATCH",
		regexp.MustCompile(`^https://api\.github\.com/repos/`+owner+`/`+repo+`/git/refs/heads/`+tempHead+`$`),
		httpmock.NewStringResponder(200, "{}"),
	)

	// Assert PR title equals commit message when PRTitle empty
	httpmock.RegisterResponder("POST",
		"https://api.github.com/repos/"+owner+"/"+repo+"/pulls",
		func(req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			if !strings.Contains(string(b), `"title":"`+want+`"`) {
				t.Fatalf("expected PR title to default to commit message %q; body=%s", want, string(b))
			}
			return httpmock.NewJsonResponse(201, map[string]any{"number": 5})
		},
	)

	// No merge; MergeWithoutReview=false when matching config present and not set to true
	// If code attempted merge, there would be a 404 on PUT, failing the test via missing responder count.

	files := []github.RepositoryContent{{
		Name: github.String("only.txt"), Path: github.String("only.txt"),
		Content: github.String(base64.StdEncoding.EncodeToString([]byte("y"))),
	}}
	// cfg := types.Configs{TargetRepo: repo, TargetBranch: baseBranch /* MergeWithoutReview: false (zero value) */}
	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{{RepoName: repo, BranchPath: "refs/heads/" + baseBranch, RuleName: "", CommitStrategy: "pr"}: {TargetBranch: baseBranch, Content: files, CommitStrategy: "pr"}}

	services.AddFilesToTargetRepoBranch() // No longer takes parameters - uses FilesToUpload map

	// Ensure a PR was created but no merge occurred
	require.Equal(t, 1, test.CountByMethodAndURLRegexp("POST", regexp.MustCompile(`/pulls$`)))
	require.Equal(t, 0, test.CountByMethodAndURLRegexp("PUT", regexp.MustCompile(`/pulls/5/merge$`)))

	services.FilesToUpload = nil
}

// TestDeleteBranchIfExists_NilReference tests that deleteBranchIfExists handles nil references gracefully
func TestDeleteBranchIfExists_NilReference(t *testing.T) {
	_ = test.WithHTTPMock(t)

	// Force fresh token
	services.InstallationAccessToken = ""
	test.MockGitHubAppTokenEndpoint(os.Getenv(configs.InstallationId))
	services.ConfigurePermissions()

	// This should not panic or make any API calls when ref is nil
	// We're testing that the function returns early without attempting to delete
	ctx := context.Background()
	client := services.GetRestClient()

	// Call with nil reference - should return immediately without error
	services.DeleteBranchIfExistsExported(ctx, client, "test-org/test-repo", nil)

	// Verify no DELETE requests were made (since ref was nil)
	require.Equal(t, 0, test.CountByMethodAndURLRegexp("DELETE", regexp.MustCompile(`/git/refs/`)))
}
