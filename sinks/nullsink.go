package sinks

import (
	"context"
	"log"
)

type NullSink struct {
	input <-chan interface{}
	log   *log.Logger
}

func Null() *NullSink {
	return new(NullSink)
}

func (s *NullSink) SetInput(in <-chan interface{}) {
	s.input = in
}

// Open opens the sink for receiving data
func (s *NullSink) Open(ctx context.Context) <-chan error {
	result := make(chan error)
	go func() {
		defer close(result)
		for _ = range s.input {
		}
	}()
	return result
}
