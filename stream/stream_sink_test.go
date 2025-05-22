package stream

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/sources"
	"github.com/vladimirvivien/automi/testutil"
)

func TestStreamToCSVSource(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	snk := new(bytes.Buffer)
	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.CSV(snk))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(snk.String()), "\n")
		if len(lines) != 5 {
			t.Error("unexpected sink data: want 5 lines, got", len(lines))
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

// TestStreamToChannelSink tests a simple stream pipeline:
// sources.Slice -> sinks.Channel
func TestStreamToChannelSink(t *testing.T) {
	data := []string{"apple", "banana", "cherry"}
	source := sources.Slice(data)

	destChan := make(chan string, len(data))

	strm := From(source)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Channel(destChan))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatalf("Stream returned an error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Stream processing timed out")
	}

	var receivedData []string
	for item := range destChan { // destChan is closed by the sink
		receivedData = append(receivedData, item)
	}

	if len(receivedData) != len(data) {
		t.Errorf("Expected %d items, got %d. Received: %v", len(data), len(receivedData), receivedData)
	}
	for i, expected := range data {
		if receivedData[i] != expected {
			t.Errorf("Expected item '%s' at index %d, got '%s'", expected, i, receivedData[i])
		}
	}
}

// TestStreamBridgeWithChanSinkAndChanSource tests the user's bridge scenario:
// strm1: sources.Slice -> sinks.Channel(bridgeChan)
// strm2: sources.Chan(bridgeChan) -> sinks.Func
func TestStreamBridgeWithChanSinkAndChanSource(t *testing.T) {
	numItems := 100
	sourceData := make([]int, numItems)
	for i := 0; i < numItems; i++ {
		sourceData[i] = i
	}

	bridgeChan := make(chan int, numItems)

	strm1 := From(sources.Slice(sourceData))
	strm1.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm1.Into(sinks.Channel(bridgeChan))

	var receivedData []int
	var mu sync.Mutex

	collectorSink := sinks.Func(func(item int) error {
		mu.Lock()
		receivedData = append(receivedData, item)
		mu.Unlock()
		return nil
	})

	strm2 := From(sources.Chan(bridgeChan))
	strm2.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm2.Into(collectorSink)
	
	var wg sync.WaitGroup
	wg.Add(2)
	var errStrm1, errStrm2 error

	go func() {
		defer wg.Done()
		select {
		case err := <-strm1.Open(context.Background()):
			if err != nil {
				errStrm1 = fmt.Errorf("stream1 error: %w", err)
			}
		case <-time.After(2 * time.Second):
			errStrm1 = fmt.Errorf("stream1 processing timed out")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case err := <-strm2.Open(context.Background()):
			if err != nil {
				errStrm2 = fmt.Errorf("stream2 error: %w", err)
			}
		case <-time.After(3 * time.Second): // Slightly longer for strm2
			errStrm2 = fmt.Errorf("stream2 processing timed out")
		}
	}()

	wg.Wait()

	if errStrm1 != nil {
		t.Errorf("Error in Stream 1: %v", errStrm1)
	}
	if errStrm2 != nil {
		t.Errorf("Error in Stream 2: %v", errStrm2)
	}
	
	mu.Lock()
	defer mu.Unlock()

	if len(receivedData) != numItems {
		t.Errorf("Expected %d items, got %d. Received: %v", numItems, len(receivedData), receivedData)
	} else {
		for i := 0; i < numItems; i++ {
			expectedValue := sourceData[i]
			found := false
			for _, val := range receivedData {
				if val == expectedValue {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected value %d not found. Received: %v", expectedValue, receivedData)
			}
		}
	}
}

func TestStreamToFunc(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	var count atomic.Int32
	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Func(func(item []string) error {
		count.Add(1)
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		if count.Load() != 10 {
			t.Error("unexpected sink data: want 10, got", count.Load())
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamToDiscard(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Discard())

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamToSlice(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	snk := sinks.Slice[[]string]()
	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(snk)

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		data := snk.Get()
		if len(data) != 10 {
			t.Errorf("unexpected sink data: want 10, got %d", len(data))
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
