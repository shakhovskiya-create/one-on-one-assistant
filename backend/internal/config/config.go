package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	Port             string
	DatabaseURL      string // PostgreSQL connection string
	OpenAIKey        string
	AnthropicKey     string
	TelegramBotToken string
	YandexAPIKey     string
	YandexFolderID   string
	ConnectorAPIKey  string
	// AD Configuration
	ADURL          string // LDAP URL (e.g., ldap://172.20.0.33:389)
	ADBaseDN       string // Base DN (e.g., OU=EKF-USERS,DC=ekfgroup,DC=ru)
	ADBindUser     string // Bind user for LDAP queries
	ADBindPassword string // Bind password
	ADSkipVerify   bool   // Skip TLS verification
	// EWS Configuration
	EWSURL           string
	EWSDomain        string
	EWSUsername      string // Service account for EWS
	EWSPassword      string // Service account password
	EWSSkipTLSVerify bool   // Only for internal/self-signed certs
	JWTSecret        string // For signing auth tokens
	CamundaURL       string // Camunda BPMN engine URL
	CamundaUser      string // Camunda username
	CamundaPassword  string // Camunda password
	// MinIO Storage Configuration
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinioBucket    string
	MinIOUseSSL    bool
	MinIOPublicURL string // Public URL for media access (e.g., http://localhost/media)
	// External APIs
	GiphyAPIKey string // GIPHY API key for GIF search
	// Confluence Configuration
	ConfluenceURL      string // Confluence Server URL
	ConfluenceUsername string // Confluence username
	ConfluencePassword string // Confluence password
	// GitHub Configuration
	GitHubToken string // GitHub personal access token for API access
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		OpenAIKey:        getEnv("OPENAI_API_KEY", ""),
		AnthropicKey:     getEnv("ANTHROPIC_API_KEY", ""),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		YandexAPIKey:     getEnv("YANDEX_API_KEY", ""),
		YandexFolderID:   getEnv("YANDEX_FOLDER_ID", ""),
		ConnectorAPIKey:  getEnv("CONNECTOR_API_KEY", ""),
		// AD Configuration
		ADURL:          getEnv("AD_URL", "ldap://172.20.0.33:389"),
		ADBaseDN:       getEnv("AD_BASE_DN", "OU=EKF-USERS,DC=ekfgroup,DC=ru"),
		ADBindUser:     getEnv("AD_BIND_USER", ""),
		ADBindPassword: getEnv("AD_BIND_PASSWORD", ""),
		ADSkipVerify:   getEnv("AD_SKIP_VERIFY", "false") == "true", // Default FALSE for security
		// EWS Configuration
		EWSURL:           getEnv("EWS_URL", "https://post.ekf.su/EWS/Exchange.asmx"),
		EWSDomain:        getEnv("EWS_DOMAIN", "ekfgroup"),
		EWSUsername:      getEnv("EWS_USERNAME", ""),
		EWSPassword:      getEnv("EWS_PASSWORD", ""),
		EWSSkipTLSVerify: getEnv("EWS_SKIP_TLS_VERIFY", "false") == "true",
		JWTSecret:        getEnvRequired("JWT_SECRET"), // REQUIRED - no default for security
		CamundaURL:       getEnv("CAMUNDA_URL", ""),
		CamundaUser:      getEnv("CAMUNDA_USER", ""),
		CamundaPassword:  getEnv("CAMUNDA_PASSWORD", ""),
		// MinIO Storage
		MinIOEndpoint:      getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:     getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinioBucket:        getEnv("MINIO_BUCKET", "media"),
		MinIOUseSSL:        getEnv("MINIO_USE_SSL", "false") == "true",
		MinIOPublicURL:     getEnv("MINIO_PUBLIC_URL", "/media"),
		GiphyAPIKey:        getEnv("GIPHY_API_KEY", ""),
		ConfluenceURL:      getEnv("CONFLUENCE_URL", "https://confluence.ekf.su"),
		ConfluenceUsername: getEnv("CONFLUENCE_USERNAME", ""),
		ConfluencePassword: getEnv("CONFLUENCE_PASSWORD", ""),
		GitHubToken:        getEnv("GITHUB_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired returns env var value or panics if not set (for critical security configs)
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("SECURITY ERROR: Required environment variable " + key + " is not set")
	}
	return value
}

// ValidateSecuritySettings logs warnings for insecure configurations
// Should be called after Load() in main.go
func (c *Config) ValidateSecuritySettings() {
	warnings := []string{}

	// Check TLS verification settings
	if c.ADSkipVerify {
		warnings = append(warnings, "AD_SKIP_VERIFY=true: LDAP TLS verification is DISABLED - vulnerable to MITM attacks")
	}

	if c.EWSSkipTLSVerify {
		warnings = append(warnings, "EWS_SKIP_TLS_VERIFY=true: Exchange TLS verification is DISABLED - vulnerable to MITM attacks")
	}

	// Check PostgreSQL sslmode
	if c.DatabaseURL != "" {
		if strings.Contains(c.DatabaseURL, "sslmode=disable") {
			warnings = append(warnings, "DATABASE_URL contains sslmode=disable: Database connection is UNENCRYPTED - use sslmode=require or sslmode=verify-full")
		} else if !strings.Contains(c.DatabaseURL, "sslmode=") {
			warnings = append(warnings, "DATABASE_URL missing sslmode parameter: Consider adding sslmode=require for encrypted connections")
		}
	}

	// Check MinIO SSL
	if !c.MinIOUseSSL && c.MinIOEndpoint != "" {
		warnings = append(warnings, "MINIO_USE_SSL=false: MinIO storage connection is UNENCRYPTED")
	}

	// Check JWT secret strength
	if len(c.JWTSecret) < 32 {
		warnings = append(warnings, "JWT_SECRET is less than 32 characters: Use a stronger secret for production")
	}

	// Log all warnings
	if len(warnings) > 0 {
		log.Println("=== SECURITY CONFIGURATION WARNINGS ===")
		for _, w := range warnings {
			log.Printf("⚠️  SECURITY WARNING: %s", w)
		}
		log.Println("========================================")
	}
}
