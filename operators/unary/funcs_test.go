package unary

import (
	"bytes"
	"context"
	"testing"
)

func TestUnaryFunc_Process(t *testing.T) {
	op, err := ProcessFunc(func(item int) int {
		return item * 2
	})

	if err != nil {
		t.Fatal(err)
	}

	sum := 0
	ctx := context.TODO()
	for _, v := range []int{2, 4, 6, 8} {
		result := op.Apply(ctx, v)
		sum += result.(int)
	}

	if sum != 40 {
		t.Fatal("unexpected result from ProcessFunc:", sum)
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
