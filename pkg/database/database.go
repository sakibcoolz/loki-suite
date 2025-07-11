package database

import (
	"fmt"

	"loki-suite/internal/config"
	"loki-suite/internal/models"
	"loki-suite/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Database holds the database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	// Configure GORM logger to use Zap
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Name),
		zap.String("port", cfg.Port))

	return &Database{DB: db}, nil
}

// AutoMigrate runs database migrations
func (d *Database) AutoMigrate() error {
	err := d.DB.AutoMigrate(
		&models.WebhookSubscription{},
		&models.WebhookEvent{},
		&models.ExecutionChain{},
		&models.ExecutionChainStep{},
		&models.ExecutionChainRun{},
		&models.ExecutionChainStepRun{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Health checks the database connection health
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
