package sinks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
)

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
	destChan := make(chan string, len(data)) 

	for _, item := range data {
		sourceChan <- item
	}
	close(sourceChan) 

	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)

	errChan := sink.Open(context.Background())

	var receivedData []string
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

	err := <-errChan 
	if err != nil {
		t.Fatalf("Sink returned an error: %v", err)
	}

	wg.Wait() 

	if len(receivedData) != len(data) {
		t.Errorf("Expected %d items, got %d. Received: %v", len(data), len(receivedData), receivedData)
	}
	for i, expected := range data {
		if receivedData[i] != expected {
			t.Errorf("Expected item %s at index %d, got %s", expected, i, receivedData[i])
		}
	}
	close(destChan) 
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
	close(destChan)
}

func TestChanSink_ContextCancellation(t *testing.T) {
	sourceChan := make(chan any)    
	destChan := make(chan int)       

	sink := Channel(destChan)
	sink.SetLogFunc(testSinkLogFunc(t))
	sink.SetInput(sourceChan)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := sink.Open(ctx)

	go func() {
		select {
		case sourceChan <- 123:
		case <-time.After(1 * time.Second):
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
			t.Log("destChan closed or no item sent as expected for type mismatch test")
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

	finalErr := <-errChan 

	if finalErr == nil {
		t.Fatalf("Expected an error from sink due to panic (send on closed channel), but got nil (errChan closed without error)")
	}

	expectedPanicSubstring := "panic in ChanSink worker" 
	if !strings.Contains(finalErr.Error(), expectedPanicSubstring) {
		t.Errorf("Error received does not indicate a panic. Got: '%v', Expected substring: '%s'", finalErr, expectedPanicSubstring)
	} else {
		t.Logf("Test: Correctly received panic-related error from sink: %v", finalErr)
	}

	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		defer func() {
			_ = recover() 
		}()
		
		time.Sleep(100 * time.Millisecond) 
		doneSending := false
		for i := 5; i < 8 && !doneSending; i++ { 
			select {
			case sourceChan <- fmt.Sprintf("late-item-%d", i):
			case <-time.After(50 * time.Millisecond):
				doneSending = true 
			}
		}
		
		select {
		case _, stillOpen := <-sourceChan:
			if stillOpen { 
				t.Log("Test Cleanup Goroutine: sourceChan was still open or had items during cleanup attempt.")
				close(sourceChan)
			} else {
				t.Log("Test Cleanup Goroutine: sourceChan already closed during cleanup.")
			}
		default:
			t.Log("Test Cleanup Goroutine: sourceChan empty or already closed; attempting close.")
			close(sourceChan)
		}
	}()

	var receivedItemsAfterClose []string
	for remainingItem := range destChan { 
		receivedItemsAfterClose = append(receivedItemsAfterClose, remainingItem)
	}
	t.Logf("Test: Received %d items from destChan after it was closed by test's goroutine: %v", len(receivedItemsAfterClose), receivedItemsAfterClose)
	
	select {
	case <-cleanupDone:
	case <-time.After(1 * time.Second): 
		t.Log("Test: Cleanup goroutine for sourceChan timed out")
	}
}

