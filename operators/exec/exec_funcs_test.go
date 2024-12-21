package exec

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
)

type funcExecutorTest[IN, OUT any] struct {
	funcExec *ExecOperator[IN, OUT]
	data     func() <-chan interface{}
	tester   func(*testing.T, <-chan interface{})
}

func runFuncExecutorTest[IN, OUT any](t *testing.T, test funcExecutorTest[IN, OUT]) {
	o := test.funcExec
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
}

func TestExecuteFunc(t *testing.T) {
	t.Run("normal exec", func(t *testing.T) {
		test := funcExecutorTest[int, int]{
			funcExec: Execute(func(ctx context.Context, i int) int {
				return i * 2
			}),
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
			tester: func(t *testing.T, out <-chan interface{}) {
				total := 0
				for data := range out {
					val := data.(int)
					total = total + val
				}
				if total != 1200 {
					t.Fatal("unexpected result from operator func:", total)
				}
			},
		}
		runFuncExecutorTest[int, int](t, test)
	})

	t.Run("return nil val", func(t *testing.T) {
		test := funcExecutorTest[string, any]{
			funcExec: Execute(func(ctx context.Context, data string) any {
				if data == "L" {
					return nil
				}
				return data
			}),
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
		runFuncExecutorTest[string, any](t, test)
	})

	t.Run("return struct", func(t *testing.T) {
		test := funcExecutorTest[int, api.StreamItem[int]]{
			funcExec: Execute(func(ctx context.Context, data int) api.StreamItem[int] {
				return api.StreamItem[int]{Item: data * 2}
			}),
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
		runFuncExecutorTest[int, api.StreamItem[int]](t, test)
	})
}
func TestFilterFunc(t *testing.T) {
	t.Run("simple filter", func(t *testing.T) {
		test := funcExecutorTest[[]string, api.FilterItem[[]string]]{
			funcExec: Filter(func(ctx context.Context, data []string) bool {
				return len(data) > 3
			}),
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- []string{"H", "E", "L", "L", "O"}
					in <- []string{"H", "I", "!"}
					in <- []string{"W", "O", "R", "L", "D", "!"}
					close(in)
				}()
				return in
			},
			tester: func(t *testing.T, out <-chan interface{}) {
				var result []string
				for data := range out {
					result = append(result, data.([]string)...)
				}
				if len(result) != 11 {
					t.Errorf("Unexpected items filtered")
				}
			},
		}
		runFuncExecutorTest[[]string, api.FilterItem[[]string]](t, test)
	})
}

func TestMapFunc(t *testing.T) {
	t.Run("simple map", func(t *testing.T) {
		test := funcExecutorTest[[]byte, string]{
			funcExec: Map(func(ctx context.Context, data []byte) string {
				return string(data)
			}),
			data: func() <-chan interface{} {
				in := make(chan interface{})
				go func() {
					in <- []byte{'H', 'E', 'L', 'L', 'O', ' '}
					in <- []byte{'W', 'O', 'R', 'L', 'D', '!'}
					close(in)
				}()
				return in
			},
			tester: func(t *testing.T, out <-chan interface{}) {
				var result strings.Builder
				for data := range out {
					result.WriteString(data.(string))
				}
				if result.String() != "HELLO WORLD!" {
					t.Errorf("Unexpected mapped item: %s", result.String())
				}
			},
		}
		runFuncExecutorTest[[]byte, string](t, test)
	})
}
