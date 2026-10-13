package main

import (
	"context"

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

	log.Info().Msg("seeder ready - waiting for new domain models matching Frontend")
}
