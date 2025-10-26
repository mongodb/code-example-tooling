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

// installationTokenCache caches installation access tokens by organization name
var installationTokenCache = make(map[string]string)

// jwtToken caches the GitHub App JWT token
var jwtToken string
var jwtExpiry time.Time

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

// getWebhookSecretFromSecretManager retrieves the webhook secret from Google Cloud Secret Manager
func getWebhookSecretFromSecretManager(secretName string) (string, error) {
	if os.Getenv("SKIP_SECRET_MANAGER") == "true" {
		// For tests and local runs, use direct env var
		if secret := os.Getenv(configs.WebhookSecret); secret != "" {
			return secret, nil
		}
		return "", fmt.Errorf("SKIP_SECRET_MANAGER=true but no WEBHOOK_SECRET set")
	}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create Secret Manager client: %w", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}
	return string(result.Payload.Data), nil
}

// LoadWebhookSecret loads the webhook secret from Secret Manager or environment variable
func LoadWebhookSecret(config *configs.Config) error {
	// If webhook secret is already set directly, use it
	if config.WebhookSecret != "" {
		return nil
	}

	// Otherwise, load from Secret Manager
	secret, err := getWebhookSecretFromSecretManager(config.WebhookSecretName)
	if err != nil {
		return fmt.Errorf("failed to load webhook secret: %w", err)
	}
	config.WebhookSecret = secret
	return nil
}

// LoadMongoURI loads the MongoDB URI from Secret Manager or environment variable
func LoadMongoURI(config *configs.Config) error {
	// If MongoDB URI is already set directly, use it
	if config.MongoURI != "" {
		return nil
	}

	// If no secret name is configured, skip (audit logging is optional)
	if config.MongoURISecretName == "" {
		return nil
	}

	// Load from Secret Manager
	uri, err := getSecretFromSecretManager(config.MongoURISecretName, "MONGO_URI")
	if err != nil {
		return fmt.Errorf("failed to load MongoDB URI: %w", err)
	}
	config.MongoURI = uri
	return nil
}

// getSecretFromSecretManager is a generic function to retrieve any secret from Secret Manager
func getSecretFromSecretManager(secretName, envVarName string) (string, error) {
	if os.Getenv("SKIP_SECRET_MANAGER") == "true" {
		// For tests and local runs, use direct env var
		if secret := os.Getenv(envVarName); secret != "" {
			return secret, nil
		}
		return "", fmt.Errorf("SKIP_SECRET_MANAGER=true but no %s set", envVarName)
	}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create Secret Manager client: %w", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}
	return string(result.Payload.Data), nil
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
	if InstallationAccessToken == "" {
		ConfigurePermissions()
	}
	client := graphql.NewClient("https://api.github.com/graphql", &http.Client{
		Transport: &transport{token: InstallationAccessToken},
	})
	return client
}

// getOrRefreshJWT returns a valid JWT token, generating a new one if expired
func getOrRefreshJWT() (string, error) {
	// Check if we have a valid cached JWT
	if jwtToken != "" && time.Now().Before(jwtExpiry) {
		return jwtToken, nil
	}

	// Generate new JWT
	pemKey := getPrivateKeyFromSecret()
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemKey)
	if err != nil {
		return "", fmt.Errorf("unable to parse RSA private key: %w", err)
	}

	token, err := generateGitHubJWT(os.Getenv(configs.AppId), privateKey)
	if err != nil {
		return "", fmt.Errorf("error generating JWT: %w", err)
	}

	// Cache the JWT (expires in 10 minutes, cache for 9 to be safe)
	jwtToken = token
	jwtExpiry = time.Now().Add(9 * time.Minute)

	return token, nil
}

// getInstallationIDForOrg retrieves the installation ID for a specific organization
func getInstallationIDForOrg(org string) (string, error) {
	token, err := getOrRefreshJWT()
	if err != nil {
		return "", fmt.Errorf("failed to get JWT: %w", err)
	}

	url := "https://api.github.com/app/installations"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	hc := HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GET %s: %d %s %s", url, resp.StatusCode, resp.Status, body)
	}

	var installations []struct {
		ID      int64 `json:"id"`
		Account struct {
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"account"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&installations); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	// Find the installation for the specified organization
	for _, inst := range installations {
		if inst.Account.Login == org {
			return fmt.Sprintf("%d", inst.ID), nil
		}
	}

	return "", fmt.Errorf("no installation found for organization: %s", org)
}

// SetInstallationTokenForOrg sets a cached installation token for an organization.
// This is primarily used for testing to bypass the GitHub App authentication flow.
func SetInstallationTokenForOrg(org, token string) {
	installationTokenCache[org] = token
}

// GetRestClientForOrg returns a GitHub REST API client authenticated for a specific organization
func GetRestClientForOrg(org string) (*github.Client, error) {
	// Check if we have a cached token for this org
	if token, ok := installationTokenCache[org]; ok && token != "" {
		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
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
		return github.NewClient(httpClient), nil
	}

	// Get installation ID for the organization
	installationID, err := getInstallationIDForOrg(org)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation ID for org %s: %w", org, err)
	}

	// Get JWT token
	token, err := getOrRefreshJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT: %w", err)
	}

	// Get installation access token
	installationToken, err := getInstallationAccessToken(installationID, token, HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation token for org %s: %w", org, err)
	}

	// Cache the token
	installationTokenCache[org] = installationToken

	// Create and return client
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: installationToken})
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
	return github.NewClient(httpClient), nil
}

// RoundTrip adds the Authorization header to each request.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	return http.DefaultTransport.RoundTrip(req)
}
