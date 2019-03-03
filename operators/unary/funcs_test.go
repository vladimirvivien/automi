package unary

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestUnaryFunc_Process(t *testing.T) {
	tests := []struct {
		name           string
		procFunc       interface{}
		input          interface{}
		expected       interface{}
		ctx            context.Context
		funcShouldFail bool
		opShouldFail   bool
	}{
		{
			name: "unary funcForm1, OK",
			procFunc: func(item int) int {
				return item * 2
			},
			input:    6,
			expected: 12,
		},
		{
			name: "unary funcForm1, procFunc fail",
			procFunc: func(item string) string {
				return strings.ToUpper(item)
			},
			input:        "hello",
			expected:     "BELLO",
			opShouldFail: true,
		},
		{
			name: "unary funcForm2, OK",
			ctx:  context.Background(),
			procFunc: func(ctx context.Context, item string) string {
				return strings.ToUpper(item)
			},
			input:    "hello",
			expected: "HELLO",
		},
		{
			name: "unary with two returns",
			procFunc: func(item string) (string, int) {
				return item, 0
			},
			funcShouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op, err := ProcessFunc(test.procFunc)
			if test.funcShouldFail && err == nil {
				t.Errorf("expecting failure, but error is nil")
			}
			switch {
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
		})
	}

}

func TestUnaryFunc_Filter(t *testing.T) {
	op, err := FilterFunc(func(item string) bool {
		return item[0] == []byte("T")[0]
	})

	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()
	var strings []string
	for _, v := range []string{"Mon", "Wed", "Tr", "Sun", "Tu", "Sat"} {
		result := op.Apply(ctx, v)
		str, ok := result.(string)
		if ok {
			strings = append(strings, str)
		}
	}
	if len(strings) != 2 {
		t.Fatal("unexpected result from FilterFunc:", strings)
	}
}

func TestUnaryFunc_Map(t *testing.T) {
	op, err := MapFunc(func(item rune) rune {
		return item - 32
	})

	if err != nil {
		t.Fatal(err)
	}

	str := new(bytes.Buffer)
	ctx := context.TODO()
	for _, v := range []rune{'h', 'e', 'l', 'l', 'o'} {
		result := op.Apply(ctx, v)
		chr := result.(rune)
		str.WriteRune(chr)
	}

	if str.String() != "HELLO" {
		t.Fatal("unexpected result from MapFunc:", str)
	}
}

func TestUnaryFunc_FlatMap(t *testing.T) {
	op, err := FlatMapFunc(func(msg string) []rune {
		return []rune(msg)
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()
	result := op.Apply(ctx, "hello")
	slice := result.([]rune)
	if len(slice) != 5 {
		t.Fatal("unexpected result from FlatMapFunc:", len(slice))
	}
}
