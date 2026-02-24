package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
)

// Open opens a database connection based on config and runs migrations for SQLite.
func Open(cfg *config.Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch cfg.Database.Driver {
	case "sqlite", "sqlite3":
		dsn := cfg.Database.DSN
		if dsn != ":memory:" {
			dir := filepath.Dir(dsn)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("create sqlite dir: %w", err)
			}
		}
		db, err = sql.Open("sqlite", dsn+"?_foreign_keys=off")
		if err != nil {
			return nil, err
		}
		if err := runSQLiteMigrations(db); err != nil {
			db.Close()
			return nil, err
		}
	case "mysql":
		db, err = sql.Open("mysql", cfg.Database.DSN)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, err
		}
	case "postgres", "pg":
		db, err = sql.Open("postgres", cfg.Database.DSN)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	db.SetMaxOpenConns(10)
	if cfg.Database.Driver == "sqlite" || cfg.Database.Driver == "sqlite3" {
		db.SetMaxOpenConns(1)
	}
	return db, nil
}

func runSQLiteMigrations(db *sql.DB) error {
	const name = "000001_sqlite_init.up.sql"
	paths := []string{
		"migrations/" + name,
		"backend/migrations/" + name,
	}
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	statements := splitSQL(string(data))
	for _, s := range statements {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// 不按 "--" 跳过，否则首段 "注释 + CREATE TABLE" 会被整段跳过导致表未创建
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}
	return nil
}

func splitSQL(sql string) []string {
	var out []string
	for _, s := range strings.Split(sql, ";") {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t+";")
		}
	}
	return out
}
