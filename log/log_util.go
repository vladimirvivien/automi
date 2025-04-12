package log

import (
	"context"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
)

func NoLogFunc(_ context.Context, _ api.StreamLog) {
	// no op
}

func LogDebug(msg string, attrs ...slog.Attr) api.StreamLog {
	return api.StreamLog{
		Level:   slog.LevelDebug,
		Message: msg,
		Attrs:   attrs,
	}
}

func LogError(msg string, attrs ...slog.Attr) api.StreamLog {
	return api.StreamLog{
		Level:   slog.LevelError,
		Message: msg,
		Attrs:   attrs,
	}
}

func LogInfo(msg string, attrs ...slog.Attr) api.StreamLog {
	return api.StreamLog{
		Level:   slog.LevelInfo,
		Message: msg,
		Attrs:   attrs,
	}
}

func LogWarn(msg string, attrs ...slog.Attr) api.StreamLog {
	return api.StreamLog{
		Level:   slog.LevelWarn,
		Message: msg,
		Attrs:   attrs,
	}
}
