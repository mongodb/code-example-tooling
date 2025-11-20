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
	ConfigRepoName       string // Repository where config file is stored
	ConfigRepoOwner      string // Owner of repository where config file is stored
	AppId                string
	AppClientId          string
	InstallationId       string
	CommitterName        string
	CommitterEmail       string
	ConfigFile           string
	MainConfigFile       string // Main config file with workflow references (optional)
	UseMainConfig        bool   // Whether to use main config format
	DeprecationFile      string
	WebserverPath        string
	ConfigRepoBranch     string // Branch to fetch config file from
	PEMKeyName           string
	WebhookSecretName    string
	WebhookSecret        string
	CopierLogName        string
	GoogleCloudProjectId string
	DefaultRecursiveCopy bool
	DefaultPRMerge       bool
	DefaultCommitMessage string

	// Optional features
	DryRun             bool
	AuditEnabled       bool
	MongoURI           string
	MongoURISecretName string
	AuditDatabase      string
	AuditCollection    string
	MetricsEnabled     bool

	// Slack notifications
	SlackWebhookURL string
	SlackChannel    string
	SlackUsername   string
	SlackIconEmoji  string
	SlackEnabled    bool

	// GitHub API retry configuration
	GitHubAPIMaxRetries        int
	GitHubAPIInitialRetryDelay int // in milliseconds

	// PR merge polling configuration
	PRMergePollMaxAttempts int
	PRMergePollInterval    int // in milliseconds
}

const (
	EnvFile                    = "ENV"
	Port                       = "PORT"
	ConfigRepoName             = "CONFIG_REPO_NAME"
	ConfigRepoOwner            = "CONFIG_REPO_OWNER"
	AppId                      = "GITHUB_APP_ID"
	AppClientId                = "GITHUB_APP_CLIENT_ID"
	InstallationId             = "INSTALLATION_ID"
	CommitterName              = "COMMITTER_NAME"
	CommitterEmail             = "COMMITTER_EMAIL"
	ConfigFile                 = "CONFIG_FILE"
	MainConfigFile             = "MAIN_CONFIG_FILE"
	UseMainConfig              = "USE_MAIN_CONFIG"
	DeprecationFile            = "DEPRECATION_FILE"
	WebserverPath              = "WEBSERVER_PATH"
	ConfigRepoBranch           = "CONFIG_REPO_BRANCH"
	PEMKeyName                 = "PEM_NAME"
	WebhookSecretName          = "WEBHOOK_SECRET_NAME"
	WebhookSecret              = "WEBHOOK_SECRET"
	CopierLogName              = "COPIER_LOG_NAME"
	GoogleCloudProjectId       = "GOOGLE_CLOUD_PROJECT_ID"
	DefaultRecursiveCopy       = "DEFAULT_RECURSIVE_COPY"
	DefaultPRMerge             = "DEFAULT_PR_MERGE"
	DefaultCommitMessage       = "DEFAULT_COMMIT_MESSAGE"
	DryRun                     = "DRY_RUN"
	AuditEnabled               = "AUDIT_ENABLED"
	MongoURI                   = "MONGO_URI"
	MongoURISecretName         = "MONGO_URI_SECRET_NAME"
	AuditDatabase              = "AUDIT_DATABASE"
	AuditCollection            = "AUDIT_COLLECTION"
	MetricsEnabled             = "METRICS_ENABLED"
	SlackWebhookURL            = "SLACK_WEBHOOK_URL"
	SlackChannel               = "SLACK_CHANNEL"
	SlackUsername              = "SLACK_USERNAME"
	SlackIconEmoji             = "SLACK_ICON_EMOJI"
	SlackEnabled               = "SLACK_ENABLED"
	GitHubAPIMaxRetries        = "GITHUB_API_MAX_RETRIES"
	GitHubAPIInitialRetryDelay = "GITHUB_API_INITIAL_RETRY_DELAY"
	PRMergePollMaxAttempts     = "PR_MERGE_POLL_MAX_ATTEMPTS"
	PRMergePollInterval        = "PR_MERGE_POLL_INTERVAL"
)

// NewConfig returns a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Port:                       "8080",
		CommitterName:              "Copier Bot",
		CommitterEmail:             "bot@example.com",
		ConfigFile:                 "copier-config.yaml",
		DeprecationFile:            "deprecated_examples.json",
		WebserverPath:              "/webhook",
		ConfigRepoBranch:           "main",                                                           // Default branch to fetch config file from
		PEMKeyName:                 "projects/1054147886816/secrets/CODE_COPIER_PEM/versions/latest", // default secret name for GCP Secret Manager
		WebhookSecretName:          "projects/1054147886816/secrets/webhook-secret/versions/latest",  // default webhook secret name for GCP Secret Manager
		CopierLogName:              "copy-copier-log",                                                // default log name for logging to GCP
		GoogleCloudProjectId:       "github-copy-code-examples",                                      // default project ID for logging to GCP
		DefaultRecursiveCopy:       true,                                                             // system-wide default for recursive copying that individual config entries can override.
		DefaultPRMerge:             false,                                                            // system-wide default for PR merge without review that individual config entries can override.
		DefaultCommitMessage:       "Automated PR with updated examples",                             // default commit message used when per-config commit_message is absent.
		GitHubAPIMaxRetries:        3,                                                                // default number of retry attempts for GitHub API calls
		GitHubAPIInitialRetryDelay: 500,                                                              // default initial retry delay in milliseconds (exponential backoff)
		PRMergePollMaxAttempts:     20,                                                               // default max attempts to poll PR for mergeability (~10 seconds with 500ms interval)
		PRMergePollInterval:        500,                                                              // default polling interval in milliseconds
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
	config.ConfigRepoName = os.Getenv(ConfigRepoName)
	config.ConfigRepoOwner = os.Getenv(ConfigRepoOwner)
	config.AppId = os.Getenv(AppId)
	config.AppClientId = os.Getenv(AppClientId)
	config.InstallationId = os.Getenv(InstallationId)
	config.CommitterName = getEnvWithDefault(CommitterName, config.CommitterName)
	config.CommitterEmail = getEnvWithDefault(CommitterEmail, config.CommitterEmail)
	config.ConfigFile = getEnvWithDefault(ConfigFile, config.ConfigFile)
	config.MainConfigFile = os.Getenv(MainConfigFile)
	config.UseMainConfig = getBoolEnvWithDefault(UseMainConfig, config.MainConfigFile != "")
	config.DeprecationFile = getEnvWithDefault(DeprecationFile, config.DeprecationFile)
	config.WebserverPath = getEnvWithDefault(WebserverPath, config.WebserverPath)
	config.ConfigRepoBranch = getEnvWithDefault(ConfigRepoBranch, config.ConfigRepoBranch)
	config.PEMKeyName = getEnvWithDefault(PEMKeyName, config.PEMKeyName)
	config.WebhookSecretName = getEnvWithDefault(WebhookSecretName, config.WebhookSecretName)
	config.WebhookSecret = os.Getenv(WebhookSecret)
	config.DefaultRecursiveCopy = getBoolEnvWithDefault(DefaultRecursiveCopy, config.DefaultRecursiveCopy)
	config.DefaultPRMerge = getBoolEnvWithDefault(DefaultPRMerge, config.DefaultPRMerge)
	config.CopierLogName = getEnvWithDefault(CopierLogName, config.CopierLogName)
	config.GoogleCloudProjectId = getEnvWithDefault(GoogleCloudProjectId, config.GoogleCloudProjectId)
	config.DefaultCommitMessage = getEnvWithDefault(DefaultCommitMessage, config.DefaultCommitMessage)

	// Optional features
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

	// GitHub API retry configuration
	config.GitHubAPIMaxRetries = getIntEnvWithDefault(GitHubAPIMaxRetries, config.GitHubAPIMaxRetries)
	config.GitHubAPIInitialRetryDelay = getIntEnvWithDefault(GitHubAPIInitialRetryDelay, config.GitHubAPIInitialRetryDelay)

	// PR merge polling configuration
	config.PRMergePollMaxAttempts = getIntEnvWithDefault(PRMergePollMaxAttempts, config.PRMergePollMaxAttempts)
	config.PRMergePollInterval = getIntEnvWithDefault(PRMergePollInterval, config.PRMergePollInterval)

	// Export resolved values back into environment so downstream os.Getenv sees defaults
	_ = os.Setenv(Port, config.Port)
	_ = os.Setenv(ConfigRepoName, config.ConfigRepoName)
	_ = os.Setenv(ConfigRepoOwner, config.ConfigRepoOwner)
	_ = os.Setenv(AppId, config.AppId)
	_ = os.Setenv(AppClientId, config.AppClientId)
	_ = os.Setenv(InstallationId, config.InstallationId)
	_ = os.Setenv(CommitterName, config.CommitterName)
	_ = os.Setenv(CommitterEmail, config.CommitterEmail)
	_ = os.Setenv(ConfigFile, config.ConfigFile)
	_ = os.Setenv(MainConfigFile, config.MainConfigFile)
	_ = os.Setenv(UseMainConfig, fmt.Sprintf("%t", config.UseMainConfig))
	_ = os.Setenv(DeprecationFile, config.DeprecationFile)
	_ = os.Setenv(WebserverPath, config.WebserverPath)
	_ = os.Setenv(ConfigRepoBranch, config.ConfigRepoBranch)
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

// getIntEnvWithDefault returns the integer environment variable value or default if not set
func getIntEnvWithDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var intValue int
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		return defaultValue
	}
	return intValue
}

// validateConfig checks if all required configuration values are set
func validateConfig(config *Config) error {
	var missingVars []string

	requiredVars := map[string]string{
		ConfigRepoName:  config.ConfigRepoName,
		ConfigRepoOwner: config.ConfigRepoOwner,
		AppId:           config.AppId,
		InstallationId:  config.InstallationId,
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
