package unary

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/testutil"
)

func TestUnaryOp_New(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		in     <-chan interface{}
		op     api.UnOperation
		concur int
	}{
		{
			name: "new op1",
			ctx:  context.Background(),
			op: api.UnFunc(func(context.Context, interface{}) interface{} {
				return nil
			}),
			concur: 2,
		},
		{
			name:   "new op2",
			ctx:    context.Background(),
			concur: 2,
			in:     make(chan interface{}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := New()

			if o.GetOutput() == nil {
				t.Fatal("default output should not be nil")
			}

			if o.op != nil {
				t.Fatal("Processing element should be nil")
			}

			if o.concurrency != 1 {
				t.Fatal("Concurrency should be initialized to 1.")
			}

			o.SetOperation(test.op)
			if (test.op != nil) && (o.op == nil) {
				t.Fatal("operation not set properly")
			}

			o.SetConcurrency(test.concur)
			if test.concur != o.concurrency {
				t.Fatal("unexpected concurrency value set:", o.concurrency)
			}

			o.SetInput(test.in)
			if test.in != nil && o.input == nil {
				t.Fatal("unexpected input set:", o.input)
			}
		})
	}
}

func TestUnaryOp_Exec(t *testing.T) {
	tests := []struct {
		name   string
		data   func() <-chan interface{}
		op     api.UnOperation
		tester func(*testing.T, <-chan interface{})
	}{
		{
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
			op: api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
				values := data.([]string)
				return len(values)
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				for data := range out {
					val := data.(int)
					if val != 3 && val != 2 && val != 1 {
						t.Fatalf("Expecting values 3, 2, or 1, but got %d", val)
					}
				}
			},
		},
		{
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
			op: api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
				value := data.(int)
				return api.StreamItem{Item: value * 2}
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				total := 0
				for data := range out {
					val := data.(api.StreamItem)
					total = total + val.Item.(int)
				}
				if total != 1200 {
					t.Fatal("unexpected result from operator func:", total)
				}
			},
		},
		{
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
			op: api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
				if data.(string) == "L" {
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
		},
		{
			name: "op return StreamError",
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
			op: api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
				if data.(string) == "L" {
					return api.Error("unauthorized letter")
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
		},
		{
			name: "op return StreamError with StreamItem",
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
			op: api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
				if data.(string) == "L" {
					return api.ErrorWithItem("unauthorized letter", &api.StreamItem{Item: data})
				}
				return api.StreamItem{Item: data}
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				var result strings.Builder
				for data := range out {
					item := data.(api.StreamItem)
					result.WriteString(item.Item.(string))
				}
				if result.String() != "HELLO" {
					t.Fatal("unexpected result from func:", result.String())
				}
			},
		},
		{
			name: "op return Cancelling stream error",
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- "H"
					in <- "I"
					close(in)
				}()
				return in
			},
			op: api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
				if data.(string) == "I" {
					return api.CancelStreamError(api.Error("panicking on unauthorized letter"))
				}
				return data
			}),
			tester: func(t *testing.T, out <-chan interface{}) {
				var result strings.Builder
				for data := range out {
					val := data.(string)
					result.WriteString(val)
				}
				if result.String() != "H" {
					t.Fatal("CancelStreamError not cancelling stream:", result.String())
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := New()
			o.SetInput(test.data())
			o.SetOperation(test.op)

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
}

func BenchmarkUnaryOp_Exec(b *testing.B) {
	o := New()
	N := b.N

	chanSize := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	in := make(chan interface{}, chanSize)
	o.SetInput(in)
	go func() {
		for i := 0; i < N; i++ {
			in <- testutil.GenWord()
		}
		close(in)
	}()

	counter := 0
	var m sync.RWMutex

	op := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		m.Lock()
		counter++
		m.Unlock()
		return data
	})
	o.SetOperation(op)

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
