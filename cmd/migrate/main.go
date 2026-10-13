package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hanjo-erp/hanjo/internal/common/database"
	"github.com/hanjo-erp/hanjo/pkg/config"
	"github.com/hanjo-erp/hanjo/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.App.Environment, true)

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.Database, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	log.Info().Msg("running database migrations...")

	// Find migrations folder
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "backend/migrations"
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read migrations directory")
	}

	var upFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	for _, filename := range upFiles {
		log.Info().Str("migration", filename).Msg("applying migration...")
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			log.Fatal().Err(err).Str("file", filename).Msg("failed to read migration file")
		}

		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			log.Fatal().Err(err).Str("file", filename).Msg("failed to execute migration")
		}
		log.Info().Str("migration", filename).Msg("migration applied successfully")
	}

	fmt.Println("\nAll migrations applied successfully!")
}
