package main

import (
	"log"

	"github.com/Dorrrke/g2-books/internal/config"
	"github.com/Dorrrke/g2-books/internal/server"
	"github.com/Dorrrke/g2-books/internal/storage"
)

func main() {
	cfg := config.ReadConfig()
	log.Println(cfg)
	storage := storage.New()
	server := server.New(cfg.Host, storage)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
