package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// config is everything the process reads from its environment. Every
// variable is documented in .env.example; nothing else is read.
type config struct {
	DatabaseURL   string
	JWTSecret     string
	EncryptionKey string
	Port          string
	SeedOnBoot    bool
}

const minJWTSecretLength = 32

func loadConfig() (config, error) {
	cfg := config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		Port:          os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if raw := os.Getenv("SEED_ON_BOOT"); raw != "" {
		seed, err := strconv.ParseBool(raw)
		if err != nil {
			return config{}, fmt.Errorf("SEED_ON_BOOT must be true or false, got %q", raw)
		}
		cfg.SeedOnBoot = seed
	}
	switch {
	case cfg.DatabaseURL == "":
		return config{}, errors.New("DATABASE_URL is required")
	case len(cfg.JWTSecret) < minJWTSecretLength:
		return config{}, fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLength)
	case cfg.EncryptionKey == "":
		return config{}, errors.New("ENCRYPTION_KEY is required")
	}
	return cfg, nil
}

// loadDotEnv reads KEY=VALUE lines from the first file that exists and sets
// any variable the environment does not already define. The repo's .env sits
// one level above backend/, which is where `make backend` runs from.
func loadDotEnv(paths ...string) error {
	for _, path := range paths {
		file, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("open %s: %w", path, err)
		}
		applyErr := applyDotEnv(bufio.NewScanner(file))
		if closeErr := file.Close(); applyErr == nil && closeErr != nil {
			return fmt.Errorf("close %s: %w", path, closeErr)
		}
		return applyErr
	}
	return nil
}

func applyDotEnv(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		value = stripInlineComment(strings.TrimSpace(value))
		if _, set := os.LookupEnv(key); set {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
	}
	return scanner.Err()
}

// stripInlineComment drops a trailing "  # note" from an unquoted value and
// the quotes from a quoted one.
func stripInlineComment(value string) string {
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return value[1 : len(value)-1]
	}
	if i := strings.Index(value, " #"); i >= 0 {
		value = value[:i]
	}
	return strings.TrimSpace(value)
}
