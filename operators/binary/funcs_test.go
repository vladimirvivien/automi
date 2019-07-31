package binary

import (
	"context"
	"testing"
)

func TestBinaryFunc_Reduce(t *testing.T) {
	op, err := ReduceFunc(func(op0, op1 int) int {
		return op0 + op1
	})
	if err != nil {
		t.Fatal(err)
	}

	seed := 0
	ctx := context.TODO()

	for _, v := range []int{1, 2, 3, 4, 5} {
		result := op.Apply(ctx, seed, v)
		seed = result.(int)
	}

	if seed != 15 {
		t.Fatal("unexpected result from ReduceFunc: ", seed)
	}
}
