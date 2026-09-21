package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/api"
	"github.com/nwasiq/fieldops/backend/internal/database"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

func main() {
	seedOnly := flag.Bool("seed-only", false, "migrate, seed and exit without serving")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger, *seedOnly); err != nil {
		logger.Error("fatal", "error", err.Error())
		os.Exit(1)
	}
}

func run(logger *slog.Logger, seedOnly bool) error {
	if err := loadDotEnv(".env", "../.env"); err != nil {
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	cipher, err := services.NewCipher(cfg.EncryptionKey)
	if err != nil {
		return err
	}
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	if err := database.Migrate(db); err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	server := api.NewServer(db, api.Config{JWTSecret: cfg.JWTSecret, Cipher: cipher, Logger: logger})

	if seedOnly || cfg.SeedOnBoot {
		if err := database.Seed(context.Background(), db, server.Users, server.Sites, time.Now()); err != nil {
			return fmt.Errorf("seed: %w", err)
		}
		logger.Info("seed complete")
	}
	if seedOnly {
		return nil
	}

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Engine,
		ReadHeaderTimeout: 10 * time.Second,
	}
	errs := make(chan error, 1)
	go func() {
		logger.Info("listening", "port", cfg.Port)
		errs <- httpServer.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errs:
		return err
	case sig := <-stop:
		logger.Info("shutting down", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(ctx)
	}
}
