package sup

import (
	"fmt"

	"github.com/vladimirvivien/automi/api"
	"golang.org/x/net/context"
)

// NoopProc represents a non-operational processor.
// It simply return its input as the output channel.
// A Noop processor can be useful in testing and
// setting up an Automi process graph.
type NoopProc struct {
	Name string

	input <-chan interface{}
}

func (n *NoopProc) GetName() string {
	return n.Name
}

func (n *NoopProc) SetInput(in <-chan interface{}) {
	n.input = in
}

func (n *NoopProc) GetOutput() <-chan interface{} {
	return n.input
}

func (n *NoopProc) Init(ctx context.Context) error {
	if n.input == nil {
		return api.ProcError{Err: fmt.Errorf("Input attribute not set")}
	}
	return nil
}

func (n *NoopProc) Exec(ctx context.Context) error {
	return nil
}

func (n *NoopProc) Uninit(ctx context.Context) error {
	return nil
}
