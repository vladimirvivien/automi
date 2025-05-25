package sinks

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestChanSink_Open(t *testing.T) {
	outputChan := make(chan string)
	cs := Chan[string](outputChan)

	inputChan := make(chan interface{})
	go func() {
		inputChan <- "A"
		inputChan <- "B"
		inputChan <- "C"
		close(inputChan)
	}()
	cs.SetInput(inputChan)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	sinkErr := cs.Open(ctx)

	var receivedData []string
	doneReceiving := make(chan struct{})

	go func() {
		defer close(doneReceiving)
		for data := range cs.Get() {
			receivedData = append(receivedData, data)
		}
	}()

	select {
	case err := <-sinkErr:
		if err != nil {
			t.Fatalf("ChanSink.Open() returned an error: %v", err)
		}
	case <-ctx.Done():
		// This case should ideally not be hit if the sink closes properly.
		// If it is, it might indicate the sink didn't close when the input channel closed.
		t.Log("Context timed out, checking received data.")
	}

	// Wait for the receiving goroutine to finish.
	// This is important to ensure all data is collected before assertions.
	<-doneReceiving

	expectedData := []string{"A", "B", "C"}
	if !reflect.DeepEqual(expectedData, receivedData) {
		t.Errorf("Received data did not match expected data. Got %v, expected %v", receivedData, expectedData)
	}

	// Check if the sink's output channel was closed by the sink (indirectly via the receiving goroutine)
	// This is a bit tricky to test directly for a send-only channel from the sink's perspective.
	// However, if the receiving goroutine exited, it implies the channel was closed.
	// A more direct test might involve trying to send to cs.Get() after it's supposed to be closed,
	// but that's not how this sink is designed to be used.
}

func TestChanSink_Open_ContextCancel(t *testing.T) {
	outputChan := make(chan int)
	cs := Chan[int](outputChan)
	inputChan := make(chan interface{})

	cs.SetInput(inputChan)

	ctx, cancel := context.WithCancel(context.Background())

	sinkErrCh := cs.Open(ctx)

	// Cancel the context after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Consume from the output channel to prevent blockage
	go func() {
		for range cs.Get() {
			// Discard data
		}
	}()

	select {
	case err := <-sinkErrCh:
		if err != nil {
			t.Fatalf("ChanSink.Open() returned an unexpected error: %v", err)
		}
		// If err is nil, it means the sink closed gracefully, which is expected after context cancellation.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ChanSink did not close after context cancellation")
	}
}

func TestChanSink_Open_WrongDataType(t *testing.T) {
	outputChan := make(chan string)
	cs := Chan[string](outputChan) // Expecting string
	inputChan := make(chan interface{})

	go func() {
		inputChan <- "CorrectType"
		inputChan <- 123 // Incorrect type
		inputChan <- "AnotherCorrectType"
		close(inputChan)
	}()
	cs.SetInput(inputChan)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sinkErr := cs.Open(ctx)

	var receivedData []string
	doneReceiving := make(chan struct{})
	go func() {
		defer close(doneReceiving)
		for data := range cs.Get() {
			receivedData = append(receivedData, data)
		}
	}()

	select {
	case err := <-sinkErr:
		if err != nil {
			t.Fatalf("ChanSink.Open() returned an error: %v", err)
		}
	case <-ctx.Done():
		t.Log("Context timed out, checking received data.")
	}

	<-doneReceiving

	expectedData := []string{"CorrectType", "AnotherCorrectType"}
	if !reflect.DeepEqual(expectedData, receivedData) {
		t.Errorf("Received data did not match expected data. Got %v, expected %v", receivedData, expectedData)
	}
}
