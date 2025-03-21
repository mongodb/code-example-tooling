package services

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v48/github"
	"github.com/shurcooL/graphql"
	. "github.com/thompsch/app-tester/configs"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"time"
)

var InstallationAccessToken string

func ConfigurePermissions() {
	LoadEnvironment()
	privateKeyPath := "go-github-mdb-app.2025-03-04.private-key.pem"
	// Read the private key file
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		LogError(fmt.Sprintf("Unable to read private key: %v", err))
	}
	// Parse RSA private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		LogError(fmt.Sprintf("Unable to parse RSA private key: %v", err))
	}
	// Generate JWT
	token, err := generateGitHubJWT(AppClientId, privateKey)
	if err != nil {
		LogError(fmt.Sprintf("Error generating JWT: %v", err))
	}
	installation_token := getInstallationAccessToken(InstallationId, token)
	InstallationAccessToken = installation_token
}

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

func getInstallationAccessToken(installationId, jwtToken string) string {
	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationId)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		LogError(fmt.Sprintf("failed to create request: %v", err))
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError(fmt.Sprintf("failed to execute request: %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		LogError(fmt.Sprintf("failed to get access token: status %d", resp.StatusCode))
	}
	var result struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		LogError(fmt.Sprintf("failed to decode response: %v", err))
	}
	return result.Token
}

func GetRestClient() *github.Client {
	ConfigurePermissions()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: InstallationAccessToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	gitHubClient := github.NewClient(tc)
	return gitHubClient
}

func GetGraphQLClient() *graphql.Client {
	ConfigurePermissions()
	client := graphql.NewClient("https://api.github.com/graphql", &http.Client{
		Transport: &transport{token: InstallationAccessToken},
	})
	return client
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	return http.DefaultTransport.RoundTrip(req)
}

type transport struct {
	token string
}
