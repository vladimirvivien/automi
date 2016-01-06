package stream

import (
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type SliceSrc struct {
	src    []interface{}
	output chan interface{}
	log    *logrus.Entry
}

func NewSliceSource(elems ...interface{}) *SliceSrc {
	return &SliceSrc{
		src:    elems,
		output: make(chan interface{}, 1024),
		log:    logrus.WithField("Component", "SliceSource"),
	}
}
func (s *SliceSrc) GetOutput() <-chan interface{} {
	return s.output
}
func (s *SliceSrc) Open(ctx context.Context) error {
	s.log.Infoln("Opening source")
	go func() {
		defer close(s.output)
		for _, str := range s.src {
			s.output <- str
		}
	}()
	return nil
}
