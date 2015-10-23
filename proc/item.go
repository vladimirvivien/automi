package proc

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
)

// Item implements a processor that applies a function to each item received.
// Processed items are expected to be placed on the output channel for down stream use.
// Use and endpoint processor for termination.
type Item struct {
	Name        string                                         // Name identifier for component
	Function    func(context.Context, interface{}) interface{} // function to execute
	Concurrency int                                            // Concurrency level, default 1

	input     <-chan interface{}
	output    chan interface{}
	log       *logrus.Entry
	cancelled bool
	mutex     sync.RWMutex
}

func (p *Item) Init(ctx context.Context) error {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "Item")
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

	p.output = make(chan interface{}, 1024)

	p.log.Infof("Component [%s] initialized", p.Name)
	return nil
}

func (p *Item) Uninit(ctx context.Context) error {
	return nil
}

func (p *Item) GetName() string {
	return p.Name
}

func (p *Item) SetInput(in <-chan interface{}) {
	p.input = in
}

func (p *Item) GetOutput() <-chan interface{} {
	return p.output
}

func (p *Item) Exec(ctx context.Context) (err error) {
	p.log.Info("Execution started for ", p.Name)

	exeCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer func() {
			close(p.output)
			p.log.Info("Shuttingdown component ", p.Name)
		}()

		var barrier sync.WaitGroup
		barrier.Add(p.Concurrency)

		for i := 0; i < p.Concurrency; i++ {
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				p.doProc(exeCtx, p.input)
			}(&barrier)
		}

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			barrier.Wait()
		}()

		select {
		case <-wait:
		case <-ctx.Done():
			p.log.Infof("Component [%s] cancelling...", p.Name)
			cancel()
			p.mutex.Lock()
			p.cancelled = true
			p.mutex.Unlock()
			return
		}
	}()
	return
}

func (p *Item) doProc(exeCtx context.Context, input <-chan interface{}) {
	for item := range input {
		procd := p.Function(exeCtx, item)
		switch val := procd.(type) {
		case nil:
			continue
		case api.ProcError:
			p.log.Error(val)
			continue
		}

		p.mutex.Lock()
		if !p.cancelled {
			p.output <- item
		} else {
			return
		}
		p.mutex.Unlock()
	}
}
