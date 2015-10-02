package proc

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
)

// Endpoint implements an endpoint processor that applies a function to each item
// received from its input. Once all items are processed, the endpoin closes Done().
// This is a terminal component.  If you need to pass output downstream, use  the
// Item processor.
type Endpoint struct {
	Name        string                  // Name identifier for component
	Function    func(interface{}) error // function to execute
	Concurrency int                     // Concurrency level, default 1

	input <-chan interface{}
	done  chan struct{}
	log   *logrus.Entry
}

func (p *Endpoint) Init(ctx context.Context) error {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "Endpoint")
		log.Error("Logger not found in context")
	}

	p.log = log.WithFields(logrus.Fields{
		"Component": p.Name,
		"Type":      fmt.Sprintf("%T", p),
	})

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

	p.done = make(chan struct{})
	p.log.Info("Component initialized")
	return nil
}

func (p *Endpoint) Uninit(ctx context.Context) error {
	return nil
}

func (p *Endpoint) GetName() string {
	return p.Name
}

func (p *Endpoint) SetInput(in <-chan interface{}) {
	p.input = in
}

func (p *Endpoint) Done() <-chan struct{} {
	return p.done
}

func (p *Endpoint) Exec(ctx context.Context) (err error) {
	p.log.Info("Execution started")
	go func() {
		defer func() {
			close(p.done)
			p.log.Info("Execution completed")
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

func (p *Endpoint) doProc(input <-chan interface{}) {
	for item := range input {
		err := p.Function(item)
		if err != nil {
			p.log.Error(err)
		}
	}
}
