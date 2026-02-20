// import-mysql reads a MySQL dump (e.g. xxjz.sql) and imports into the configured database (SQLite by default).
package main

import (
	"log"
	"os"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/importsql"
	"github.com/xiaoxinpro/xxjz-go-web/backend/pkg/db"
)

func main() {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = "../config.yaml"
		}
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	filePath := os.Getenv("FILE")
	if filePath == "" && len(os.Args) >= 2 {
		filePath = os.Args[1]
	}
	if filePath == "" {
		log.Fatal("usage: import-mysql <path-to-xxjz.sql> or set FILE=path")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("read file: %v", err)
	}

	if cfg.Database.Driver != "sqlite" && cfg.Database.Driver != "sqlite3" {
		log.Fatalf("import-mysql currently only supports target driver sqlite; got %s", cfg.Database.Driver)
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	statements := importsql.MySQLToSQLite(string(data))
	log.Printf("converted %d statements", len(statements))
	if err := importsql.RunSQLiteStatements(database, statements); err != nil {
		log.Fatalf("run statements: %v", err)
	}
	log.Println("import done")
}
