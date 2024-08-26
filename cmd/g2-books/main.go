package main

import (
	"context"
	"log"

	"github.com/Dorrrke/g2-books/internal/config"
	"github.com/Dorrrke/g2-books/internal/server"
	"github.com/Dorrrke/g2-books/internal/storage"
)

func main() {
	cfg := config.ReadConfig()
	log.Println(cfg)
	var stor server.Storage
	stor, err := storage.NewRepo(context.Background(), cfg.DbDsn)
	if err != nil {
		log.Fatal(err.Error())
	}
	if err = storage.Migrations(cfg.DbDsn, cfg.MigratePath); err != nil {
		log.Fatal(err.Error())
	}

	server := server.New(cfg.Host, stor)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
