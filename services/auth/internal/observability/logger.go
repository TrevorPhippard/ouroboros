package observability

import (
	"log/slog"
	"os"
)

func NewLogger(cfg any) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
