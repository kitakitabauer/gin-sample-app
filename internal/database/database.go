package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// Config はデータベース接続に必要な設定値です。
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Open は driver/DSN を元に *sql.DB を初期化します。
func Open(ctx context.Context, cfg Config) (*sql.DB, error) {
	driverName, err := normalizeDriver(cfg.Driver)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// normalizeDriver は設定値を sql.Open に渡す実ドライバ名へ正規化します。
func normalizeDriver(name string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "sqlite", "sqlite3":
		return "sqlite", nil
	case "postgres", "postgresql", "pgx":
		return "pgx", nil
	default:
		if strings.TrimSpace(name) == "" {
			return "", fmt.Errorf("database driver is required")
		}
		return name, nil
	}
}
