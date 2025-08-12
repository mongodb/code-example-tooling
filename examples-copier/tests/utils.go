package test

import (
	"encoding/base64"
	"github.com/jarcoal/httpmock"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"os"
	"regexp"
	"strings"
	"testing"
)

func MockGitHubAppTokenEndpoint(installationID string) {
	url := "https://api.github.com/app/installations/" + installationID + "/access_tokens"
	httpmock.RegisterResponder("POST", url,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{
			"token": "test-installation-token",
		}),
	)
}

func MockGitHubWriteEndpoints(owner, repo, branch string) (baseRefURL, commitsURL, updateRefURL string) {
	// GET base ref
	baseRefURL = "https://api.github.com/repos/" + owner + "/" + repo + "/git/ref/heads/" + branch
	httpmock.RegisterResponder("GET", baseRefURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ref": "refs/heads/" + branch,
			"object": map[string]any{
				"sha": "baseSha",
			},
		}),
	)

	// POST create tree (allow optional query like ?base_tree=...)
	treesRe := regexp.MustCompile(`^https://api\.github\.com/repos/` + owner + `/` + repo + `/git/trees(\?.*)?$`)
	httpmock.RegisterRegexpResponder("POST", treesRe,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{
			"sha": "newTreeSha",
		}),
	)

	// POST create commit
	commitsURL = "https://api.github.com/repos/" + owner + "/" + repo + "/git/commits"
	httpmock.RegisterResponder("POST", commitsURL,
		httpmock.NewJsonResponderOrPanic(201, map[string]any{
			"sha": "newCommitSha",
		}),
	)

	// PATCH update ref
	updateRefURL = "https://api.github.com/repos/" + owner + "/" + repo + "/git/refs/heads/" + branch
	httpmock.RegisterResponder("PATCH", updateRefURL,
		httpmock.NewStringResponder(200, `{}`),
	)

	return
}

// Belt-and-suspenders: stub Google OAuth + Secret Manager so tests never fail if GSM is touched.
func MockGoogleSecretsEndpoints() {
	// OAuth token exchange
	httpmock.RegisterResponder(
		"POST",
		"https://oauth2.googleapis.com/token",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"access_token": "ya29.test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}),
	)

	// Secret Manager access endpoint
	smAccess := regexp.MustCompile(`^https://secretmanager\.googleapis\.com/v1/projects/[^/]+/secrets/[^/]+/versions/[^:]+:access$`)
	httpmock.RegisterRegexpResponder(
		"GET",
		smAccess,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"payload": map[string]any{
				// any base64 payload is fine; loader shouldn’t be called if SKIP_SECRET_MANAGER=true
				"data": base64.StdEncoding.EncodeToString([]byte("test-secret")),
			},
		}),
	)
}

// Utility: count POST /git/trees even when matched by regex
func GetTreePostCount() int {
	info := httpmock.GetCallCountInfo()
	count := 0
	for k, v := range info {
		if strings.HasPrefix(k, "POST https://api.github.com/repos/") && strings.HasSuffix(k, "/git/trees") {
			count += v
		}
		// Some httpmock versions include the query string; be flexible:
		if strings.HasPrefix(k, "POST https://api.github.com/repos/") && strings.Contains(k, "/git/trees?") {
			count += v
		}
	}
	return count
}

// count calls in httpmock.GetCallCountInfo() that start with METHOD (e.g., "GET"/"POST"/"PATCH"/"PUT")
// and contain ALL provided substrings (works for both exact and regex-registered responders).
func CountByMethodAndContains(method string, subs ...string) int {
	info := httpmock.GetCallCountInfo()
	total := 0
	for k, v := range info {
		if !strings.HasPrefix(k, method) { // matches both "GET https..." and "GET=~^https..."
			continue
		}
		matches := true
		for _, s := range subs {
			if !strings.Contains(k, s) {
				matches = false
				break
			}
		}
		if matches {
			total += v
		}
	}
	return total
}

// Count calls by METHOD and a URL regex against httpmock keys (works for exact + regex responders)
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
			urlish = strings.TrimPrefix(k, method+"=~") // "^https:\/\/api\.github\.com\/..."
		case strings.HasPrefix(k, method+" "):
			urlish = strings.TrimPrefix(k, method+" ") // "https://api.github.com/..."
		default:
			continue
		}

		// Normalize regex keys
		urlish = strings.Trim(urlish, "^$")
		urlish = strings.ReplaceAll(urlish, `\`, "")

		if urlRE.MatchString(urlish) {
			total += v
		}
	}
	return total
}

// Convenience for GET /git/ref/(refs/)?heads/<branch>
func GetRefGetCount(owner, repo, branch string) int {
	re := regexp.MustCompile(`/repos/` + regexp.QuoteMeta(owner) + `/` + regexp.QuoteMeta(repo) + `/git/ref/(?:refs/)?heads/` + regexp.QuoteMeta(branch) + `$`)
	return CountByMethodAndURLRegexp("GET", re)
}

// Return env values for owner/repo (fail fast if unset)
func EnvOwnerRepo(t testing.TB) (string, string) {
	t.Helper()
	owner := os.Getenv(configs.RepoOwner)
	repo := os.Getenv(configs.RepoName)
	if owner == "" || repo == "" {
		t.Fatalf("REPO_OWNER/REPO_NAME not set")
	}
	return owner, repo
}
