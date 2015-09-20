package proc

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
)

// ItemProc implements a processor that applies a function to each item received.
// Processed items are placed in the output channel for down stream use.
type ItemProc struct {
	Name        string                        // Name identifier for component
	Function    func(interface{}) interface{} // function to execute
	Concurrency int                           // Concurrency level, default 1

	input  <-chan interface{}
	output chan interface{}
	done   chan struct{}
	log    *logrus.Entry
}

func (p *ItemProc) Init(ctx context.Context) error {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "ItemProc")
		log.Error("Logger not found in context")
	}
	p.log = log
	p.log.Info("Initializing component", p.Name)

	if p.Name == "" {
		return api.ProcError{Err: fmt.Errorf("Name attribute is required")}
	}
	if p.input == nil {
		return api.ProcError{
			ProcName: p.Name,
			Err:      fmt.Errorf("Input attribute required"),
		}
	}

	if p.Function == nil {
		return api.ProcError{
			ProcName: p.Name,
			Err:      fmt.Errorf("Function attribute must be provided"),
		}
	}

	if p.Concurrency == 0 {
		p.Concurrency = 1
	}

	p.output = make(chan interface{})
	p.done = make(chan struct{})

	return nil
}

func (p *ItemProc) Uninit(ctx context.Context) error {
	return nil
}

func (p *ItemProc) GetName() string {
	return p.Name
}

func (p *ItemProc) SetInput(in <-chan interface{}) {
	p.input = in
}

func (p *ItemProc) GetOutput() <-chan interface{} {
	return p.output
}

func (p *ItemProc) Done() <-chan struct{} {
	return p.done
}

func (p *ItemProc) Exec(ctx context.Context) (err error) {
	go func() {
		defer func() {
			close(p.output)
			close(p.done)
		}()

		var barrier sync.WaitGroup
		barrier.Add(p.Concurrency)

		for i := 0; i < p.Concurrency; i++ {
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				p.doProc(p.input)
			}(&barrier)
		}

		barrier.Wait()
	}()
	return
}

func (p *ItemProc) doProc(input <-chan interface{}) {
	for item := range input {
		procd := p.Function(item)
		switch val := procd.(type) {
		case nil:
			continue
		case api.ProcError:
			p.log.Error(val)
		default:
			p.output <- val
		}
	}
}
