package stream

import (
	"context"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_Log_No_Log(t *testing.T) {
	src := emitters.Slice([]string{"hello", "world"})
	strm := New(src).Into(collectors.Null())

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}

func TestStream_Log(t *testing.T) {
	lock := sync.Mutex{}
	count := 0

	ctx := context.Background()
	src := emitters.Slice([]string{"hello", "world"})

	strm := New(src).WithContext(ctx)
	strm.Process(func(val string) error {
		if err := autoctx.Log(strm.GetContext(), "Processing"); err != nil {
			t.Fatal(err)
		}
		return nil
	})

	strm.WithLogFunc(func(val interface{}) {
		lock.Lock()
		count++
		lock.Unlock()
	})

	strm.Into(collectors.Null())

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}

	lock.Lock()
	if count < 2 {
		t.Log("count: ", count)
		t.Fatal("logger func not logging properly")
	}
	lock.Unlock()
}

func TestStream_Log_With_Logger(t *testing.T) {
	var logs strings.Builder
	logger := log.New(&logs, "", log.Flags())

	ctx := context.Background()
	src := emitters.Slice([]string{"hello", "world"})

	strm := New(src).WithContext(ctx)
	strm.Process(func(val string) error {
		if err := autoctx.Log(strm.GetContext(), "Processing"); err != nil {
			t.Fatal(err)
		}
		return nil
	})

	strm.WithLogFunc(func(val interface{}) {
		logger.Println(val)
	})

	strm.Into(collectors.Null())

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}

	lines := strings.Split(logs.String(), "\n")
	if len(lines) < 2 {
		t.Fatal("logger func not logging properly")
	}
}
