package stream

import (
	"context"
	"log"
	"os"
)

type Drain struct {
	output chan interface{}
	input  <-chan interface{}
	log    *log.Logger
}

func NewDrain() *Drain {
	return &Drain{
		output: make(chan interface{}, 1024),
		log:    log.New(os.Stderr, "", log.Flags()),
	}
}

func (s *Drain) SetInput(in <-chan interface{}) {
	s.input = in
}

func (s *Drain) GetOutput() <-chan interface{} {
	return s.output
}

func (s *Drain) Open(ctx context.Context) <-chan error {
	s.log.Print("cpening component")
	result := make(chan error)
	go func() {
		defer func() {
			s.log.Print("closing component")
			close(s.output)
			close(result)
		}()
		for data := range s.input {
			s.output <- data
		}
	}()
	return result
}
