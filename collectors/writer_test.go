package collectors

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestCollector_Writer_Bytes(t *testing.T) {
	sink := bytes.NewBufferString("")
	w := Writer(sink)
	in := make(chan interface{})
	go func() {
		in <- []byte("What a ")
		in <- []byte("wonderful ")
		in <- []byte("world")
		close(in)
	}()
	w.SetInput(in)
	expected := "What a wonderful world"
	select {
	case err := <-w.Open(context.TODO()):
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(sink.String()) != expected {
			t.Fatal("unexpected result ", sink.String())
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}

}

func TestCollector_Writer_String(t *testing.T) {
	sink := bytes.NewBufferString("")
	w := Writer(sink)
	in := make(chan interface{})
	go func() {
		in <- "What a "
		in <- "wonderful "
		in <- "world"
		close(in)
	}()
	w.SetInput(in)
	expected := "What a wonderful world"
	select {
	case err := <-w.Open(context.TODO()):
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(sink.String()) != expected {
			t.Fatal("unexpected result ", sink.String())
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}

}
