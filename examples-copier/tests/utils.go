package test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
)

//
// Environment helpers
//

// EnvOwnerRepo returns owner/repo from env and fails the test if either is missing.
func EnvOwnerRepo(t testing.TB) (string, string) {
	t.Helper()
	owner := os.Getenv(configs.RepoOwner)
	repo := os.Getenv(configs.RepoName)
	if owner == "" || repo == "" {
		t.Fatalf("REPO_OWNER/REPO_NAME not set")
	}
	return owner, repo
}

//
// HTTP/test wiring helpers
//

// WithHTTPMock wraps a test in `httpmock` activation on a dedicated http.Client and routes services.HTTPClient through it.
// Used in any test that needs multiple mock endpoints. Wrap t.Run blocks to avoid leftover mocks affecting other tests.
func WithHTTPMock(t testing.TB) *http.Client {
	t.Helper()
	c := &http.Client{}
	httpmock.ActivateNonDefault(c)
	t.Cleanup(func() { httpmock.DeactivateAndReset() })
	prev := services.HTTPClient
	services.HTTPClient = c
	t.Cleanup(func() { services.HTTPClient = prev })
	return c
}

// DumpHttpmockCalls logs all recorded httpmock keys and counts. Used for debugging while writing tests.
func DumpHttpmockCalls(t testing.TB) {
	t.Helper()
	for k, v := range httpmock.GetCallCountInfo() {
		t.Logf("httpmock key: %q -> %d", k, v)
	}
}

//
// Mock registration helpers
//

// MockGitHubAppTokenEndpoint mocks the GitHub App installation token endpoint with a fixed fake token.
// Used in to simulate any auth-triggered flow without needing a real installation ID.
func MockGitHubAppTokenEndpoint(installationID string) {
	httpmock.RegisterResponder("POST",
		"https://api.github.com/app/installations/"+installationID+"/access_tokens",
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"token": "test-installation-token"}),
	)
}

// MockGitHubAppInstallations mocks the GitHub App installations list endpoint.
// Used to simulate fetching installation IDs for organizations.
func MockGitHubAppInstallations(orgToInstallationID map[string]string) {
	installations := []map[string]any{}
	for org, installID := range orgToInstallationID {
		installations = append(installations, map[string]any{
			"id": installID,
			"account": map[string]any{
				"login": org,
				"type":  "Organization",
			},
		})
	}
	httpmock.RegisterResponder("GET",
		"https://api.github.com/app/installations",
		httpmock.NewJsonResponderOrPanic(200, installations),
	)
}

// SetupOrgToken sets up a cached installation token for an organization.
// This bypasses the need to mock the installations and token endpoints.
func SetupOrgToken(org, token string) {
	services.SetInstallationTokenForOrg(org, token)
}

// MockGitHubWriteEndpoints mocks the full direct-commit flow endpoints for a single branch: GET base ref, POST trees, POST commits, PATCH ref.
// Used to simulate writing to a GitHub repo without creating a PR.
// Returns the URLs for the base ref, commits, and update ref endpoints.
func MockGitHubWriteEndpoints(owner, repo, branch string) (baseRefURL, commitsURL, updateRefURL string) {
	baseRefURL = "https://api.github.com/repos/" + owner + "/" + repo + "/git/ref/heads/" + branch
	httpmock.RegisterResponder("GET", baseRefURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref": "refs/heads/" + branch,
			"object": map[string]any{
				"sha": "baseSha",
			},
		}),
	)

	treesRe := regexp.MustCompile(`^https://api\.github\.com/repos/` + regexp.QuoteMeta(owner) + `/` +
		regexp.QuoteMeta(repo) + `/git/trees(\?.*)?$`)
	httpmock.RegisterRegexpResponder("POST", treesRe,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{
			"sha": "newTreeSha",
		}),
	)

	commitsURL = "https://api.github.com/repos/" + owner + "/" + repo + "/git/commits"
	httpmock.RegisterResponder("POST", commitsURL,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{
			"sha": "newCommitSha",
		}),
	)

	updateRefURL = "https://api.github.com/repos/" + owner + "/" + repo + "/git/refs/heads/" + branch
	httpmock.RegisterResponder("PATCH", updateRefURL,
		httpmock.NewStringResponder(200, `{}`),
	)

	return
}

// MockContentsEndpoint mocks GET file contents for a given path/ref.
// Used to simulate reading a file from a GitHub repo.
func MockContentsEndpoint(owner, repo, path, contentB64 string) {
	re := regexp.MustCompile(
		`^https://api\.github\.com/repos/` + regexp.QuoteMeta(owner) + `/` +
			regexp.QuoteMeta(repo) + `/contents/` + regexp.QuoteMeta(path) +
			`\?ref=(?:main|SRC_BRANCH|release/[0-9.]+)$`,
	)
	httpmock.RegisterRegexpResponder("GET", re,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"type":     "file",
			"encoding": "base64",
			"path":     path,
			"content":  contentB64,
		}),
	)
}

// MockCreateRef mocks POST to create a new temp branch ref. Returns the exact URL for call-count asserts.
// Used to simulate creating a new branch for writing files without actually pushing to GitHub.
func MockCreateRef(owner, repo string) string {
	url := "https://api.github.com/repos/" + owner + "/" + repo + "/git/refs"
	httpmock.RegisterResponder("POST", url,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{
			"ref":    "refs/heads/copier/20250101-000000",
			"object": map[string]any{"sha": "baseSha"},
		}),
	)
	return url
}

// MockPullsAndMerge mocks creating and merging a PR.
// Used to simulate the full PR flow for functions that create a PR and then merge it
func MockPullsAndMerge(owner, repo string, number int) {
	httpmock.RegisterResponder("POST",
		"https://api.github.com/repos/"+owner+"/"+repo+"/pulls",
		httpmock.NewJsonResponderOrPanic(201, map[string]any{"number": number}),
	)
	httpmock.RegisterResponder("PUT",
		"https://api.github.com/repos/"+owner+"/"+repo+fmt.Sprintf("/pulls/%d/merge", number),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{"merged": true}),
	)
}

// MockDeleteTempRef mocks DELETE to remove a temporary branch ref.
// Used to simulate cleaning up after writing files without actually deleting a branch on GitHub.
func MockDeleteTempRef(owner, repo string) {
	re := regexp.MustCompile(
		`^https://api\.github\.com/repos/` + regexp.QuoteMeta(owner) + `/` +
			regexp.QuoteMeta(repo) + `/git/refs/heads/copier/\d{8}-\d{6}$`,
	)
	httpmock.RegisterRegexpResponder("DELETE", re, httpmock.NewStringResponder(204, ""))
}

//
// Staging/assertion helpers
//

// NormalizeUpload flattens FilesToUpload to UploadKey -> []names for simpler comparisons.
func NormalizeUpload(in map[types.UploadKey]types.UploadFileContent) map[types.UploadKey][]string {
	out := make(map[types.UploadKey][]string, len(in))
	for k, v := range in {
		names := make([]string, 0, len(v.Content))
		for _, c := range v.Content {
			names = append(names, c.GetName())
		}
		out[k] = names
	}
	return out
}

// MakeChanged is a shorthand to build ChangedFile entries.
func MakeChanged(status, path string) types.ChangedFile {
	return types.ChangedFile{Status: status, Path: path}
}

// ResetGlobals clears FilesToUpload and FilesToDeprecate.
func ResetGlobals() {
	services.FilesToUpload = nil
	services.FilesToDeprecate = nil
}

// AssertUploadedPaths asserts that the staged filenames match the want for the given repo/branch (order-insensitive).
func AssertUploadedPaths(t *testing.T, repo, branch string, want []string) {
	t.Helper()
	key := types.UploadKey{RepoName: repo, BranchPath: "refs/heads/" + branch}
	got, ok := services.FilesToUpload[key]
	if !ok {
		t.Fatalf("expected FilesToUpload to contain key for %s/%s", repo, branch)
	}

	var names []string
	for _, c := range got.Content {
		n := c.GetName()
		if n == "" {
			n = c.GetPath() // fallback: some code paths populate only Path
		}
		names = append(names, n)
	}

	// exact, order-insensitive comparison
	if len(want) == 0 && len(names) == 0 {
		return
	}
	if len(want) != len(names) {
		t.Fatalf("staged names length mismatch: got=%v want=%v", names, want)
	}
	wantSet := map[string]struct{}{}
	for _, w := range want {
		wantSet[w] = struct{}{}
	}
	for _, n := range names {
		if _, ok := wantSet[n]; !ok {
			t.Fatalf("unexpected staged path %q; got=%v want=%v", n, names, want)
		}
	}
}

// AssertUploadedPathsFromConfig converts staged source paths to target paths using cfg,
// then compares against want - i.e. target paths (order-insensitive).
// Used when the staged files are from a config that specifies source/target directories.
func AssertUploadedPathsFromConfig(t *testing.T, cfg types.Configs, want []string) {
	t.Helper()
	key := types.UploadKey{RepoName: cfg.TargetRepo, BranchPath: "refs/heads/" + cfg.TargetBranch}
	got, ok := services.FilesToUpload[key]
	if !ok {
		t.Fatalf("expected FilesToUpload to contain key for %s/%s", cfg.TargetRepo, cfg.TargetBranch)
	}
	var names []string
	for _, c := range got.Content {
		// Prefer Name if present
		n := c.GetName()
		if n == "" {
			n = c.GetPath() // usually the *source* path (e.g. examples/…)
		}
		// If the staged name looks like a source path, rewrite to target
		if cfg.SourceDirectory != "" && strings.HasPrefix(n, cfg.SourceDirectory) {
			rel := strings.TrimPrefix(n, cfg.SourceDirectory)
			rel = strings.TrimPrefix(rel, "/")
			n = cfg.TargetDirectory
			if rel != "" {
				n = cfg.TargetDirectory + "/" + rel
			}
		}
		names = append(names, n)
	}

	// order-insensitive compare
	if len(want) == 0 && len(names) == 0 {
		return
	}
	if len(want) != len(names) {
		t.Fatalf("staged names length mismatch: got=%v want=%v", names, want)
	}
	wantSet := map[string]struct{}{}
	for _, w := range want {
		wantSet[w] = struct{}{}
	}
	for _, n := range names {
		if _, ok = wantSet[n]; !ok {
			t.Fatalf("unexpected staged path %q; got=%v want=%v", n, names, want)
		}
	}
}

// CountByMethodAndURLRegexp adds up call counts for a given METHOD whose stored httpmock key's URL matches urlRE.
// Works for both exact and regex-registered responders.
// Used to assert that a specific endpoint was called a certain number of times.
func CountByMethodAndURLRegexp(method string, urlRE *regexp.Regexp) int {
	info := httpmock.GetCallCountInfo()
	total := 0
	for k, v := range info {
		if !(strings.HasPrefix(k, method+" ") || strings.HasPrefix(k, method+"=~")) {
			continue
		}
		var urlish string
		switch {
		case strings.HasPrefix(k, method+"=~"):
			urlish = strings.TrimPrefix(k, method+"=~")
		case strings.HasPrefix(k, method+" "):
			urlish = strings.TrimPrefix(k, method+" ")
		default:
			continue
		}
		urlish = strings.Trim(urlish, "^$")
		urlish = strings.ReplaceAll(urlish, `\`, "")
		if urlRE.MatchString(urlish) {
			total += v
		}
	}
	return total
}

// GetRefGetCount counts GET calls to /git/ref/(refs/)?heads/<branch>
// for the given owner/repo/branch. Used to assert that a ref was fetched.
func GetRefGetCount(owner, repo, branch string) int {
	re := regexp.MustCompile(`/repos/` + regexp.QuoteMeta(owner) + `/` + regexp.QuoteMeta(repo) +
		`/git/ref/(?:refs/)?heads/` + regexp.QuoteMeta(branch) + `$`)
	return CountByMethodAndURLRegexp("GET", re)
}
