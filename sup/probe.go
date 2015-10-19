package sup

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	autoctx "github.com/vladimirvivien/automi/context"
)

// ProbeFunc type for implementing function to be executed for each item probed.
type ProbeFunc func(context.Context, interface{}) interface{}

// The Probe is a processor designed for testing and inspecting data flow.
// It captures data in its input channel, apply specified function,
// then outputs the result to its outupt channel
type Probe struct {
	Name    string
	Examine ProbeFunc

	input  <-chan interface{}
	output chan interface{}
	log    *logrus.Entry
}

func (p *Probe) Init(ctx context.Context) error {
	// validation
	if p.Name == "" {
		return fmt.Errorf("Missing name identifier")
	}

	if p.input == nil {
		return fmt.Errorf("Probe [%s] input not set", p.Name)
	}

	p.output = make(chan interface{})

	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("ProcName", p.Name)
		log.Errorf("No valid logger set for %s", p.Name)
	}

	p.log = log.WithFields(logrus.Fields{
		"Component": p.Name,
		"Type":      fmt.Sprintf("%T", p),
	})
	p.log.Info("Component initialized")

	return nil
}

func (p *Probe) Uninit(ctx context.Context) error {
	return nil
}

func (p *Probe) GetName() string {
	return p.Name
}

func (p *Probe) SetInput(in <-chan interface{}) {
	p.input = in
}

func (p *Probe) GetOutput() <-chan interface{} {
	return p.output
}

func (p *Probe) Exec(ctx context.Context) error {
	exeCtx, cancel := context.WithCancel(ctx)
	defer cancel() // cancel everthing downstream (if necessary)

	p.log.Info("Execution started")
	go func() {
		defer func() {
			close(p.output)
			p.log.Info("Execution completed")
		}()

		// output data
		for item := range p.input {
			if p.Examine != nil {
				p.output <- p.Examine(exeCtx, item)
			}

			select {
			case <-ctx.Done():
				cancel()
				return
			default:
			}
		}
	}()
	return nil
}
