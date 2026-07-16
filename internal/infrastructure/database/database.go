package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"erp/internal/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		driver = "sqlite"
	}
	var dialector gorm.Dialector
	switch driver {
	case "sqlite", "sqlite3":
		dsn := strings.TrimSpace(cfg.DSN)
		if dsn == "" {
			dsn = "data/erp.db?_foreign_keys=on&_busy_timeout=5000"
		}
		if err := ensureSQLiteDir(dsn); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if driver == "sqlite" || driver == "sqlite3" {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}
		sqlDB.SetMaxOpenConns(1)
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			return nil, err
		}
	}
	return db, nil
}

func ensureSQLiteDir(dsn string) error {
	path := strings.TrimPrefix(dsn, "file:")
	if i := strings.IndexAny(path, "?"); i >= 0 {
		path = path[:i]
	}
	path = strings.TrimSpace(path)
	if path == "" || path == ":memory:" || strings.HasPrefix(path, "mode=memory") {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
