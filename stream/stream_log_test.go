package stream

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/operators/exec"
	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/sources"
)

func TestStreamLog(t *testing.T) {
	t.Run("stream with no log setup", func(t *testing.T) {
		src := sources.Slice([][]string{
			{"request", "/i/a"},
			{"response", "/i/a/", "00:11:51:AA", "served"},
			{"response", "/i/a/", "00:11:51:AA"},
		})

		strm := From(src)
		strm.Run(
			exec.Filter(func(ctx context.Context, in []string) bool {
				return len(in) > 2
			}),
		)
		strm.Into(sinks.Discard())

		select {
		case err := <-strm.Open(context.Background()):
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Took too long")
		}
	})

	t.Run("stream with log setup", func(t *testing.T) {
		src := sources.Slice([][]string{
			{"request", "/i/a"},
			{"response", "/i/a/", "00:11:51:AA", "served"},
			{"response", "/i/a/", "00:11:51:AA"},
		})

		var logCount atomic.Int32
		strm := From(src)
		strm.WithLogSink(sinks.Func(func(log api.StreamLog) error {
			logCount.Add(1)
			return nil
		}))
		strm.Run(
			exec.Filter(func(ctx context.Context, in []string) bool {
				return len(in) > 2
			}),
		)
		strm.Into(sinks.Discard())

		select {
		case err := <-strm.Open(context.Background()):
			if err != nil {
				t.Fatal(err)
			}
			if logCount.Load() == 0 {
				t.Fatal("expecting count 5, got", logCount.Load())
			}
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Took too long")
		}
	})

	t.Run("with log displayed", func(t *testing.T) {
		src := sources.Slice([][]string{
			{"request", "/i/a"},
			{"response", "/i/a/", "00:11:51:AA", "served"},
			{"response", "/i/a/", "00:11:51:AA"},
		})

		strm := From(src)
		strm.WithLogSink(sinks.Func(func(log api.StreamLog) error {
			slog.LogAttrs(context.Background(), log.Level, log.Message, log.Attrs...)
			return nil
		}))
		strm.Run(
			exec.Filter(func(ctx context.Context, in []string) bool {
				return len(in) > 2
			}),
		)
		strm.Into(sinks.Discard())

		select {
		case err := <-strm.Open(context.Background()):
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Took too long")
		}
	})
}
