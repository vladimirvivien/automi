package sinks

import (
	"context"
	"log/slog"
	"os"

	"github.com/vladimirvivien/automi/api"
)

// SlogWithHandler builds a FuncSink to receive api.StreamLog items.
// Internally, the function emits items as log events using slog.
func SlogWithHandler(handler slog.Handler) *FuncSink[api.StreamLog] {
	ctx := context.TODO()
	logger := slog.New(handler)
	return Func(func(log api.StreamLog) error {
		logger.LogAttrs(ctx, log.Level, log.Message, log.Attrs...)
		return nil
	})
}

// SlogJSON builds a FuncSink to receive api.StreamLog items.
// Internally, the function constructs a slog.JSONHandler
// to emit items as log events.
func SlogJSON(level slog.Level) *FuncSink[api.StreamLog] {
	handler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: level,
		},
	)
	return SlogWithHandler(handler)
}

// SlogText builds a FuncSink to receive api.StreamLog items.
// Internally, the function constructs a slog.TextHandler
// to emit items as log events.
func SlogText(level slog.Level) *FuncSink[api.StreamLog] {
	handler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: level,
		},
	)
	return SlogWithHandler(handler)
}
