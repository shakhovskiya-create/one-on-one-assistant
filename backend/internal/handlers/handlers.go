package handlers

import (
	"github.com/ekf/one-on-one-backend/internal/config"
	"github.com/ekf/one-on-one-backend/internal/database"
	"github.com/ekf/one-on-one-backend/internal/services"
	"github.com/ekf/one-on-one-backend/pkg/ai"
	"github.com/ekf/one-on-one-backend/pkg/ews"
	"github.com/ekf/one-on-one-backend/pkg/telegram"
)

// Handler holds all handler dependencies
type Handler struct {
	Config    *config.Config
	DB        *database.SupabaseClient
	AI        *ai.Client
	EWS       *ews.Client
	Telegram  *telegram.Client
	Connector *services.ConnectorManager
}

// NewHandler creates a new handler with all dependencies
func NewHandler(cfg *config.Config) *Handler {
	var db *database.SupabaseClient
	if cfg.SupabaseURL != "" && cfg.SupabaseKey != "" {
		db = database.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseKey)
	}

	aiClient := ai.NewClient(
		cfg.OpenAIKey,
		cfg.AnthropicKey,
		cfg.YandexAPIKey,
		cfg.YandexFolderID,
	)

	ewsClient := ews.NewClient(cfg.EWSURL, cfg.EWSDomain)
	tgClient := telegram.NewClient(cfg.TelegramBotToken)
	connector := services.NewConnectorManager(cfg.ConnectorAPIKey)

	return &Handler{
		Config:    cfg,
		DB:        db,
		AI:        aiClient,
		EWS:       ewsClient,
		Telegram:  tgClient,
		Connector: connector,
	}
}
