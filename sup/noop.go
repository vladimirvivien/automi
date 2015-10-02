package sup

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
	"golang.org/x/net/context"
)

// NoopProc represents a non-operational processor.
// It simply return its input as the output channel.
// A Noop processor can be useful in testing and
// setting up an Automi process graph.
type NoopProc struct {
	Name string

	input <-chan interface{}
	log   *logrus.Entry
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
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "Endpoint")
		log.Error("Logger not found in context")
	}

	n.log = log.WithFields(logrus.Fields{
		"Component": n.Name,
		"Type":      fmt.Sprintf("%T", n),
	})

	if n.input == nil {
		return api.ProcError{Err: fmt.Errorf("Input attribute not set")}
	}

	n.log.Info("Component initialized")

	return nil
}

func (n *NoopProc) Exec(ctx context.Context) error {
	return nil
}

func (n *NoopProc) Uninit(ctx context.Context) error {
	return nil
}
