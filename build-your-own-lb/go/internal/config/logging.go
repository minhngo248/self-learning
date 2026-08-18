package config

import (
	"context"
	log "log/slog"
	"os"
	"runtime"
)

type goroutineHandler struct {
	handler log.Handler
}

func (h goroutineHandler) Enabled(ctx context.Context, level log.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h goroutineHandler) Handle(ctx context.Context, record log.Record) error {
	record.AddAttrs(log.Int("goroutines", runtime.NumGoroutine()))
	return h.handler.Handle(ctx, record)
}

func (h goroutineHandler) WithAttrs(attrs []log.Attr) log.Handler {
	return goroutineHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h goroutineHandler) WithGroup(name string) log.Handler {
	return goroutineHandler{handler: h.handler.WithGroup(name)}
}

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

	// Set as the default global logger with goroutine count in every record.
	log.SetDefault(log.New(goroutineHandler{handler: handler}))
}
