package services_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/jarcoal/httpmock"
	test "github.com/mongodb/code-example-tooling/code-copier/tests"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/require"
)

//
// Shared helpers in test/utils.go
//

// Only one registration for POST /git/refs (the branch create). Use the exact URL and assert with info["POST "+createRefURL].
//
// Use the regex-based counter for dynamic endpoints (temp branch in path).
//
// getRefGetCount uses a tolerant regex so it matches both /git/ref/heads/main and /git/ref/refs/heads/main.
//
// We inject a dedicated http.Client and activate httpmock on it; services.HTTPClient points to that client, so both the app-installation token and the go-github client go through the mock.

//
// Global test setup
//

func TestMain(m *testing.M) {
	os.Setenv("REPO_OWNER", "my-org")
	os.Setenv("REPO_NAME", "target-repo")
	os.Setenv("INSTALLATION_ID", "12345")
	os.Setenv("GITHUB_APP_CLIENT_ID", "dummy-client-id")
	os.Setenv("SKIP_SECRET_MANAGER", "true")
	os.Setenv("GITHUB_APP_ID", "999999")

	// Generate a valid PKCS#1 RSA private key and export as both raw + base64
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	os.Setenv("GITHUB_APP_PRIVATE_KEY", string(pemBytes))
	os.Setenv("GITHUB_APP_PRIVATE_KEY_B64", base64.StdEncoding.EncodeToString(pemBytes))

	code := m.Run()

	// cleanup
	os.Unsetenv("REPO_OWNER")
	os.Unsetenv("REPO_NAME")
	os.Unsetenv("INSTALLATION_ID")
	os.Unsetenv("GITHUB_APP_CLIENT_ID")
	os.Unsetenv("SKIP_SECRET_MANAGER")
	os.Unsetenv("GITHUB_APP_ID")
	os.Unsetenv("GITHUB_APP_PRIVATE_KEY")
	os.Unsetenv("GITHUB_APP_PRIVATE_KEY_B64")

	os.Exit(code)
}

//
// Unit tests for the small map helpers
//

func TestAddToRepoAndFilesMap_NewEntry(t *testing.T) {
	services.FilesToUpload = nil
	name := "example.txt"
	dummyFile := github.RepositoryContent{Name: &name}

	services.AddToRepoAndFilesMap("TargetRepo1", "main", dummyFile)

	require.NotNil(t, services.FilesToUpload)
	key := types.UploadKey{RepoName: "TargetRepo1", BranchPath: "refs/heads/main"}
	entry, exists := services.FilesToUpload[key]
	require.True(t, exists)
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
	require.ElementsMatch(t,
		[]string{"level1/first.txt", "level1/level2/level3/nested-second.txt"},
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

//
// Integration-y tests for the write flows
//

func TestAddFilesToTargetRepoBranch_Succeeds(t *testing.T) {
	// Use a dedicated client and activate httpmock on it
	httpClient := &http.Client{}
	httpmock.ActivateNonDefault(httpClient)
	t.Cleanup(func() { httpmock.DeactivateAndReset() })
	prev := services.HTTPClient
	services.HTTPClient = httpClient
	t.Cleanup(func() { services.HTTPClient = prev })

	owner, repo := test.EnvOwnerRepo(t)
	branch := "main"

	// Installation token
	httpmock.RegisterResponder("POST",
		"https://api.github.com/app/installations/12345/access_tokens",
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"token": "test-installation-token"}),
	)

	// Git flow: GET ref -> POST tree -> POST commit -> PATCH ref
	getRefRe := regexp.MustCompile(`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/ref/(?:refs/)?heads/` + branch + `$`)
	httpmock.RegisterRegexpResponder("GET", getRefRe,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref":    "refs/heads/" + branch,
			"object": map[string]any{"sha": "baseSha"},
		}),
	)
	treesRe := regexp.MustCompile(`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/trees(\?.*)?$`)
	httpmock.RegisterRegexpResponder("POST", treesRe,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newTreeSha"}),
	)
	commitsURL := "https://api.github.com/repos/" + owner + "/" + repo + "/git/commits"
	httpmock.RegisterResponder("POST", commitsURL,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newCommitSha"}),
	)
	updateRefURL := "https://api.github.com/repos/" + owner + "/" + repo + "/git/refs/heads/" + branch
	httpmock.RegisterResponder("PATCH", updateRefURL, httpmock.NewStringResponder(200, `{}`))

	// Seed FilesToUpload
	files := []github.RepositoryContent{
		{Name: github.String("dir/example1.txt"), Path: github.String("dir/example1.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 1")))},
		{Name: github.String("dir/subdir/example2.txt"), Path: github.String("dir/subdir/example2.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 2")))},
	}
	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{
		{RepoName: repo, BranchPath: "refs/heads/" + branch}: {TargetBranch: branch, Content: files},
	}
	t.Cleanup(func() { services.FilesToUpload = nil })

	// Execute
	services.AddFilesToTargetRepoBranch()

	// Assert calls
	info := httpmock.GetCallCountInfo()
	require.Equal(t, 1, test.GetRefGetCount(owner, repo, branch))
	require.Equal(t, 1, info["POST "+commitsURL])
	require.Equal(t, 1, info["PATCH "+updateRefURL])

	// trees registered via regex
	treeCalls := 0
	for k, v := range info {
		if strings.HasPrefix(k, "POST https://api.github.com/repos/"+owner+"/"+repo+"/git/trees") {
			treeCalls += v
		}
	}
	require.Equal(t, 1, treeCalls)
}

func TestAddFilesToTargetRepoBranch_ViaPR_Succeeds(t *testing.T) {
	httpClient := &http.Client{}
	httpmock.ActivateNonDefault(httpClient)
	t.Cleanup(func() { httpmock.DeactivateAndReset() })
	prev := services.HTTPClient
	services.HTTPClient = httpClient
	t.Cleanup(func() { services.HTTPClient = prev })

	// force a fresh token
	services.InstallationAccessToken = ""
	t.Cleanup(func() { services.InstallationAccessToken = "" })

	t.Setenv("COPIER_COMMIT_STRATEGY", "pr")
	owner, repo := test.EnvOwnerRepo(t)
	branch := "main"
	tempHead := `copier/\d{8}-\d{6}`

	// 1) token stub
	httpmock.RegisterResponder("POST",
		"https://api.github.com/app/installations/12345/access_tokens",
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"token": "test-installation-token"}),
	)

	// 2) Create temp branch: GET base ref (main) + POST create ref (exact URL to avoid counting ambiguity)
	getMainRef := regexp.MustCompile(
		`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/ref/(?:refs/)?heads/` + branch + `$`,
	)
	httpmock.RegisterRegexpResponder("GET", getMainRef,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref":    "refs/heads/" + branch,
			"object": map[string]any{"sha": "baseSha"},
		}),
	)
	createRefURL := "https://api.github.com/repos/" + owner + "/" + repo + "/git/refs"
	httpmock.RegisterResponder("POST", createRefURL,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{
			"ref":    "refs/heads/copier/20250101-000000",
			"object": map[string]any{"sha": "baseSha"},
		}),
	)

	// 3) Write on temp branch: GET temp ref → POST tree → POST commit → PATCH temp ref
	getTempRef := regexp.MustCompile(
		`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/ref/(?:refs/)?heads/` + tempHead + `$`,
	)
	httpmock.RegisterRegexpResponder("GET", getTempRef,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref":    "refs/heads/copier/20250101-000000",
			"object": map[string]any{"sha": "baseSha"},
		}),
	)
	treesRe := regexp.MustCompile(`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/trees(\?.*)?$`)
	httpmock.RegisterRegexpResponder("POST", treesRe,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newTreeSha"}),
	)
	commitsURL := "https://api.github.com/repos/" + owner + "/" + repo + "/git/commits"
	httpmock.RegisterResponder("POST", commitsURL,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"sha": "newCommitSha"}),
	)
	updateTempRef := regexp.MustCompile(
		`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/refs/heads/` + tempHead + `$`,
	)
	httpmock.RegisterRegexpResponder("PATCH", updateTempRef, httpmock.NewStringResponder(200, `{}`))

	// 4) Create PR and merge PR
	httpmock.RegisterResponder("POST",
		"https://api.github.com/repos/"+owner+"/"+repo+"/pulls",
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"number": 42}),
	)
	httpmock.RegisterResponder("PUT",
		"https://api.github.com/repos/"+owner+"/"+repo+"/pulls/42/merge",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{"merged": true}),
	)

	// 5) Delete temp branch
	deleteTempRef := regexp.MustCompile(
		`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/refs/heads/` + tempHead + `$`,
	)
	httpmock.RegisterRegexpResponder("DELETE", deleteTempRef, httpmock.NewStringResponder(204, ""))

	services.ConfigurePermissions()

	// Seed FilesToUpload
	files := []github.RepositoryContent{
		{Name: github.String("dir/example1.txt"), Path: github.String("dir/example1.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 1")))},
		{Name: github.String("dir/example2.txt"), Path: github.String("dir/example2.txt"),
			Content: github.String(base64.StdEncoding.EncodeToString([]byte("hello 2")))},
	}
	services.FilesToUpload = map[types.UploadKey]types.UploadFileContent{
		{RepoName: repo, BranchPath: "refs/heads/" + branch}: {TargetBranch: branch, Content: files},
	}
	t.Cleanup(func() { services.FilesToUpload = nil })

	// Execute
	services.AddFilesToTargetRepoBranch()

	// Assert calls (exact where possible, regex-counter for dynamic paths)
	for k, v := range httpmock.GetCallCountInfo() {
		t.Logf("httpmock key: %q -> %d", k, v)
	}
	require.Equal(t, 1, test.CountByMethodAndURLRegexp("POST",
		regexp.MustCompile(`/app/installations/12345/access_tokens$`),
	))
	require.Equal(t, 1, test.CountByMethodAndURLRegexp("POST",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/pulls$`),
	))
	require.Equal(t, 1, test.CountByMethodAndURLRegexp("PUT",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/pulls/42/merge$`),
	))

	require.GreaterOrEqual(t, 1, test.GetRefGetCount(owner, repo, branch))
	require.GreaterOrEqual(t,
		test.CountByMethodAndURLRegexp("GET",
			regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/ref/(?:refs/)?heads/copier/\d{8}-\d{6}$`),
		),
		1,
	)
	require.GreaterOrEqual(t, test.CountByMethodAndURLRegexp("POST",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/trees`),
	),
		1,
	)
	require.GreaterOrEqual(t, test.CountByMethodAndURLRegexp("PATCH",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/refs/heads/copier/\d{8}-\d{6}$`),
	),
		1,
	)
	require.GreaterOrEqual(t, test.CountByMethodAndURLRegexp("DELETE",
		regexp.MustCompile(`/repos/`+regexp.QuoteMeta(owner)+`/`+regexp.QuoteMeta(repo)+`/git/refs/heads/copier/\d{8}-\d{6}$`),
	),
		1,
	)
}
