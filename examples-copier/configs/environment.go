package configs

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

// Config holds all environment configuration
type Config struct {
	EnvFile              string
	Port                 string
	RepoName             string
	RepoOwner            string
	AppClientId          string
	InstallationId       string
	CommiterName         string
	CommiterEmail        string
	ConfigFile           string
	DeprecationFile      string
	WebserverPath        string
	SrcBranch            string
	PEMKeyName           string
	CopierLogName        string
	GoogleCloudProjectId string // Google Cloud project ID, used for logging
	DefaultRecursiveCopy bool
}

const (
	EnvFile              = "ENV"
	Port                 = "PORT"
	RepoName             = "REPO_NAME"
	RepoOwner            = "REPO_OWNER"
	AppClientId          = "GITHUB_APP_CLIENT_ID"
	InstallationId       = "INSTALLATION_ID"
	CommiterName         = "COMMITER_NAME"
	CommiterEmail        = "COMMITER_EMAIL"
	ConfigFile           = "CONFIG_FILE"
	DeprecationFile      = "DEPRECATION_FILE"
	WebserverPath        = "WEBSERVER_PATH"
	SrcBranch            = "SRC_BRANCH"
	PEMKeyName           = "PEM_NAME"
	CopierLogName        = "COPIER_LOG_NAME"
	GoogleCloudProjectId = "GOOGLE_CLOUD_PROJECT_ID"
	DefaultRecursiveCopy = "DEFAULT_RECURSIVE_COPY"
)

// NewConfig returns a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Port:                 "8080",
		CommiterName:         "Copier Bot",
		CommiterEmail:        "bot@example.com",
		ConfigFile:           "config.json",
		DeprecationFile:      "deprecation.json",
		WebserverPath:        "/webhook",
		SrcBranch:            "main",                                                           // Default branch to copy from (NOTE: we are purposefully only allowing copying from `main` branch right now)
		PEMKeyName:           "projects/1054147886816/secrets/CODE_COPIER_PEM/versions/latest", // default secret name for GCP Secret Manager
		CopierLogName:        "copy-copier-log",                                                // default log name for logging to GCP
		GoogleCloudProjectId: "github-copy-code-examples",                                      // default project ID for logging to GCP
		DefaultRecursiveCopy: true,                                                             // system-wide default for recursive copying that individual config entries can override.
	}
}

// LoadEnvironment loads environment variables and returns populated Config
func LoadEnvironment(envFile string) (*Config, error) {
	// Initialize with defaults
	config := NewConfig()

	// Set the provided env file
	config.EnvFile = envFile

	// Get current environment (default to test)
	env := getEnvWithDefault(EnvFile, "test")

	// Define env files to load in order of precedence
	envFiles := []string{
		".env",
		".env." + env,
	}

	if config.EnvFile != "" {
		envFiles = append(envFiles, config.EnvFile)
	}

	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			if err = godotenv.Load(file); err != nil {
				return nil, fmt.Errorf("error loading env file %s: %w", file, err)
			}
		}
	}

	// Populate config from environment variables, with defaults where applicable
	config.Port = getEnvWithDefault(Port, config.Port)
	config.RepoName = os.Getenv(RepoName)
	config.RepoOwner = os.Getenv(RepoOwner)
	config.AppClientId = os.Getenv(AppClientId)
	config.InstallationId = os.Getenv(InstallationId)
	config.CommiterName = getEnvWithDefault(CommiterName, config.CommiterName)
	config.CommiterEmail = getEnvWithDefault(CommiterEmail, config.CommiterEmail)
	config.ConfigFile = getEnvWithDefault(ConfigFile, config.ConfigFile)
	config.DeprecationFile = getEnvWithDefault(DeprecationFile, config.DeprecationFile)
	config.WebserverPath = getEnvWithDefault(WebserverPath, config.WebserverPath)
	config.SrcBranch = getEnvWithDefault(SrcBranch, config.SrcBranch)
	config.PEMKeyName = getEnvWithDefault(PEMKeyName, config.PEMKeyName)
	config.DefaultRecursiveCopy = getBoolEnvWithDefault(DefaultRecursiveCopy, config.DefaultRecursiveCopy)

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// getEnvWithDefault returns the environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getBoolEnvWithDefault returns the boolean environment variable value or default if not set
func getBoolEnvWithDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(value) == "true"
}

// validateConfig checks if all required configuration values are set
func validateConfig(config *Config) error {
	var missingVars []string

	requiredVars := map[string]string{
		RepoName:       config.RepoName,
		RepoOwner:      config.RepoOwner,
		AppClientId:    config.AppClientId,
		InstallationId: config.InstallationId,
	}

	for name, value := range requiredVars {
		if value == "" {
			missingVars = append(missingVars, name)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}
