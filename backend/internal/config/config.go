package config

import (
	"os"
)

type Config struct {
	Port             string
	DatabaseURL      string // Direct PostgreSQL connection string
	SupabaseURL      string
	SupabaseKey      string
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
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		SupabaseURL:      getEnv("SUPABASE_URL", ""),
		SupabaseKey:      getEnv("SUPABASE_KEY", ""),
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
		ADSkipVerify:   getEnv("AD_SKIP_VERIFY", "true") == "true",
		// EWS Configuration
		EWSURL:           getEnv("EWS_URL", "https://post.ekf.su/EWS/Exchange.asmx"),
		EWSDomain:        getEnv("EWS_DOMAIN", "ekfgroup"),
		EWSUsername:      getEnv("EWS_USERNAME", ""),
		EWSPassword:      getEnv("EWS_PASSWORD", ""),
		EWSSkipTLSVerify: getEnv("EWS_SKIP_TLS_VERIFY", "false") == "true",
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		CamundaURL:       getEnv("CAMUNDA_URL", ""),
		CamundaUser:      getEnv("CAMUNDA_USER", ""),
		CamundaPassword:  getEnv("CAMUNDA_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
