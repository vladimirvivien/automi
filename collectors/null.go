package collectors

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

type NullCollector struct {
	input <-chan interface{}
	logf  api.LogFunc
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
	s.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(s.logf, "Opening null collector")

	go func() {
		defer func() {
			util.Logfn(s.logf, "Closing null collector")
			close(result)
		}()
		// TODO check for ctx cancellation
		for range s.input {
		}
	}()
	return result
}
