package exec

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/funcs"
	"github.com/vladimirvivien/automi/testutil"
)

func TestNewExecOperator(t *testing.T) {
	op := New(func(ctx context.Context, item any) any {
		return nil
	})

	if op.GetOutput() == nil {
		t.Fatal("default output should not be nil")
	}

	if op.concurrency != 1 {
		t.Fatal("Concurrency should be initialized to 1.")
	}

	op.SetConcurrency(3)
	if op.concurrency != 3 {
		t.Fatal("unexpected concurrency value set:", op.concurrency)
	}

	op.SetInput(make(chan interface{}))
	if op.input == nil {
		t.Fatal("unexpected input set:", op.input)
	}
}

type opExecTest[IN, OUT any] struct {
	name     string
	data     func() <-chan interface{}
	execFunc funcs.ExecFunction[IN, OUT]
	tester   func(*testing.T, <-chan interface{})
}

func runOpExecTest[IN, OUT any](test opExecTest[IN, OUT], t *testing.T) {
	t.Run(test.name, func(t *testing.T) {
		o := New(test.execFunc)
		o.SetInput(test.data())

		wait := make(chan struct{})
		tc := test
		go func() {
			defer close(wait)
			tc.tester(t, o.GetOutput())
		}()

		if err := o.Exec(context.TODO()); err != nil {
			t.Fatal(err)
		}

		select {
		case <-wait:
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Took too long...")
		}
	})
}

func TestExecOperator(t *testing.T) {
	t.Run("normal proc", func(t *testing.T) {
		test := opExecTest[[]string, int]{
			name: "normal processing",
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- []string{"A", "B", "C"}
					in <- []string{"D", "E"}
					in <- []string{"G"}
					close(in)
				}()
				return in
			},
			execFunc: funcs.ExecFunction[[]string, int](func(ctx context.Context, data []string) int {
				return len(data)
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				for data := range out {
					val := data.(int)
					if val != 3 && val != 2 && val != 1 {
						t.Fatalf("Expecting values 3, 2, or 1, but got %d", val)
					}
				}
			},
		}

		runOpExecTest[[]string, int](test, t)
	})

	t.Run("return streamitem", func(t *testing.T) {
		test := opExecTest[int, api.StreamItem[int]]{
			name: "return StreamItem",
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- 100
					in <- 200
					in <- 300
					close(in)
				}()
				return in
			},
			execFunc: func(ctx context.Context, data int) api.StreamItem[int] {
				return api.StreamItem[int]{Item: data * 2}
			},
			tester: func(t *testing.T, out <-chan interface{}) {
				total := 0
				for data := range out {
					val := data.(api.StreamItem[int])
					total = total + val.Item
				}
				if total != 1200 {
					t.Fatal("unexpected result from operator func:", total)
				}
			},
		}

		runOpExecTest[int, api.StreamItem[int]](test, t)
	})

	t.Run("return nil value", func(t *testing.T) {
		test := opExecTest[string, any]{
			name: "op return nil",
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- "H"
					in <- "E"
					in <- "L"
					in <- "L"
					in <- "O"
					close(in)
				}()
				return in
			},
			execFunc: funcs.ExecFunction[string, any](func(ctx context.Context, data string) any {
				if data == "L" {
					return nil
				}
				return data
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				var result strings.Builder
				for data := range out {
					val := data.(string)
					result.WriteString(val)
				}
				if result.String() != "HEO" {
					t.Fatal("unexpected result from func:", result.String())
				}
			},
		}
		runOpExecTest[string, any](test, t)
	})

	t.Run("with StreamResult skipping item with no error", func(t *testing.T) {
		test := opExecTest[string, api.StreamResult]{
			name: "op return StreamResult",
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- "H"
					in <- "E"
					in <- "L"
					in <- "L"
					in <- "O"
					close(in)
				}()
				return in
			},
			execFunc: funcs.ExecFunction[string, api.StreamResult](func(ctx context.Context, data string) api.StreamResult {
				if data == "L" {
					return api.StreamResult{
						Value:  data,
						Action: api.ActionSkipItem,
					}
				}
				return api.StreamResult{Value: data}
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				var result strings.Builder
				for data := range out {
					val := data.(string)
					result.WriteString(val)
				}
				if result.String() != "HEO" {
					t.Fatal("unexpected result from func:", result.String())
				}
			},
		}
		runOpExecTest[string, api.StreamResult](test, t)
	})

	t.Run("with StreamResult forward item with no error", func(t *testing.T) {
		test := opExecTest[string, api.StreamResult]{
			name: "op return StreamResult",
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- "H"
					in <- "E"
					in <- "L"
					in <- "L"
					in <- "O"
					close(in)
				}()
				return in
			},
			execFunc: funcs.ExecFunction[string, api.StreamResult](func(ctx context.Context, data string) api.StreamResult {
				if data == "L" {
					return api.StreamResult{
						Value: data,
					}
				}
				return api.StreamResult{Value: data}
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				var result strings.Builder
				for data := range out {
					val := data.(string)
					result.WriteString(val)
				}
				if result.String() != "HELLO" {
					t.Fatal("unexpected result from func:", result.String())
				}
			},
		}
		runOpExecTest[string, api.StreamResult](test, t)
	})

	t.Run("with StreamResult skipping item with error", func(t *testing.T) {
		test := opExecTest[string, api.StreamResult]{
			name: "op return StreamResult",
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- "H"
					in <- "E"
					in <- "L"
					in <- "L"
					in <- "O"
					close(in)
				}()
				return in
			},
			execFunc: funcs.ExecFunction[string, api.StreamResult](func(ctx context.Context, data string) api.StreamResult {
				if data == "L" {
					return api.StreamResult{
						Value:  data,
						Err:    errors.New("unauthorized letter"),
						Action: api.ActionSkipItem,
					}
				}
				return api.StreamResult{Value: data}
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				var result strings.Builder
				for data := range out {
					val := data.(string)
					result.WriteString(val)
				}
				if result.String() != "HEO" {
					t.Fatal("unexpected result from func:", result.String())
				}
			},
		}
		runOpExecTest[string, api.StreamResult](test, t)
	})
}

func BenchmarkExecOperator(b *testing.B) {
	N := b.N

	chanSize := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	counter := 0
	var m sync.RWMutex
	op := funcs.ExecFunction[any, any](func(ctx context.Context, data any) any {
		m.Lock()
		counter++
		m.Unlock()
		return data
	})
	o := New(op)

	in := make(chan interface{}, chanSize)
	o.SetInput(in)
	go func() {
		for i := 0; i < N; i++ {
			in <- testutil.GenWord()
		}
		close(in)
	}()

	// process output
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range o.GetOutput() {
		}
	}()

	if err := o.Exec(context.TODO()); err != nil {
		b.Fatal("Error during execution:", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 60):
		b.Fatal("Took too long")
	}
	m.RLock()
	b.Logf("Input %d, counted %d", N, counter)
	if counter != N {
		b.Fatalf("Expected %d items processed,  got %d", N, counter)
	}
	m.RUnlock()
}
