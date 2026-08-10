package main

import (
	"log/slog"
	"os"
)

func main() {

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Sets minimum level to DEBUG
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	slog.Info("redis server starting", "addr", ":6379")
	listen()
}
