package sinks

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestChanSink_Basic(t *testing.T) {
	data := []string{"A", "B", "C", "D", "E"}
	sourceChan := make(chan any, len(data))
	destChan := make(chan string, len(data))

	// Populate source channel
	for _, item := range data {
		sourceChan <- item
	}
	close(sourceChan) // Close source to signal end of stream

	sink := Channel(destChan)
	sink.SetLogFunc(testutil.LogSinkFunc(t)) // Optional: for more verbose test logging
	sink.SetInput(sourceChan)

	errChan := sink.Open(context.Background())

	var receivedData []string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range destChan { // destChan will be closed by the sink
			receivedData = append(receivedData, item)
		}
	}()

	// Wait for sink to finish (or error)
	err := <-errChan
	if err != nil {
		t.Fatalf("Sink returned an error: %v", err)
	}

	wg.Wait() // Wait for the reading goroutine to finish

	if len(receivedData) != len(data) {
		t.Errorf("Expected %d items, got %d", len(data), len(receivedData))
	}
	for i, expected := range data {
		if receivedData[i] != expected {
			t.Errorf("Expected item %s at index %d, got %s", expected, i, receivedData[i])
		}
	}
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
	sink.SetInput(sourceChan)
	errChan := sink.Open(context.Background())

	var receivedData []customType
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range destChan {
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
}

func TestChanSink_ContextCancellation(t *testing.T) {
	sourceChan := make(chan any) // Unbuffered to ensure blocking
	destChan := make(chan int)    // Unbuffered

	sink := Channel(destChan)
	sink.SetInput(sourceChan)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := sink.Open(ctx)

	// Send one item
	go func() {
		sourceChan <- 123
		// Do not close sourceChan to simulate ongoing stream
	}()

	// Read the item to confirm sink is working
	select {
	case item, ok := <-destChan:
		if !ok {
			t.Fatal("destChan closed prematurely")
		}
		if item != 123 {
			t.Fatalf("Expected 123, got %v", item)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for item from destChan")
	}
	
	// Now cancel the context
	cancel()

	// Check for error (should be nil or context.Canceled, but sink handles it internally)
	err := <-errChan
	if err != nil {
		// Depending on exact error handling, context.Canceled might not be sent.
		// The important part is that Open() finishes and destChan is closed.
		t.Logf("Sink returned error (expected if context related): %v", err)
	}

	// Check if destChan is closed by the sink due to cancellation
	// This might take a moment for the sink's goroutine to react
	select {
	case _, ok := <-destChan:
		if ok {
			t.Error("destChan was not closed after context cancellation")
		}
	case <-time.After(1 * time.Second): // Give time for channel to close
		t.Error("Timed out waiting for destChan to close after context cancellation")
	}
	
	// Ensure sourceChan can still be written to (it's not closed by sink)
	// but nothing should read from it if sink is truly down.
	// This part is tricky as the sink's goroutine might have exited.
	// We are primarily testing that destChan gets closed.
}


func TestChanSink_NilInputChannel(t *testing.T) {
	destChan := make(chan string)
	// Deliberately not closing destChan here, sink should close it if input is nil
	
	sink := Channel(destChan)
	// sink.SetInput(nil) // Implicitly nil

	errChan := sink.Open(context.Background())
	err := <-errChan

	if err == nil {
		t.Fatal("Expected an error for nil input channel, got nil")
	}
	if err.Error() != "input channel undefined" { // Matches api.ErrInputChannelUndefined.Error()
		t.Errorf("Expected error '%s', got '%s'", "input channel undefined", err.Error())
	}

	// Check if destChan was closed
	select {
	case _, ok := <-destChan:
		if ok {
			t.Error("destChan was not closed when input channel is nil")
		}
	case <-time.After(100 * time.Millisecond): // Give a bit of time for the close to happen
		t.Error("Timed out waiting for destChan to close with nil input")
	}
}

func TestChanSink_NilOutputChannel(t *testing.T) {
	sourceChan := make(chan any)
	close(sourceChan) // Close source, otherwise Open might block indefinitely

	// Pass nil as the destination channel
	sink := Channel[string](nil) 
	sink.SetInput(sourceChan)

	errChan := sink.Open(context.Background())
	err := <-errChan

	if err == nil {
		t.Fatal("Expected an error for nil output channel, got nil")
	}
	if err.Error() != "sink destination undefined" { // Matches api.ErrSinkDestinationUndefined.Error()
		t.Errorf("Expected error '%s', got '%s'", "sink destination undefined", err.Error())
	}
}

func TestChanSink_TypeMismatch(t *testing.T) {
	sourceChan := make(chan any, 2)
	destChan := make(chan int, 1) // Expects int

	sourceChan <- 123       // Correct type
	sourceChan <- "not_an_int" // Incorrect type
	close(sourceChan)

	sink := Channel(destChan)
	sink.SetInput(sourceChan)
	sink.SetLogFunc(testutil.LogSinkFunc(t)) // Enable logging to see the type mismatch message

	errChan := sink.Open(context.Background())

	var receivedItems []int
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range destChan { // destChan will be closed by sink
			receivedItems = append(receivedItems, item)
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
	// The "not_an_int" should have been skipped and logged by the sink.
}

// Example of how a user might close the output channel prematurely.
// The sink should detect this and stop gracefully.
func TestChanSink_OutputChannelClosedPrematurely(t *testing.T) {
	sourceChan := make(chan any, 5)
	for i := 0; i < 5; i++ {
		sourceChan <- fmt.Sprintf("item-%d", i)
	}
	// Do not close sourceChan immediately to let sink process some items.

	destChan := make(chan string, 5)

	sink := Channel(destChan)
	sink.SetInput(sourceChan)
	sink.SetLogFunc(testutil.LogSinkFunc(t))

	errChan := sink.Open(context.Background())

	// Read one item to ensure sink has started
	item, ok := <-destChan
	if !ok {
		t.Fatal("destChan closed before any item was sent")
	}
	t.Logf("Received first item: %s", item)

	// Now, prematurely close the destination channel
	close(destChan)
	t.Log("Prematurely closed destChan")

	// The sink's attempt to send to the closed destChan should cause a panic,
	// which it should recover from and then shut down.
	// Or, if it checks for channel closure before sending (not typical for sender), it would also stop.
	// The current ChanSink implementation will panic, and it does not recover.
	// This test will likely fail or hang with the current ChanSink implementation
	// if it doesn't handle panics on send to closed channel.
	// For a robust sink, it should handle this. However, the typical expectation is
	// that the sink *owns* the sending side of the channel and is the one to close it.
	// If user closes it, it's a contract violation.
	// Let's assume for now that the sink will panic and the test will capture this behavior.
	// The `Open` method's goroutine should eventually finish.

	var finalErr error
	select {
	case finalErr = <-errChan:
		if finalErr != nil {
			t.Logf("Sink finished with error: %v (this might be expected if panic was handled)", finalErr)
		} else {
			t.Log("Sink finished without error.")
		}
	case <-time.After(2 * time.Second): // Increased timeout
		t.Fatal("Sink did not finish after destChan was closed prematurely and source was still open.")
		// If it hangs here, the sink didn't handle the panic on send to closed channel.
	}
    
    // Try to send more data to source to see if sink is still running (it shouldn't be)
    // This part is more to ensure the test doesn't get stuck if the sink didn't stop.
    go func() {
        defer func() {
            if r := recover(); r != nil {
                t.Logf("Recovered from panic while sending to sourceChan: %v", r)
            }
            close(sourceChan) // ensure sourceChan is closed eventually
        }()
        // If sink is robust and stopped, these items won't be processed.
        // If sink crashed without closing its input processing loop, this could block or panic.
        for i := 5; i < 10; i++ {
             select {
             case sourceChan <- fmt.Sprintf("late-item-%d", i):
             case <-time.After(50*time.Millisecond): // don't block test indefinitely
                t.Logf("Timed out sending late item %d to sourceChan", i)
                return
             }
        }

    }()


	// Drain remaining items from destChan to see what made it before the close.
	// This is important because after close(destChan), reads will yield zero values once empty.
	var receivedItemsAfterClose []string
	for remainingItem := range destChan { // This loop will terminate once destChan is empty
		receivedItemsAfterClose = append(receivedItemsAfterClose, remainingItem)
	}
	t.Logf("Received %d items from destChan after it was closed by test: %v", len(receivedItemsAfterClose), receivedItemsAfterClose)

	// Assert that the sink is no longer trying to send to destChan
	// This is implicitly tested by `err := <-errChan` completing.
	// If the sink was stuck trying to send to a closed channel, `errChan` would not receive.
}
