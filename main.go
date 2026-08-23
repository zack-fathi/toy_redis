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

	db := &database{
		Store:  make(map[string]string),
		Hashes: make(map[string]map[string]string),
	}

	aof := &aof{}
	err := aof.openAOF("database.aof")
	if err != nil {
		slog.Error("failed to initialize AOF", "error", err)
		return
	}
	defer aof.close()
	err = aof.replay(db)
	if err != nil {
		slog.Error("failed to restore database from AOF", "error", err)
		return
	}

	listen(db, aof)
}
