package sinks

import (
	"context"
	"fmt"
	"strings" // Ensure strings is imported
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api" // For api.StreamLog
)

// Helper for logging in tests
func testSinkLogFunc(t *testing.T) api.StreamLogFunc {
	return func(_ context.Context, logEntry api.StreamLog) {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("SINK_TEST_LOG: %s", logEntry.Message))
		for _, attr := range logEntry.Attrs {
			builder.WriteString(fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any()))
		}
		t.Log(builder.String())
	}
}

func TestChanSink_Basic(t *testing.T) {
	data := []string{"A", "B", "C", "D", "E"}
	sourceChan := make(chan any, len(data))
	destChan := make(chan string, len(data)) // This is the channel the sink sends to

	for _, item := range data {
		sourceChan <- item
	}
	close(sourceChan) // Close source to signal end of stream to the sink's input

	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)

	errChan := sink.Open(context.Background())

	var receivedData []string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// destChan is NOT closed by the sink.
		// We read expected number of items.
		for i := 0; i < len(data); i++ {
			item, ok := <-destChan
			if !ok {
				t.Errorf("destChan closed prematurely after %d items", i)
				return
			}
			receivedData = append(receivedData, item)
		}
	}()

	err := <-errChan // Wait for sink to finish (errChan closed)
	if err != nil {
		t.Fatalf("Sink returned an error: %v", err)
	}

	wg.Wait() // Wait for the reading goroutine to finish

	if len(receivedData) != len(data) {
		t.Errorf("Expected %d items, got %d. Received: %v", len(data), len(receivedData), receivedData)
	}
	for i, expected := range data {
		if receivedData[i] != expected {
			t.Errorf("Expected item %s at index %d, got %s", expected, i, receivedData[i])
		}
	}
	close(destChan) // Test owns destChan, so it closes it.
}

func TestChanSink_DifferentTypes(t *testing.T) {
	type customType struct {
		ID   int
		Name string
	}
	data := []customType{{1, "Alice"}, {2, "Bob"}}
	sourceChan := make(chan any, len(data))
	destChan := make(chan customType, len(data))

	for _, item := range data {
		sourceChan <- item
	}
	close(sourceChan)

	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)
	errChan := sink.Open(context.Background())

	var receivedData []customType
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(data); i++ {
			item, ok := <-destChan
			if !ok {
				t.Errorf("destChan closed prematurely after %d items", i)
				return
			}
			receivedData = append(receivedData, item)
		}
	}()

	if err := <-errChan; err != nil {
		t.Fatalf("Sink error: %v", err)
	}
	wg.Wait()

	if len(receivedData) != len(data) {
		t.Fatalf("Expected %d items, got %d", len(data), len(receivedData))
	}
	if receivedData[0] != data[0] || receivedData[1] != data[1] {
		t.Errorf("Data mismatch: expected %v, got %v", data, receivedData)
	}
	close(destChan) // Test owns destChan
}

func TestChanSink_ContextCancellation(t *testing.T) {
	sourceChan := make(chan any)    // Unbuffered
	destChan := make(chan int)       // Unbuffered, sink sends here

	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := sink.Open(ctx)

	go func() {
		select {
		case sourceChan <- 123:
		case <-time.After(1 * time.Second):
			// Use t.Error or t.Fatal for reporting errors in goroutines if using `t` directly.
			// For simplicity, this might just log or do nothing, relying on main thread's timeout.
		}
	}()

	select {
	case item, ok := <-destChan:
		if !ok {
			t.Fatal("destChan closed prematurely before item received")
		}
		if item != 123 {
			t.Fatalf("Expected 123, got %v", item)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for item from destChan")
	}
	
	cancel() 

	err := <-errChan 
	if err != nil {
		t.Logf("Sink returned error: %v (this is acceptable for context cancellation)", err)
	} else {
		t.Log("Sink errChan closed without error on context cancellation.")
	}

	// destChan should remain open as sink does not close it.
	select {
	case _, ok := <-destChan:
		if !ok {
			t.Error("destChan was unexpectedly closed")
		} else {
			t.Log("destChan still open and has items (if any were sent post-cancellation check), or is empty.")
		}
	default:
		t.Log("destChan is open and would block (empty), as expected.")
	}

	close(sourceChan) 
	close(destChan)   
}

func TestChanSink_NilInputChannel(t *testing.T) {
	destChan := make(chan string)
	
	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	// SetInput not called

	errChan := sink.Open(context.Background()) 
	err := <-errChan

	if err == nil {
		t.Fatal("Expected an error for nil input channel, got nil")
	}
	if err.Error() != api.ErrInputChannelUndefined.Error() {
		t.Errorf("Expected error '%s', got '%s'", api.ErrInputChannelUndefined.Error(), err.Error())
	}
	close(destChan) 
}

func TestChanSink_NilOutputChannel(t *testing.T) {
	sourceChan := make(chan any)
	close(sourceChan) 

	sink := Channel[string](nil) 
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)

	errChan := sink.Open(context.Background()) 
	err := <-errChan

	if err == nil {
		t.Fatal("Expected an error for nil output channel, got nil")
	}
	if err.Error() != api.ErrSinkDestinationUndefined.Error() {
		t.Errorf("Expected error '%s', got '%s'", api.ErrSinkDestinationUndefined.Error(), err.Error())
	}
}

func TestChanSink_TypeMismatch(t *testing.T) {
	sourceChan := make(chan any, 2)
	destChan := make(chan int, 1) 

	sourceChan <- 123       
	sourceChan <- "not_an_int" 
	close(sourceChan)

	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)

	errChan := sink.Open(context.Background())

	var receivedItems []int
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		item, ok := <-destChan
		if ok {
			receivedItems = append(receivedItems, item)
		} else {
			// This path might be taken if destChan is closed by other means,
			// or if no item is sent. Test expects one item.
		}
	}()
	
	err := <-errChan 
	if err != nil {
		t.Fatalf("Sink returned an error: %v", err) 
	}
	wg.Wait()

	if len(receivedItems) != 1 {
		t.Fatalf("Expected 1 item of correct type, got %d. Received: %v", len(receivedItems), receivedItems)
	}
	if receivedItems[0] != 123 {
		t.Errorf("Expected item 123, got %d", receivedItems[0])
	}
	close(destChan) 
}

func TestChanSink_OutputChannelClosedPrematurely(t *testing.T) {
	sourceChan := make(chan any, 10) 
	for i := 0; i < 5; i++ {        
		sourceChan <- fmt.Sprintf("item-%d", i)
	}

	destChan := make(chan string, 10) 

	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)

	errChan := sink.Open(context.Background())

	var firstItem string
	var ok bool
	select {
	case firstItem, ok = <-destChan:
		if !ok {
			t.Fatal("destChan closed before any item was sent by sink")
		}
		t.Logf("Test: Received first item: %s", firstItem)
	case <-time.After(1 * time.Second):
		t.Fatal("Test: Timed out waiting for the first item from destChan")
	}

	go func() {
		time.Sleep(50 * time.Millisecond) 
		t.Log("Test Goroutine: Prematurely closing destChan")
		close(destChan)
	}()

	finalErr := <-errChan // This blocks until errChan is closed.

	if finalErr == nil {
		t.Fatalf("Expected an error from sink due to panic (send on closed channel), but got nil (errChan closed without error)")
	}

	expectedPanicSubstring := "panic in ChanSink worker" 
	if !strings.Contains(finalErr.Error(), expectedPanicSubstring) {
		t.Errorf("Error received does not indicate a panic. Got: '%v', Expected substring: '%s'", finalErr, expectedPanicSubstring)
	} else {
		t.Logf("Test: Correctly received panic-related error from sink: %v", finalErr)
	}

	// Cleanup sourceChan: send remaining items or close it.
	// This goroutine is for cleanup and observation.
	go func() {
		defer func() {
			// Recover from potential panic if sourceChan is already closed,
			// though this test structure doesn't close it elsewhere until here.
			recover() 
			// Ensure sourceChan is closed if not already by previous logic
			// This is complex because sourceChan could be closed by another part of a failing test.
			// A select-based close is safer.
			select {
			case _, stillOpen := <-sourceChan:
				if stillOpen { // If it had an item, this would be true.
					// If it was empty and open, this would block.
					// A non-blocking read to check if closed is better:
					// chk := make(chan struct{})
					// go func() { _, _ = <-sourceChan; close(chk)}()
					// select { case <-chk: close(sourceChan) if not already... etc.}
					// For simplicity now, just try to close it.
					// If this panics because it's already closed, the recover handles it.
				}
			default: // sourceChan is empty or already closed
			}
			// At this point, it's hard to know state of sourceChan if test failed early.
			// Best effort: try to close. If it panics, outer test already failed.
			// Or, just remove this potentially problematic cleanup.
			// For now, let's assume we might need to close it.
			// close(sourceChan) // This might panic if test failed and sourceChan was already closed.
		}()
		// Try to send remaining items from original loop, though sink is likely dead.
		// This mainly ensures this goroutine doesn't block indefinitely on sourceChan if it's unbuffered.
		for i := 5; i < 8; i++ { 
			select {
			case sourceChan <- fmt.Sprintf("late-item-%d", i):
			case <-time.After(50 * time.Millisecond):
				return 
			}
		}
		// After attempting to send, close sourceChan if it wasn't closed by panic recovery.
		// This is tricky because its state is uncertain if the main test assertions fail.
		// A robust way: try a non-blocking send of a special "close signal" or just close.
		// If this sub-goroutine is just for "observational" purposes post-assertion,
		// its cleanup is secondary to the main test logic.
		close(sourceChan) // Close it once done sending.
	}()

	var receivedItemsAfterClose []string
	for remainingItem := range destChan { 
		receivedItemsAfterClose = append(receivedItemsAfterClose, remainingItem)
	}
	t.Logf("Test: Received %d items from destChan after it was closed by test's goroutine: %v", len(receivedItemsAfterClose), receivedItemsAfterClose)
}
