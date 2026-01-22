package handlers

import (
	"github.com/ekf/one-on-one-backend/internal/ad"
	"github.com/ekf/one-on-one-backend/internal/config"
	"github.com/ekf/one-on-one-backend/internal/database"
	"github.com/ekf/one-on-one-backend/internal/ews"
	"github.com/ekf/one-on-one-backend/internal/services"
	"github.com/ekf/one-on-one-backend/internal/services/confluence"
	"github.com/ekf/one-on-one-backend/internal/services/github"
	"github.com/ekf/one-on-one-backend/internal/storage"
	"github.com/ekf/one-on-one-backend/pkg/ai"
	"github.com/ekf/one-on-one-backend/pkg/auth"
	"github.com/ekf/one-on-one-backend/pkg/camunda"
	"github.com/ekf/one-on-one-backend/pkg/telegram"
)

// Handler holds all handler dependencies
type Handler struct {
	Config     *config.Config
	DB         database.DBClient
	Storage    *storage.MinIOClient
	AI         *ai.Client
	AD         *ad.Client
	EWS        *ews.Client
	Telegram   *telegram.Client
	Connector  *services.ConnectorManager
	JWT        *auth.JWTManager
	Camunda    *camunda.Client
	Confluence *confluence.Client
	GitHub     *github.Client
}

// NewHandler creates a new handler with all dependencies
func NewHandler(cfg *config.Config) *Handler {
	var db database.DBClient

	// Connect to PostgreSQL database
	if cfg.DatabaseURL != "" {
		pgClient, err := database.NewPostgresClient(cfg.DatabaseURL)
		if err != nil {
			// Log error but continue - some handlers may not need DB
			println("Warning: Failed to connect to database:", err.Error())
		} else {
			db = pgClient
		}
	}

	aiClient := ai.NewClient(
		cfg.OpenAIKey,
		cfg.AnthropicKey,
		cfg.YandexAPIKey,
		cfg.YandexFolderID,
	)

	// Initialize AD and EWS clients for direct access (no connector needed)
	adClient := ad.NewClient(cfg.ADURL, cfg.ADBaseDN, cfg.ADBindUser, cfg.ADBindPassword, cfg.ADSkipVerify)
	ewsClient := ews.NewClient(cfg.EWSURL, cfg.EWSDomain, cfg.EWSSkipTLSVerify)

	tgClient := telegram.NewClient(cfg.TelegramBotToken)
	connector := services.NewConnectorManager(cfg.ConnectorAPIKey)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 24) // 24 hours expiration
	camundaClient := camunda.NewClient(cfg.CamundaURL, cfg.CamundaUser, cfg.CamundaPassword)

	// Initialize MinIO storage client
	var storageClient *storage.MinIOClient
	if cfg.MinIOEndpoint != "" {
		minioClient, err := storage.NewMinIOClient(storage.MinIOConfig{
			Endpoint:  cfg.MinIOEndpoint,
			AccessKey: cfg.MinIOAccessKey,
			SecretKey: cfg.MinIOSecretKey,
			Bucket:    cfg.MinioBucket,
			UseSSL:    cfg.MinIOUseSSL,
			PublicURL: cfg.MinIOPublicURL,
		})
		if err != nil {
			println("Warning: Failed to connect to MinIO storage:", err.Error())
		} else {
			storageClient = minioClient
		}
	}

	// Initialize Confluence client
	confluenceClient := confluence.NewClient(cfg.ConfluenceURL, cfg.ConfluenceUsername, cfg.ConfluencePassword)

	// Initialize GitHub client
	githubClient := github.NewClient(cfg.GitHubToken)

	return &Handler{
		Config:     cfg,
		DB:         db,
		Storage:    storageClient,
		AI:         aiClient,
		AD:         adClient,
		EWS:        ewsClient,
		Telegram:   tgClient,
		Connector:  connector,
		JWT:        jwtManager,
		Camunda:    camundaClient,
		Confluence: confluenceClient,
		GitHub:     githubClient,
	}
}
