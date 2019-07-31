package unary

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/vladimirvivien/automi/api"
)

type unaryFuncTestCase struct {
	name           string
	opBuilder      func(interface{}) (api.UnFunc, error)
	procFunc       interface{}
	input          interface{}
	expected       interface{}
	ctx            context.Context
	funcShouldFail bool
	opShouldFail   bool
}

func testUnaryFunc(t *testing.T, test unaryFuncTestCase) {
	op, err := test.opBuilder(test.procFunc)

	switch {
	case test.funcShouldFail && err == nil:
		t.Errorf("expecting failure, but error is nil")
	case test.funcShouldFail && err != nil:
		t.Log(err)
	case !test.funcShouldFail:
		if err != nil {
			t.Fatal(err)
		}
		result := op.Apply(test.ctx, test.input)
		if !test.opShouldFail && !reflect.DeepEqual(result, test.expected) {
			t.Errorf("expecting %v got %v", test.expected, result)
		}
	case test.opShouldFail:
		if err == nil {
			t.Fatal("expecting failure, but error is nil")
		}
	}
}

func TestUnaryFunc_Process(t *testing.T) {
	tests := []unaryFuncTestCase{
		{
			name:      "unary funcForm1, OK",
			opBuilder: ProcessFunc,
			procFunc: func(item int) int {
				return item * 2
			},
			input:    6,
			expected: 12,
		},
		{
			name:      "unary funcForm1, procFunc fail",
			opBuilder: ProcessFunc,
			procFunc: func(item string) string {
				return strings.ToUpper(item)
			},
			input:        "hello",
			expected:     "BELLO",
			opShouldFail: true,
		},
		{
			name:      "unary funcForm2, OK",
			opBuilder: ProcessFunc,
			ctx:       context.Background(),
			procFunc: func(ctx context.Context, item string) string {
				return strings.ToUpper(item)
			},
			input:    "hello",
			expected: "HELLO",
		},
		{
			name:      "unary with two returns",
			opBuilder: ProcessFunc,
			procFunc: func(item string) (string, int) {
				return item, 0
			},
			funcShouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testUnaryFunc(t, test)
		})
	}
}
func TestUnaryFunc_Filter(t *testing.T) {
	tests := []unaryFuncTestCase{
		{
			name:      "filter out data with form f(in)bool",
			opBuilder: FilterFunc,
			input:     []string{"Mon", "Tue", "Wed"},
			procFunc: func(days []string) bool {
				return (len(days) < 3)
			},
			expected: nil,
		},
		{
			name:      "allow with form f(in)bool",
			opBuilder: FilterFunc,
			input:     []string{"Mon", "Tue", "Wed"},
			procFunc: func(days []string) bool {
				return (len(days) >= 3)
			},
			expected: []string{"Mon", "Tue", "Wed"},
		},
		{
			name:      "unexpected return value with form f(in)bool",
			opBuilder: FilterFunc,
			input:     "hello",
			procFunc: func(item string) bool {
				return (item == "HELLO")
			},
			expected:     "HELLO",
			opShouldFail: true,
		},
		{
			name:      "filter out data with form f(ctx,in)bool",
			opBuilder: FilterFunc,
			ctx:       context.Background(),
			input:     []string{"Mon", "Tue", "Wed"},
			procFunc: func(ctx context.Context, days []string) bool {
				return (len(days) < 3)
			},
			expected: nil,
		},
		{
			name:      "filter bad with func f(in)(bool,err)",
			opBuilder: FilterFunc,
			input:     []string{"Mon", "Tue", "Wed"},
			procFunc: func(days []string) (bool, error) {
				return (len(days) < 3), nil
			},
			funcShouldFail: true,
		},
		{
			name:      "filter bad with func f(ctx, in)(bool,err)",
			opBuilder: FilterFunc,
			input:     []string{"Mon", "Tue", "Wed"},
			procFunc: func(ctx context.Context, days []string) (bool, error) {
				return (len(days) < 3), nil
			},
			funcShouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testUnaryFunc(t, test)
		})
	}
}

func TestUnaryFunc_Map(t *testing.T) {
	tests := []unaryFuncTestCase{
		{
			name:      "Map data with function f(in)out",
			opBuilder: MapFunc,
			input:     []byte{'H', 'e', 'l', 'l', 'o'},
			procFunc: func(item []byte) string {
				return string(item)
			},
			expected: "Hello",
		},
		{
			name:      "Map data with op f(in)out with failure",
			opBuilder: MapFunc,
			input:     'A',
			procFunc: func(item rune) rune {
				return item - 32
			},
			expected:     'A' - 31,
			opShouldFail: true,
		},
		{
			name:      "Map data with function f(ctx, in)out",
			opBuilder: MapFunc,
			ctx:       context.Background(),
			input:     []byte{'H', 'e', 'l', 'l', 'o'},
			procFunc: func(ctx context.Context, item []byte) string {
				return string(item)
			},
			expected: "Hello",
		},
		{
			name:      "Map data with function f(ctx,in)out with op failure",
			opBuilder: MapFunc,
			input:     'H',
			procFunc: func(item rune) string {
				return string(item)
			},
			expected:     "H",
			opShouldFail: true,
		},
		{
			name:      "Map data with function f(ctx,in)(out0,out1) with op failure",
			opBuilder: MapFunc,
			input:     'H',
			procFunc: func(item rune) (string, string) {
				return string(item), string(item)
			},
			expected:       "HH",
			funcShouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testUnaryFunc(t, test)
		})
	}
}

func TestUnaryFunc_FlatMap(t *testing.T) {
	tests := []unaryFuncTestCase{
		{
			name:      "FlatMap data func function f(in)out",
			opBuilder: FlatMapFunc,
			input:     "Hello",
			procFunc: func(item string) []rune {
				return []rune(item)
			},
			expected: []rune("Hello"),
		},
		{
			name:      "FlatMap data with func f(in)out with wrong result param type",
			opBuilder: FlatMapFunc,
			input:     "Hello",
			procFunc: func(item string) string {
				return item
			},
			funcShouldFail: true,
		},
		{
			name:      "FlatMap data op fun f(ctx,in)out",
			ctx:       context.Background(),
			opBuilder: FlatMapFunc,
			input:     "Hello",
			procFunc: func(item string) []rune {
				return []rune(item)
			},
			expected: []rune("Hello"),
		},
		{
			name:      "FlatMap data op func f(ctx,in)out",
			ctx:       context.Background(),
			opBuilder: FlatMapFunc,
			input:     "Hello",
			procFunc: func(item string) []rune {
				return []rune(item)
			},
			expected:     []rune("Bello"),
			opShouldFail: true,
		},
		{
			name:      "FlatMap data with func f(ctx,in)(out0,out1) with op failure",
			opBuilder: MapFunc,
			input:     `Hello`,
			procFunc: func(item string) ([]byte, []byte) {
				return []byte(item), []byte(item)
			},
			funcShouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testUnaryFunc(t, test)
		})
	}
}
