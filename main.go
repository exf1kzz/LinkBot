package main

import (
	tgClient "LinkBot/clients/telegram"
	event_consumer "LinkBot/consumer/event-consumer"
	"LinkBot/events/telegram"
	"LinkBot/storage/sqlite"
	"context"
	"flag"
	"log"
	"os"
)

const (
	defaultTgBotHost         = "api.telegram.org"
	defaultSQLiteStoragePath = "data/sqlite/storage.db"
	batchSize                = 100
)

func main() {
	cfg := mustConfig()

	s, err := sqlite.New(cfg.storagePath)
	if err != nil {
		log.Fatalf("can't create sqlite storage: %v", err)
	}

	if err := s.Init(context.TODO()); err != nil {
		log.Fatalf("can't init sqlite storage: %v", err)
	}

	eventsProcessor := telegram.New(
		tgClient.New(cfg.tgBotHost, cfg.token),
		s)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service stopped", err)
	}
}

type config struct {
	token       string
	tgBotHost   string
	storagePath string
}

func mustConfig() config {
	token := flag.String(
		"token",
		"",
		"Telegram bot token",
	)
	legacyToken := flag.String(
		"token-bot",
		"",
		"Telegram bot token (deprecated, use -token or TG_BOT_TOKEN)",
	)
	storagePath := flag.String(
		"storage-path",
		getEnv("STORAGE_PATH", defaultSQLiteStoragePath),
		"path to SQLite database file",
	)
	tgBotHost := flag.String(
		"tg-host",
		getEnv("TG_BOT_HOST", defaultTgBotHost),
		"Telegram API host",
	)

	flag.Parse()

	selectedToken := firstNonEmpty(*token, *legacyToken, os.Getenv("TG_BOT_TOKEN"), os.Getenv("TELEGRAM_BOT_TOKEN"))
	if selectedToken == "" {
		log.Fatal("token is not specified")
	}

	return config{
		token:       selectedToken,
		tgBotHost:   *tgBotHost,
		storagePath: *storagePath,
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
