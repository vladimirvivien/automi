package sinks

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/log"
)

func TestSlogSink(t *testing.T) {
	t.Run("with slog text", func(t *testing.T) {
		f := SlogText(slog.LevelDebug)
		in := make(chan any)
		go func() {
			in <- log.LogDebug("Hello")
			in <- log.LogInfo("Goodbye")
			in <- log.LogDebug("Hello again")
			close(in)
		}()
		f.SetInput(in)

		select {
		case err := <-f.Open(context.TODO()):
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Waited too long ...")
		}
	})

	t.Run("with slog json", func(t *testing.T) {
		f := SlogJSON(slog.LevelDebug)
		in := make(chan any)
		go func() {
			in <- log.LogDebug("Hello")
			in <- log.LogInfo("Goodbye")
			in <- log.LogDebug("Hello again")
			close(in)
		}()
		f.SetInput(in)

		select {
		case err := <-f.Open(context.TODO()):
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Waited too long ...")
		}
	})
	t.Run("with slog handler", func(t *testing.T) {
		buf := new(bytes.Buffer)
		handler := slog.NewTextHandler(
			buf,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		)
		f := SlogWithHandler(handler)
		in := make(chan any)
		go func() {
			in <- log.LogDebug("Hello")
			in <- log.LogInfo("Goodbye")
			in <- log.LogDebug("Hello again")
			close(in)
		}()
		f.SetInput(in)

		select {
		case err := <-f.Open(context.TODO()):
			if err != nil {
				t.Fatal(err)
			}
			output := buf.String()
			if !strings.Contains(output, "Hello again") {
				t.Error("Log sink not receiving values")
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Waited too long ...")
		}
	})

}
