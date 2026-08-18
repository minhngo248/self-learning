package config

import (
	log "log/slog"
	"os"
)

func InitLogger(profile string) {
	var level log.Level
	var handler log.Handler

	if profile == "dev" {
		level = log.LevelDebug
		handler = log.NewTextHandler(os.Stdout, &log.HandlerOptions{
			Level: level,
		})
	} else {
		level = log.LevelInfo
		handler = log.NewJSONHandler(os.Stdout, &log.HandlerOptions{
			Level: level,
		})
	}

	// Set as the default global logger
	log.SetDefault(log.New(handler))
}
