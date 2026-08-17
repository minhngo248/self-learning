package utils

import (
	log "log/slog"
	"os"
)

func Fatal(msg string, args ...any) {
	log.Error(msg, args...)
	os.Exit(1)
}
