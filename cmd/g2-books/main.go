package main

import (
	"context"

	"github.com/Dorrrke/g2-books/internal/config"
	"github.com/Dorrrke/g2-books/internal/logger"
	"github.com/Dorrrke/g2-books/internal/server"
	"github.com/Dorrrke/g2-books/internal/storage"
)

func main() {
	cfg := config.ReadConfig()
	log := logger.Get(cfg.Debug)
	log.Debug().Msg("logger was inited")
	log.Debug().Any("config", cfg).Send()
	var stor server.Storage
	stor, err := storage.NewRepo(context.Background(), cfg.DbDsn)
	if err != nil {
		log.Fatal().Err(err).Msg("init storage failed")
	}
	if err = storage.Migrations(cfg.DbDsn, cfg.MigratePath); err != nil {
		log.Fatal().Err(err).Msg("migrations failed")
	}

	server := server.New(cfg.Host, stor)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
