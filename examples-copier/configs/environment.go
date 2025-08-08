package configs

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

// Config holds all environment configuration
type Config struct {
	EnvFile              string // Backward compatibility
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
	CopierLogName        string // Name of the log for the copier
	GoogleCloudProjectId string // Google Cloud project ID, used for logging
	DefaultRecursiveCopy bool
}

const (
	EnvFile              = "ENV" // Backward compatibility
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
	PEMKeyName           = "PEM_NAME"        // Name of the environment variable for the PEM key
	CopierLogName        = "COPIER_LOG_NAME" // Name of the log for the copier
	GoogleCloudProjectId = "GOOGLE_CLOUD_PROJECT_ID"
	DefaultRecursiveCopy = "DEFAULT_RECURSIVE_COPY"
)

// NewConfig returns a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Port:            "8080",
		CommiterName:    "Copier Bot",
		CommiterEmail:   "bot@example.com",
		ConfigFile:      "config.json",
		DeprecationFile: "deprecation.json",
		WebserverPath:   "/webhook",
		SrcBranch:       "main", // Default branch to copy from
		// NOTE: we are purposefully only allowing copying from `main` branch right now
		PEMKeyName:           "projects/1054147886816/secrets/CODE_COPIER_PEM/versions/latest",
		CopierLogName:        "copy-copier-log",
		GoogleCloudProjectId: "github-copy-code-examples",
		DefaultRecursiveCopy: true, // system-wide default for recursive copying that individual config entries can override.
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

	// Add EnvFile if specified (highest precedence)
	if config.EnvFile != "" {
		envFiles = append(envFiles, config.EnvFile)
	}

	// Load each env file if it exists
	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			if err = godotenv.Load(file); err != nil {
				return nil, fmt.Errorf("error loading env file %s: %w", file, err)
			}
		}
	}

	// Populate config from environment variables
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

	// Define required variables
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
