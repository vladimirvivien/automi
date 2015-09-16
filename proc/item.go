package proc

import (
	"fmt"
	"sync"

	"github.com/vladimirvivien/automi/api"
)

// ItemProc implements a processor that applies a function to each item received.
// Processed items are placed in the output channel for down stream use.
type ItemProc struct {
	Name        string                        // Name identifier for component
	Input       <-chan interface{}            // input channel for component
	Function    func(interface{}) interface{} // function to execute
	Concurrency int                           // Concurrency level, default 1

	output  chan interface{}
	done    chan struct{}
	errChan chan api.ProcError
}

func (p *ItemProc) Init() error {
	if p.Name == "" {
		return api.ProcError{Err: fmt.Errorf("Name attribute is required")}
	}
	if p.Input == nil {
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
	p.errChan = make(chan api.ProcError)

	return nil
}

func (p *ItemProc) Uninit() error {
	return nil
}

func (p *ItemProc) GetName() string {
	return p.Name
}

func (p *ItemProc) GetInput() <-chan interface{} {
	return p.Input
}

func (p *ItemProc) GetOutput() <-chan interface{} {
	return p.output
}

func (p *ItemProc) GetErrors() <-chan api.ProcError {
	return p.errChan
}

func (p *ItemProc) Done() <-chan struct{} {
	return p.done
}

func (p *ItemProc) Exec() (err error) {
	go func() {
		defer func() {
			close(p.output)
			close(p.done)
			close(p.errChan)
		}()

		var barrier sync.WaitGroup
		barrier.Add(p.Concurrency)

		for i := 0; i < p.Concurrency; i++ {
			go func() {
				defer barrier.Done()
				p.doProc(p.Input)
			}()
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
			p.errChan <- val
		default:
			p.output <- val
		}
	}
}
