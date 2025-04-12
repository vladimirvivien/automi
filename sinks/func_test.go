package sinks

import (
	"context"
	"testing"
	"time"
)

func TestFuncSink(t *testing.T) {
	count := 0
	f := Func[string](func(val string) error {
		count++
		return nil
	})
	in := make(chan any)
	go func() {
		in <- "String 1"
		in <- "String 2"
		in <- "String 3"
		close(in)
	}()
	f.SetInput(in)

	select {
	case err := <-f.Open(context.TODO()):
		if err != nil {
			t.Fatal(err)
		}
		if count != 3 {
			t.Fatal("expecting count 3, got ", count)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}

func TestFuncSinkErr(t *testing.T) {
	f := Func[string](nil)

	select {
	case err := <-f.Open(context.TODO()):
		if err == nil {
			t.Fatal("Expecting error")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}
