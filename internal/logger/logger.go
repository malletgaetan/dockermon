package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

type Config struct {
	Level slog.Level
	// JSON  bool
}

func Initialize(cfg Config) {
	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stdout, opts)

	Log = slog.New(handler)
	slog.SetDefault(Log)
}
