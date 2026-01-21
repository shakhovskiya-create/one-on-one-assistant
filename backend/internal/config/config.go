package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string // PostgreSQL connection string
	OpenAIKey   string
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
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""),
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
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinioBucket:    getEnv("MINIO_BUCKET", "media"),
		MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		MinIOPublicURL: getEnv("MINIO_PUBLIC_URL", "/media"),
		GiphyAPIKey:    getEnv("GIPHY_API_KEY", ""),
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
