package stream

import (
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type Drain struct {
	output chan interface{}
	input  <-chan interface{}
	log    *logrus.Entry
}

func NewDrain() *Drain {
	return &Drain{
		output: make(chan interface{}, 1024),
		log:    logrus.WithField("Component", "Drain"),
	}
}

func (s *Drain) SetInput(in <-chan interface{}) {
	s.input = in
}

func (s *Drain) GetOutput() <-chan interface{} {
	return s.output
}

func (s *Drain) Open(ctx context.Context) <-chan error {
	s.log.Infoln("Opening component")
	result := make(chan error)
	go func() {
		defer func() {
			close(s.output)
			close(result)
		}()
		for data := range s.input {
			s.output <- data
		}
	}()
	return result
}
