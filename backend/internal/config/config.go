package config

import (
	"os"
)

type Config struct {
	Port               string
	SupabaseURL        string
	SupabaseKey        string
	OpenAIKey          string
	AnthropicKey       string
	TelegramBotToken   string
	YandexAPIKey       string
	YandexFolderID     string
	ConnectorAPIKey    string
	EWSURL             string
	EWSDomain          string
	EWSSkipTLSVerify   bool   // Only for internal/self-signed certs
	JWTSecret          string // For signing auth tokens
	CamundaURL         string // Camunda BPMN engine URL
	CamundaUser        string // Camunda username
	CamundaPassword    string // Camunda password
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseKey:       getEnv("SUPABASE_KEY", ""),
		OpenAIKey:         getEnv("OPENAI_API_KEY", ""),
		AnthropicKey:      getEnv("ANTHROPIC_API_KEY", ""),
		TelegramBotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		YandexAPIKey:      getEnv("YANDEX_API_KEY", ""),
		YandexFolderID:    getEnv("YANDEX_FOLDER_ID", ""),
		ConnectorAPIKey:   getEnv("CONNECTOR_API_KEY", ""),
		EWSURL:            getEnv("EWS_URL", "https://post.ekf.su/EWS/Exchange.asmx"),
		EWSDomain:         getEnv("EWS_DOMAIN", "ekfgroup"),
		EWSSkipTLSVerify:  getEnv("EWS_SKIP_TLS_VERIFY", "false") == "true",
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		CamundaURL:        getEnv("CAMUNDA_URL", ""),
		CamundaUser:       getEnv("CAMUNDA_USER", ""),
		CamundaPassword:   getEnv("CAMUNDA_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
