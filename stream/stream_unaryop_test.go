package stream

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_UnaryOpertors(t *testing.T) {
	tests := []struct {
		name   string
		source api.Emitter
		sink   api.Collector
		stream func(api.Emitter, api.Collector) *Stream
		tester func(snk api.Collector)
	}{
		{
			name:   "Process operator normal with f(x)R",
			source: emitters.Slice([]string{"hello", "world"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).Process(func(s string) string {
					return strings.ToUpper(s)
				}).Into(snk)
				return strm
			},
			tester: func(snk api.Collector) {
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					val := data.(string)
					if val != "HELLO" && val != "WORLD" {
						t.Fatalf("got unexpected value %v of type %T", val, val)
					}
				}
			},
		},

		{
			name:   "Filter operator normal with f(context, x)R",
			source: emitters.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).Filter(func(ctx context.Context, data string) bool {
					return !strings.Contains(data, "O")
				}).Into(snk)
				strm.WithContext(context.Background())
				return strm
			},
			tester: func(snk api.Collector) {
				var result strings.Builder
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					result.WriteString(data.(string))
				}
				if result.String() != "ARE" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},

		{
			name:   "Map operator normal with f(x)R",
			source: emitters.Slice([]string{"HELLO", "WORLD"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).Map(func(data string) int {
					return len(data)
				}).Into(snk)
				return strm
			},
			tester: func(snk api.Collector) {
				count := 0
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					count += data.(int)
				}
				if count != 10 {
					t.Fatal("unexpected data returned by Map operator:", count)
				}
			},
		},

		{
			name:   "FlatMap operator normal with f(context, x)R",
			source: emitters.Slice([]string{"HELLO WORLD", "HOW ARE YOU?"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).FlatMap(func(ctx context.Context, data string) []string {
					return strings.Split(data, " ")
				}).Into(snk)
				strm.WithContext(context.Background())
				return strm
			},
			tester: func(snk api.Collector) {
				count := 0
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					count += len(data.(string))
				}
				if count != 20 {
					t.Fatal("unexpected data returned by FlatMap operator:", count)
				}
			},
		},

		{
			name:   "Operator returning StreamItem with f(x)StreamItem",
			source: emitters.Slice([]string{"hello", "world"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).Process(func(s string) api.StreamItem {
					return api.StreamItem{Item: s}
				}).Into(snk)
				return strm
			},
			tester: func(snk api.Collector) {
				var result strings.Builder
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					result.WriteString(data.(api.StreamItem).Item.(string))
				}
				if result.String() != "helloworld" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},

		{
			name:   "Operator with error f(x)error",
			source: emitters.Slice([]string{"hello", "world"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).Process(func(s string) interface{} {
					if s == "world" {
						return fmt.Errorf("unsupported value: %s", s)
					}
					return s
				}).Into(snk)
				return strm
			},
			tester: func(snk api.Collector) {
				var result strings.Builder
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					result.WriteString(data.(string))
				}
				if result.String() != "hello" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},

		{
			name:   "Operator with StreamError f(x)StreamError",
			source: emitters.Slice([]string{"hello", "world"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).Process(func(s string) interface{} {
					if s == "world" {
						return api.Error("unsupported data")
					}
					return s
				}).Into(snk)
				return strm
			},
			tester: func(snk api.Collector) {
				var result strings.Builder
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					result.WriteString(data.(string))
				}
				if result.String() != "hello" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},

		{
			name:   "Operator with StreamError with Item: f(x)StreamError",
			source: emitters.Slice([]string{"hello", "world"}),
			sink:   collectors.Slice(),
			stream: func(src api.Emitter, snk api.Collector) *Stream {
				strm := New(src).Process(func(s string) interface{} {
					if s == "world" {
						return api.ErrorWithItem("unsupported data", &api.StreamItem{Item: s})
					}
					return s
				}).Into(snk)
				return strm
			},
			tester: func(snk api.Collector) {
				var result strings.Builder
				for _, data := range snk.(*collectors.SliceCollector).Get() {
					switch item := data.(type) {
					case api.StreamItem:
						result.WriteString(item.Item.(string))
					case string:
						result.WriteString(item)
					}
				}
				if result.String() != "helloworld" {
					t.Fatal("unexpected data returned by Filter operator:", result.String())
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			strm := test.stream(test.source, test.sink)
			select {
			case err := <-strm.Open():
				if err != nil {
					t.Fatal(err)
				}
				test.tester(test.sink)
			case <-time.After(50 * time.Millisecond):
				t.Fatal("Waited too long ...")
			}
		})
	}
}
