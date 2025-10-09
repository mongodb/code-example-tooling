package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all environment configuration
type Config struct {
	EnvFile              string
	Port                 string
	RepoName             string
	RepoOwner            string
	AppId                string
	AppClientId          string
	InstallationId       string
	CommitterName        string
	CommitterEmail       string
	ConfigFile           string
	DeprecationFile      string
	WebserverPath        string
	SrcBranch            string
	PEMKeyName           string
	WebhookSecretName    string
	WebhookSecret        string
	CopierLogName        string
	GoogleCloudProjectId string
	DefaultRecursiveCopy bool
	DefaultPRMerge       bool
	DefaultCommitMessage string

	// New features
	DryRun              bool
	AuditEnabled        bool
	MongoURI            string
	MongoURISecretName  string
	AuditDatabase       string
	AuditCollection     string
	MetricsEnabled      bool

	// Slack notifications
	SlackWebhookURL  string
	SlackChannel     string
	SlackUsername    string
	SlackIconEmoji   string
	SlackEnabled     bool
}

const (
	EnvFile              = "ENV"
	Port                 = "PORT"
	RepoName             = "REPO_NAME"
	RepoOwner            = "REPO_OWNER"
	AppId                = "GITHUB_APP_ID"
	AppClientId          = "GITHUB_APP_CLIENT_ID"
	InstallationId       = "INSTALLATION_ID"
	CommitterName        = "COMMITTER_NAME"
	CommitterEmail       = "COMMITTER_EMAIL"
	ConfigFile           = "CONFIG_FILE"
	DeprecationFile      = "DEPRECATION_FILE"
	WebserverPath        = "WEBSERVER_PATH"
	SrcBranch            = "SRC_BRANCH"
	PEMKeyName           = "PEM_NAME"
	WebhookSecretName    = "WEBHOOK_SECRET_NAME"
	WebhookSecret        = "WEBHOOK_SECRET"
	CopierLogName        = "COPIER_LOG_NAME"
	GoogleCloudProjectId = "GOOGLE_CLOUD_PROJECT_ID"
	DefaultRecursiveCopy = "DEFAULT_RECURSIVE_COPY"
	DefaultPRMerge       = "DEFAULT_PR_MERGE"
	DefaultCommitMessage = "DEFAULT_COMMIT_MESSAGE"
	DryRun               = "DRY_RUN"
	AuditEnabled         = "AUDIT_ENABLED"
	MongoURI             = "MONGO_URI"
	MongoURISecretName   = "MONGO_URI_SECRET_NAME"
	AuditDatabase        = "AUDIT_DATABASE"
	AuditCollection      = "AUDIT_COLLECTION"
	MetricsEnabled       = "METRICS_ENABLED"
	SlackWebhookURL      = "SLACK_WEBHOOK_URL"
	SlackChannel         = "SLACK_CHANNEL"
	SlackUsername        = "SLACK_USERNAME"
	SlackIconEmoji       = "SLACK_ICON_EMOJI"
	SlackEnabled         = "SLACK_ENABLED"
)

// NewConfig returns a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Port:                 "8080",
		CommitterName:        "Copier Bot",
		CommitterEmail:       "bot@example.com",
		ConfigFile:           "copier-config.yaml",
		DeprecationFile:      "deprecated_examples.json",
		WebserverPath:        "/webhook",
		SrcBranch:            "main",                                                           // Default branch to copy from (NOTE: we are purposefully only allowing copying from `main` branch right now)
		PEMKeyName:           "projects/1054147886816/secrets/CODE_COPIER_PEM/versions/latest", // default secret name for GCP Secret Manager
		WebhookSecretName:    "projects/1054147886816/secrets/webhook-secret/versions/latest",  // default webhook secret name for GCP Secret Manager
		CopierLogName:        "copy-copier-log",                                                // default log name for logging to GCP
		GoogleCloudProjectId: "github-copy-code-examples",                                      // default project ID for logging to GCP
		DefaultRecursiveCopy: true,                                                             // system-wide default for recursive copying that individual config entries can override.
		DefaultPRMerge:       false,                                                            // system-wide default for PR merge without review that individual config entries can override.
		DefaultCommitMessage: "Automated PR with updated examples",                             // default commit message used when per-config commit_message is absent.
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
	config.AppId = os.Getenv(AppId)
	config.AppClientId = os.Getenv(AppClientId)
	config.InstallationId = os.Getenv(InstallationId)
	config.CommitterName = getEnvWithDefault(CommitterName, config.CommitterName)
	config.CommitterEmail = getEnvWithDefault(CommitterEmail, config.CommitterEmail)
	config.ConfigFile = getEnvWithDefault(ConfigFile, config.ConfigFile)
	config.DeprecationFile = getEnvWithDefault(DeprecationFile, config.DeprecationFile)
	config.WebserverPath = getEnvWithDefault(WebserverPath, config.WebserverPath)
	config.SrcBranch = getEnvWithDefault(SrcBranch, config.SrcBranch)
	config.PEMKeyName = getEnvWithDefault(PEMKeyName, config.PEMKeyName)
	config.WebhookSecretName = getEnvWithDefault(WebhookSecretName, config.WebhookSecretName)
	config.WebhookSecret = os.Getenv(WebhookSecret)
	config.DefaultRecursiveCopy = getBoolEnvWithDefault(DefaultRecursiveCopy, config.DefaultRecursiveCopy)
	config.DefaultPRMerge = getBoolEnvWithDefault(DefaultPRMerge, config.DefaultPRMerge)
	config.CopierLogName = getEnvWithDefault(CopierLogName, config.CopierLogName)
	config.GoogleCloudProjectId = getEnvWithDefault(GoogleCloudProjectId, config.GoogleCloudProjectId)
	config.DefaultCommitMessage = getEnvWithDefault(DefaultCommitMessage, config.DefaultCommitMessage)

	// New features
	config.DryRun = getBoolEnvWithDefault(DryRun, false)
	config.AuditEnabled = getBoolEnvWithDefault(AuditEnabled, false)
	config.MongoURI = os.Getenv(MongoURI)
	config.MongoURISecretName = os.Getenv(MongoURISecretName)
	config.AuditDatabase = getEnvWithDefault(AuditDatabase, "copier_audit")
	config.AuditCollection = getEnvWithDefault(AuditCollection, "events")
	config.MetricsEnabled = getBoolEnvWithDefault(MetricsEnabled, true)
	config.WebhookSecret = os.Getenv(WebhookSecret)

	// Slack notifications
	config.SlackWebhookURL = os.Getenv(SlackWebhookURL)
	config.SlackChannel = getEnvWithDefault(SlackChannel, "#code-examples")
	config.SlackUsername = getEnvWithDefault(SlackUsername, "Examples Copier")
	config.SlackIconEmoji = getEnvWithDefault(SlackIconEmoji, ":robot_face:")
	config.SlackEnabled = getBoolEnvWithDefault(SlackEnabled, config.SlackWebhookURL != "")

	// Export resolved values back into environment so downstream os.Getenv sees defaults
	_ = os.Setenv(Port, config.Port)
	_ = os.Setenv(RepoName, config.RepoName)
	_ = os.Setenv(RepoOwner, config.RepoOwner)
	_ = os.Setenv(AppId, config.AppId)
	_ = os.Setenv(AppClientId, config.AppClientId)
	_ = os.Setenv(InstallationId, config.InstallationId)
	_ = os.Setenv(CommitterName, config.CommitterName)
	_ = os.Setenv(CommitterEmail, config.CommitterEmail)
	_ = os.Setenv(ConfigFile, config.ConfigFile)
	_ = os.Setenv(DeprecationFile, config.DeprecationFile)
	_ = os.Setenv(WebserverPath, config.WebserverPath)
	_ = os.Setenv(SrcBranch, config.SrcBranch)
	_ = os.Setenv(PEMKeyName, config.PEMKeyName)
	_ = os.Setenv(CopierLogName, config.CopierLogName)
	_ = os.Setenv(GoogleCloudProjectId, config.GoogleCloudProjectId)
	_ = os.Setenv(DefaultRecursiveCopy, fmt.Sprintf("%t", config.DefaultRecursiveCopy))
	_ = os.Setenv(DefaultPRMerge, fmt.Sprintf("%t", config.DefaultPRMerge))
	_ = os.Setenv(DefaultCommitMessage, config.DefaultCommitMessage)

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
		AppId:          config.AppId,
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
