package api

import (
	"context"
	"log/slog"
)

// StreamLog represents stream logging events during stream runtime.
// Logging events are captured using the `slog` package log event API.
type StreamLog struct {
	Level   slog.Level
	Message string
	Attrs   []slog.Attr
}

// StreamLogFunc defines a function that is called to propagate `slog` logging
// events from stream nodes to the stream orchestrator.
type StreamLogFunc func(context.Context, StreamLog)

// Reporter is a stream node (source, operator, or sink) that can report its stream
// activities to the stream orchestrator for side channel consumption.
type Reporter interface {
	SetLogFunc(StreamLogFunc)
}
