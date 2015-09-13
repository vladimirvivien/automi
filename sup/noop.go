package sup

import "github.com/vladimirvivien/automi/api"

// NoopProc represents a non-operational processor.
// A Noop processor can be useful in testing and
// setting up an Automi process graph.
type NoopProc struct {
	Name   string
	Input  <-chan interface{}
	Output <-chan interface{}
}

func (n *NoopProc) GetName() string {
	return n.Name
}

func (n *NoopProc) GetInput() <-chan interface{} {
	return n.Input
}

func (n *NoopProc) GetOutput() <-chan interface{} {
	return n.Output
}

func (n *NoopProc) GetError() <-chan api.ProcError {
	return nil
}

func (n *NoopProc) Init() error {
	return nil
}

func (n *NoopProc) Exec() error {
	return nil
}

func (n *NoopProc) Uninit() error {
	return nil
}
