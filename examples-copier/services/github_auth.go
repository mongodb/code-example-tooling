package services

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// transport is a custom HTTP transport that adds the Authorization header to each request.
type transport struct {
	token string
}

var InstallationAccessToken string
var HTTPClient = http.DefaultClient

// ConfigurePermissions sets up the necessary permissions to interact with the GitHub API.
// It retrieves the GitHub App's private key from Google Secret Manager, generates a JWT,
// and exchanges it for an installation access token.
func ConfigurePermissions() {
	envFilePath := os.Getenv("ENV_FILE")

	_, err := configs.LoadEnvironment(envFilePath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to load environment"))

	}

	pemKey := getPrivateKeyFromSecret()
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemKey)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Unable to parse RSA private key"))
	}

	// Generate JWT — use the numeric GitHub App ID (GITHUB_APP_ID) as "iss"
	token, err := generateGitHubJWT(os.Getenv(configs.AppId), privateKey)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error generating JWT"))
	}

	installationToken, err := getInstallationAccessToken("", token, HTTPClient)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error getting installation access token"))
	}
	InstallationAccessToken = installationToken
}

// generateGitHubJWT creates a JWT for GitHub App authentication.
func generateGitHubJWT(appID string, privateKey *rsa.PrivateKey) (string, error) {
	// Create a new JWT token
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),                       // Issued at
		"exp": now.Add(time.Minute * 10).Unix(), // Expiration time, 10 minutes from issue
		"iss": appID,                            // GitHub App ID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// Sign the JWT with the private key
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("unable to sign JWT: %v", err)
	}
	return signedToken, nil
}

// getPrivateKeyFromSecret retrieves the GitHub App's private key from Google Secret Manager.
// It supports local testing by allowing the key to be provided via environment variables.
func getPrivateKeyFromSecret() []byte {
	if os.Getenv("SKIP_SECRET_MANAGER") == "true" { // for tests and local runs
		if pem := os.Getenv("GITHUB_APP_PRIVATE_KEY"); pem != "" {
			return []byte(pem)
		}
		if b64 := os.Getenv("GITHUB_APP_PRIVATE_KEY_B64"); b64 != "" {
			dec, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				log.Fatalf("Invalid base64 private key: %v", err)
			}
			return dec
		}
		log.Fatalf("SKIP_SECRET_MANAGER=true but no GITHUB_APP_PRIVATE_KEY or GITHUB_APP_PRIVATE_KEY_B64 set")
	}
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)

	if err != nil {
		log.Fatalf("Failed to create Secret Manager client: %v", err)
	}
	defer client.Close()

	secretName := os.Getenv(configs.PEMKeyName)
	if secretName == "" {
		secretName = configs.NewConfig().PEMKeyName
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatalf("Failed to access secret version: %v", err)
	}
	return result.Payload.Data
}

// getInstallationAccessToken exchanges a JWT for a GitHub App installation access token.
func getInstallationAccessToken(installationId, jwtToken string, hc *http.Client) (string, error) {
	if installationId == "" || installationId == configs.InstallationId {
		installationId = os.Getenv(configs.InstallationId)
	}
	if installationId == "" {
		return "", fmt.Errorf("missing installation ID")
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationId)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}
	var out struct {
		Token string `json:"token"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return out.Token, nil
}

// GetRestClient returns a GitHub REST API client authenticated with the installation access token.
func GetRestClient() *github.Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: InstallationAccessToken})

	base := http.DefaultTransport
	if HTTPClient != nil && HTTPClient.Transport != nil {
		base = HTTPClient.Transport
	}

	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: src,
			Base:   base,
		},
	}
	return github.NewClient(httpClient)
}

func GetGraphQLClient() *graphql.Client {
	ConfigurePermissions()
	client := graphql.NewClient("https://api.github.com/graphql", &http.Client{
		Transport: &transport{token: InstallationAccessToken},
	})
	return client
}

// RoundTrip adds the Authorization header to each request.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	return http.DefaultTransport.RoundTrip(req)
}
