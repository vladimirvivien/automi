package stream

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/operators/exec"
	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/sources"
	"github.com/vladimirvivien/automi/testutil"
)

func TestStreamWithOpertors(t *testing.T) {
	tests := []struct {
		name   string
		stream func(*testing.T) *Stream
		tester func(*testing.T, *Stream)
	}{
		{
			name: "stream with map",
			stream: func(t *testing.T) *Stream {
				source := sources.Slice([]string{"hello", "world"})
				sink := sinks.Slice[string]()

				strm := From(source).Run(
					exec.Map[string, string](func(ctx context.Context, in string) string {
						return strings.ToUpper(in)
					}),
				)
				strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
				strm.Into(sink)
				return strm
			},
			tester: func(t *testing.T, strm *Stream) {
				snk := strm.GetSink()
				for _, data := range snk.(*sinks.SliceSink[string, []string]).Get() {
					val := data
					if val != "HELLO" && val != "WORLD" {
						t.Fatalf("got unexpected value %v of type %T", val, val)
					}
				}
			},
		},
		{
			name: "stream with filter",
			stream: func(t *testing.T) *Stream {
				source := sources.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
				sink := sinks.Slice[string]()

				strm := From(source).Run(
					exec.Filter[string](func(ctx context.Context, in string) bool {
						return !strings.Contains(in, "O")
					}),
				)
				strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
				strm.Into(sink)
				return strm
			},
			tester: func(t *testing.T, strm *Stream) {
				snk := strm.GetSink()
				var result strings.Builder
				for _, data := range snk.(*sinks.SliceSink[string, []string]).Get() {
					result.WriteString(data)
				}
				if result.String() != "ARE" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},
		{
			name: "multiple ops with StreamItem",
			stream: func(t *testing.T) *Stream {
				source := sources.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
				sink := sinks.Slice[api.StreamItem[string]]()

				strm := From(source).Run(
					exec.Filter(func(ctx context.Context, in string) bool {
						return !strings.Contains(in, "O")
					}),
					exec.Execute(func(ctx context.Context, data string) api.StreamItem[string] {
						return api.StreamItem[string]{Item: data}
					}),
				)
				strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
				strm.Into(sink)
				return strm
			},
			tester: func(t *testing.T, strm *Stream) {
				snk := strm.GetSink()
				var result strings.Builder
				for _, data := range snk.(*sinks.SliceSink[api.StreamItem[string], []api.StreamItem[string]]).Get() {
					result.WriteString(data.Item)
				}
				if result.String() != "ARE" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},
		{
			name: "error handling with error",
			stream: func(t *testing.T) *Stream {
				source := sources.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
				sink := sinks.Slice[api.StreamItem[string]]()

				strm := From(source).Run(
					exec.Execute[string, any](func(ctx context.Context, in string) any {
						if strings.Contains(in, "O") {
							return errors.New("unsupported word")
						}
						return in
					}),
					exec.Execute(func(ctx context.Context, data string) api.StreamItem[string] {
						return api.StreamItem[string]{Item: data}
					}),
				)
				strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
				strm.Into(sink)
				return strm
			},
			tester: func(t *testing.T, strm *Stream) {
				snk := strm.GetSink()
				var result strings.Builder
				for _, data := range snk.(*sinks.SliceSink[api.StreamItem[string], []api.StreamItem[string]]).Get() {
					result.WriteString(data.Item)
				}
				if result.String() != "ARE" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},
		{
			name: "error handling with StreamResult",
			stream: func(t *testing.T) *Stream {
				source := sources.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
				sink := sinks.Slice[api.StreamItem[string]]()

				strm := From(source).Run(
					exec.Execute[string, api.StreamResult](func(ctx context.Context, in string) api.StreamResult {
						if strings.Contains(in, "O") {
							return api.StreamResult{Err: errors.New("unsupported word"), Action: api.ActionSkipItem}
						}
						return api.StreamResult{Value: in}
					}),
					exec.Execute(func(ctx context.Context, data string) api.StreamItem[string] {
						return api.StreamItem[string]{Item: data}
					}),
				)
				strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
				strm.Into(sink)
				return strm
			},
			tester: func(t *testing.T, strm *Stream) {
				snk := strm.GetSink()
				var result strings.Builder
				for _, data := range snk.(*sinks.SliceSink[api.StreamItem[string], []api.StreamItem[string]]).Get() {
					result.WriteString(data.Item)
				}
				if result.String() != "ARE" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			strm := test.stream(t)
			select {
			case err := <-strm.Open(context.Background()):
				if err != nil {
					t.Fatal(err)
				}
				test.tester(t, strm)
			case <-time.After(50 * time.Millisecond):
				t.Fatal("Waited too long ...")
			}
		})
	}
}

func TestStreamSynchronization(t *testing.T) {
	t.Run("with long running source", func(t *testing.T) {
		// TODO: See long-running synchronization task
	})
}
