package sinks

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
)

func TestChanSink_Open(t *testing.T) {
	cs := Chan[string]()

	inputChan := make(chan any)
	go func() {
		inputChan <- "A"
		inputChan <- "B"
		inputChan <- "C"
		close(inputChan)
	}()
	cs.SetInput(inputChan)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	sink := cs.Open(ctx)

	var receivedData []string
	doneReceiving := make(chan struct{})

	go func() {
		defer close(doneReceiving)
		for data := range cs.Get() {
			receivedData = append(receivedData, data)
		}
	}()

	select {
	case err := <-sink:
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
}

func TestChanSink_Open_ContextCancel(t *testing.T) {
	cs := Chan[int]()
	inputChan := make(chan any)

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
	cs := Chan[string]() // Expecting string
	inputChan := make(chan any)

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

func TestChanSink_Open_InputChannelNotSet(t *testing.T) {
	// Create a ChanSink without setting input channel
	cs := Chan[string]()

	// Call Open without setting input - should return ErrInputChannelUndefined
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	sinkErrCh := cs.Open(ctx)

	select {
	case err := <-sinkErrCh:
		if err == nil {
			t.Fatal("Expected error when input channel is not set, but got nil")
		}
		if err != api.ErrInputChannelUndefined {
			t.Fatalf("Expected ErrInputChannelUndefined, but got: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Context timed out waiting for error")
	}
}
