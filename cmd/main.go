package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sakibcoolz/loki-suite/internal/controller"
	"github.com/sakibcoolz/loki-suite/internal/handler"
	"github.com/sakibcoolz/loki-suite/internal/models"
	"github.com/sakibcoolz/loki-suite/internal/repository"
	"github.com/sakibcoolz/loki-suite/internal/service"
	"github.com/sakibcoolz/zcornor/pkg/config"
	"github.com/sakibcoolz/zcornor/pkg/db/postgres"
	"github.com/sakibcoolz/zcornor/pkg/security"
	"github.com/sakibcoolz/zcornor/pkg/zlog"
	"go.uber.org/zap"
)

func main() {
	config := config.New()

	log, err := zlog.New(zlog.LoggerConfig{
		Environment: config.Env,
		Source:      config.Name,
		Debug:       config.Debug,
	})
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	logger := log

	ctx := context.Background()

	db := postgres.Connect(ctx, log, config.Postgres)
	if db == nil {
		log.FatalSimple("Failed to connect to database")
	}

	// Migrate database schema
	log.Info(ctx, "Starting database migration...")
	if err := db.AutoMigrate(
		&models.WebhookSubscription{},
		&models.WebhookEvent{},
		&models.ExecutionChain{},
		&models.ExecutionChainStep{},
		&models.ExecutionChainRun{},
		&models.ExecutionChainStepRun{},
	); err != nil {
		log.Fatal(ctx, "Failed to migrate database schema", zap.Error(err))
	}
	log.Info(ctx, "Database migration completed successfully")

	// Initialize security service
	securitySvc := security.NewSecurityService(
		config.JWT.JWTSecret,
		config.JWT.HMACKeyLength,
		int(config.JWT.Exp),
	)

	// Initialize repositories
	webhookRepo := repository.NewWebhookRepository(db)
	chainRepo := repository.NewExecutionChainRepository(db)

	// Initialize services
	webhookSvc := service.NewWebhookService(webhookRepo, securitySvc, config)
	chainSvc := service.NewExecutionChainService(chainRepo, webhookRepo, securitySvc, config)

	// Set chain service in webhook service (to avoid circular dependencies)
	webhookSvc.SetChainService(chainSvc)

	// Initialize controllers
	webhookController := controller.NewWebhookController(webhookSvc)
	chainController := controller.NewExecutionChainController(chainSvc)

	// Initialize router
	router := handler.NewRouter(webhookController, chainController)
	router.Setup()

	// Start server
	logger.Info(ctx, "Server starting",
		zap.String("host", config.Host),
		zap.String("port", config.Port))

	// Graceful shutdown
	go func() {
		if err := router.GetEngine().Run(config.Host + ":" + config.Port); err != nil {
			logger.Fatal(ctx, "Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.InfoSimple("Shutting down server...")

	// Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Error(ctx, "Error closing database connection", zap.Error(err))
		}
	}

	logger.Info(ctx, "Server stopped")
}
