package collectors

import (
	"context"

	"github.com/go-faces/logger"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

type NullCollector struct {
	input <-chan interface{}
	log   logger.Interface
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
	util.Log(s.log, "opening null collector")

	go func() {
		defer func() {
			close(result)
			util.Log(s.log, "closing null collector")
		}()
		for range s.input {
		}
	}()
	return result
}
