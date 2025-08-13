package services_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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
	os.Setenv(configs.AppClientId, "dummy-client-id")
	os.Setenv("SKIP_SECRET_MANAGER", "true")
	os.Setenv("SRC_BRANCH", "main")

	// Provide an RSA private key (both raw and b64) so ConfigurePermissions can parse.
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	os.Setenv("GITHUB_APP_ID", "999999")
	os.Setenv("GITHUB_APP_PRIVATE_KEY", string(pemBytes))
	os.Setenv("GITHUB_APP_PRIVATE_KEY_B64", base64.StdEncoding.EncodeToString(pemBytes))

	code := m.Run()

	// Cleanup
	os.Unsetenv(configs.RepoOwner)
	os.Unsetenv(configs.RepoName)
	os.Unsetenv(configs.InstallationId)
	os.Unsetenv(configs.AppClientId)
	os.Unsetenv("SKIP_SECRET_MANAGER")
	os.Unsetenv("SRC_BRANCH")
	os.Unsetenv("GITHUB_APP_ID")
	os.Unsetenv("GITHUB_APP_PRIVATE_KEY")
	os.Unsetenv("GITHUB_APP_PRIVATE_KEY_B64")

	os.Exit(code)
}

func TestAddToRepoAndFilesMap_NewEntry(t *testing.T) {
	services.FilesToUpload = nil

	name := "example.txt"
	dummyFile := github.RepositoryContent{Name: &name}

	services.AddToRepoAndFilesMap("TargetRepo1", "main", dummyFile)

	require.NotNil(t, services.FilesToUpload, "FilesToUpload map should be initialized")
	key := types.UploadKey{RepoName: "TargetRepo1", BranchPath: "refs/heads/main"}
	entry, exists := services.FilesToUpload[key]
	require.True(t, exists, "Entry for TargetRepo1/main should exist")
	require.Equal(t, "main", entry.TargetBranch)
	require.Len(t, entry.Content, 1)
	require.Equal(t, "example.txt", *entry.Content[0].Name)
}

func TestAddToRepoAndFilesMap_AppendEntry(t *testing.T) {
	services.FilesToUpload = make(map[types.UploadKey]types.UploadFileContent)
	key := types.UploadKey{RepoName: "TargetRepo1", BranchPath: "refs/heads/main"}

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
	key := types.UploadKey{RepoName: "TargetRepo1", BranchPath: "refs/heads/main"}

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

func TestAddFilesToTargetRepoBranch_Succeeds(t *testing.T) {
	_ = test.WithHTTPMock(t)

	owner, repo := test.EnvOwnerRepo(t)
	branch := "main"
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
			TargetBranch: baseBranch,
			Content:      files,
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
