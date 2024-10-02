package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Dorrrke/g2-books/internal/config"
	authservicev1 "github.com/Dorrrke/g2-books/internal/go"
	"github.com/Dorrrke/g2-books/internal/logger"
	"github.com/Dorrrke/g2-books/internal/server"
	"github.com/Dorrrke/g2-books/internal/storage"
)

func main() {
	cfg := config.ReadConfig()
	log := logger.Get(cfg.Debug)
	log.Debug().Msg("logger was inited")
	log.Debug().Any("config", cfg).Send()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-c
		cancel()
	}()
	var stor server.Storage
	stor, err := storage.NewRepo(context.Background(), cfg.DBDsn)
	if err != nil {
		log.Fatal().Err(err).Msg("init storage failed")
	}
	if err = storage.Migrations(cfg.DBDsn, cfg.MigratePath); err != nil {
		log.Fatal().Err(err).Msg("migrations failed")
	}
	// grpc.Dial("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(cfg.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("creation grpc connection failed")
	}
	defer conn.Close()

	authClien := authservicev1.NewAuthServiceClient(conn)

	server := server.New(cfg.Host, stor, authClien)

	group, gCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		defer log.Debug().Msg("server runner - end")
		log.Info().Msg("server was started")
		if err = server.Run(gCtx); err != nil {
			log.Error().Err(err).Msg("server error")
			return err
		}
		return nil
	})
	group.Go(func() error {
		defer log.Debug().Msg("error chan listener - end")
		return <-server.ErrChan
	})
	group.Go(func() error {
		defer log.Debug().Msg("gCtx listener - end")
		<-gCtx.Done()
		log.Debug().Msg("eGroup: gCtx - Done")
		return server.ShutdownServer(gCtx)
	})

	if err = group.Wait(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("server was stoped")
		} else {
			log.Error().Err(err).Msg("fatal server stop")
		}
	}
}
