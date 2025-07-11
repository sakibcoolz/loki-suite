package main

import (
	"os"
	"os/signal"
	"syscall"

	"loki-suite/internal/config"
	"loki-suite/internal/controller"
	"loki-suite/internal/handler"
	"loki-suite/internal/repository"
	"loki-suite/internal/service"
	"loki-suite/pkg/database"
	"loki-suite/pkg/logger"
	"loki-suite/pkg/security"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		panic("Invalid configuration: " + err.Error())
	}

	// Initialize logger
	if err := logger.Initialize(cfg.Logger.Level, cfg.Logger.Encoding, cfg.Logger.OutputPath); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting Loki Suite Webhook Service v2.0",
		zap.String("environment", cfg.Server.GinMode),
		zap.String("port", cfg.Server.Port))

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize security service
	securitySvc := security.NewSecurityService(
		cfg.Security.JWTSecret,
		cfg.Security.HMACKeyLength,
		cfg.Security.JWTTokenExpiration,
	)

	// Initialize repositories
	webhookRepo := repository.NewWebhookRepository(db.GetDB())
	chainRepo := repository.NewExecutionChainRepository(db.GetDB())

	// Initialize services
	webhookSvc := service.NewWebhookService(webhookRepo, securitySvc, cfg)
	chainSvc := service.NewExecutionChainService(chainRepo, webhookRepo, securitySvc, cfg)

	// Set chain service in webhook service (to avoid circular dependencies)
	webhookSvc.SetChainService(chainSvc)

	// Initialize controllers
	webhookController := controller.NewWebhookController(webhookSvc)
	chainController := controller.NewExecutionChainController(chainSvc)

	// Initialize router
	router := handler.NewRouter(webhookController, chainController)
	router.Setup()

	// Start server
	logger.Info("Server starting",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port))

	// Graceful shutdown
	go func() {
		if err := router.GetEngine().Run(cfg.Server.Host + ":" + cfg.Server.Port); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Close database connection
	if err := db.Close(); err != nil {
		logger.Error("Error closing database connection", zap.Error(err))
	}

	logger.Info("Server stopped")
}
