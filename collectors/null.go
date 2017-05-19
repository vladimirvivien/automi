package collectors

import (
	"context"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

type NullCollector struct {
	input <-chan interface{}
	log   *log.Logger
}

func Null() *NullCollector {
	return new(NullCollector)
}

func (s *NullCollector) SetInput(in <-chan interface{}) {
	s.input = in
}

// Open opens the node to start collecting
func (s *NullCollector) Open(ctx context.Context) <-chan error {
	result := make(chan error)
	s.log = autoctx.GetLogger(ctx)

	s.log.Print("opening null collector")

	go func() {
		defer func() {
			close(result)
			s.log.Print("closing null collector")
		}()
		for _ = range s.input {
		}
	}()
	return result
}
