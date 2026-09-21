package database

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

// gormConfig keeps GORM quiet (the request log is the log) and stamps rows in
// UTC so a timestamp never carries the host's zone.
func gormConfig() *gorm.Config {
	return &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time { return time.Now().UTC() },
	}
}

// Connect opens the PostgreSQL database named by a DATABASE_URL.
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig())
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

// ConnectSQLite opens an in-memory SQLite database for tests. The name keeps
// one test's database apart from another's; the shared cache lets every
// connection in the pool see the same data.
func ConnectSQLite(name string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=foreign_keys(1)", name)
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig())
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	return db, nil
}

// Migrate creates or updates every table.
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Site{},
		&models.Visit{},
		&models.ClockEvent{},
		&models.AuditLog{},
	)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
